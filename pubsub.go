package inmemory

import (
	"context"
	"sync"

	"github.com/mopo3ula/inmempubsub/internal/logger"
)

type Publisher interface {
	Send(topic string, data any)
}

type Subscriber interface {
	Handle() func(ctx context.Context, data any) error
	Topic() string
	Data() chan any
}

type handler struct {
	data   chan any
	handle func(ctx context.Context, data any) error
	done   chan struct{}
	wg     sync.WaitGroup
}

type PubSub struct {
	subscribers *subscribers
	wg          sync.WaitGroup
	logger      logger.Logger
}

func NewPubSub(logger logger.Logger) *PubSub {
	return &PubSub{
		subscribers: &subscribers{m: map[string]*handler{}},
		logger:      logger,
	}
}

func (ps *PubSub) addSubscriber(ctx context.Context, s Subscriber) {
	_, ok := ps.subscribers.Load(s.Topic())
	if ok {
		ps.logger.Errorf("subscriber '%s' already exists", s.Topic())
		return
	}

	h := &handler{
		data:   s.Data(),
		handle: s.Handle(),
		done:   make(chan struct{}),
		wg:     sync.WaitGroup{},
	}

	ps.subscribers.Store(s.Topic(), h)

	ps.wg.Add(1)
	go ps.runSubscriber(ctx, h)
}

func (ps *PubSub) AddSubscribers(ctx context.Context, subscribers ...Subscriber) {
	for _, s := range subscribers {
		ps.addSubscriber(ctx, s)
	}
}

func (ps *PubSub) stopTopic(topic string) {
	val, ok := ps.subscribers.Load(topic)
	if !ok {
		ps.logger.Errorf("no subscriber with name '%s'", topic)
		return
	}

	ps.subscribers.Del(topic)

	close(val.done) // finish signal for handler
	val.wg.Wait()   // wait for all subscribers

	// close data channel
	close(val.data)
}

func (ps *PubSub) Stop() {
	stopWg := new(sync.WaitGroup)

	ps.subscribers.Lock()
	for topic := range ps.subscribers.m {
		stopWg.Add(1)
		go func(w *sync.WaitGroup, t string) {
			defer w.Done()

			ps.stopTopic(t)
		}(stopWg, topic)
	}
	ps.subscribers.Unlock()

	stopWg.Wait()
	ps.wg.Wait()
}

func (ps *PubSub) Send(topic string, data any) {
	if h, ok := ps.subscribers.Load(topic); ok {
		select {
		case h.data <- data:
		default:
		}
	}
}

func (ps *PubSub) runSubscriber(ctx context.Context, h *handler) {
	defer ps.wg.Done()

	for {
		select {
		case <-ctx.Done():
			return
		case <-h.done:
			return
		case data, ok := <-h.data:
			if !ok {
				return
			}
			h.wg.Add(1)
			go func(d any) {
				defer h.wg.Done()
				if h.handle != nil {
					if err := h.handle(ctx, d); err != nil {
						ps.logger.Errorf("handle error: %v", err)
					}
				}
			}(data)
		}
	}
}
