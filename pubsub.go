package inmemory

import (
	"context"
	"sync"

	"github.com/mopo3ula/inmempubsub/internal/logger"
)

type Publisher interface {
	Send(topic Topic, data any)
}

type Subscriber interface {
	Handle() func(ctx context.Context, data any) error
	Topic() Topic
	Data() chan any
}

type handler struct {
	data   chan any
	handle func(ctx context.Context, data any) error
	done   chan struct{}
	wg     sync.WaitGroup
	sem    *semaphore

	maxConcurrency int // default is ConcurrencyUnlimited
}

type PubSub struct {
	subscribers *subscribers
	wg          sync.WaitGroup
	logger      logger.Logger
}

func NewPubSub(logger logger.Logger) *PubSub {
	return &PubSub{
		subscribers: &subscribers{m: map[Topic][]*handler{}},
		logger:      logger,
	}
}

func (ps *PubSub) addSubscriber(ctx context.Context, s Subscriber, opts ...Option) {
	h := &handler{
		data:           s.Data(),
		handle:         s.Handle(),
		done:           make(chan struct{}),
		wg:             sync.WaitGroup{},
		maxConcurrency: ConcurrencyUnlimited,
	}

	for _, opt := range opts {
		opt(h)
	}

	if h.maxConcurrency > ConcurrencyUnlimited {
		h.sem = &semaphore{c: make(chan struct{}, h.maxConcurrency)}
	}

	ps.subscribers.Add(s.Topic(), h)

	ps.wg.Add(1)
	go ps.runSubscriber(ctx, h)
}

func (ps *PubSub) AddSubscribers(ctx context.Context, subscribers []Subscriber, opts ...Option) {
	for _, s := range subscribers {
		ps.addSubscriber(ctx, s, opts...)
	}
}

func (ps *PubSub) AddSubscriber(ctx context.Context, subscriber Subscriber, opts ...Option) {
	ps.addSubscriber(ctx, subscriber, opts...)
}

func (ps *PubSub) stopTopic(topic Topic) {
	handlers, ok := ps.subscribers.Load(topic)
	if !ok {
		ps.logger.Errorf("no subscriber with name '%s'", topic)
		return
	}

	ps.subscribers.Del(topic)

	for _, h := range handlers {
		close(h.done) // finish signal for handler
		h.wg.Wait()   // wait for all subscribers

		// close data channel
		close(h.data)
	}
}

func (ps *PubSub) Stop() {
	stopWg := new(sync.WaitGroup)

	ps.subscribers.Lock()
	for topic := range ps.subscribers.m {
		stopWg.Add(1)
		go func(w *sync.WaitGroup, t Topic) {
			defer w.Done()

			ps.stopTopic(t)
		}(stopWg, topic)
	}
	ps.subscribers.Unlock()

	stopWg.Wait()
	ps.wg.Wait()
}

func (ps *PubSub) Send(topic Topic, data any) {
	if handlers, ok := ps.subscribers.Load(topic); ok {
		for _, h := range handlers {
			h.data <- data
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

			h.sem.acquire()

			h.wg.Add(1)
			go func(d any) {
				defer h.wg.Done()
				defer h.sem.release()

				if h.handle != nil {
					if err := h.handle(ctx, d); err != nil {
						ps.logger.Errorf("handle error: %v", err)
					}
				}
			}(data)
		}
	}
}
