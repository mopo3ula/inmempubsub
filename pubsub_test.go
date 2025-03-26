package inmemory

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/mopo3ula/inmempubsub/internal"
	"github.com/mopo3ula/inmempubsub/internal/logger"
	"github.com/stretchr/testify/assert"
)

var ctx = context.Background()
var debugLogger = StdDebugLogger{}
var emptyLogger = EmptyLogger{}

func TestPubSub_AddSubscriber(t *testing.T) {
	t.Parallel()

	t.Run("add subscriber success", func(t *testing.T) {
		t.Parallel()

		ps := NewPubSub(debugLogger)

		sub1 := newMockSubscriber()
		sub2 := newMockSubscriber()

		ps.AddSubscribers(ctx, []Subscriber{sub1, sub2})

		assert.Len(t, ps.subscribers.m, 2)

		ps.Stop()

		assert.Len(t, ps.subscribers.m, 0)
	})

	t.Run("fail only one subscriber", func(t *testing.T) {
		t.Parallel()

		ps := NewPubSub(debugLogger)

		sub1 := newMockSubscriber(withTopic("same_topic_name"))
		sub2 := newMockSubscriber(withTopic("same_topic_name"))

		ps.AddSubscribers(ctx, []Subscriber{sub1, sub2})

		assert.Len(t, ps.subscribers.m, 1)
	})
}

func TestPubSub_DeleteSubscriber(t *testing.T) {
	t.Parallel()

	ps := NewPubSub(debugLogger)

	sub := newMockSubscriber()
	ps.AddSubscriber(ctx, sub)

	_, ok := ps.subscribers.Load(sub.Topic())
	assert.True(t, ok)

	ps.stopTopic(sub.Topic())

	_, ok = ps.subscribers.Load(sub.Topic())
	assert.False(t, ok)
	assert.NotPanics(t, func() { ps.stopTopic("dont_exist") })

	ps.Stop()
}

func TestPubSub_Stop(t *testing.T) {
	t.Parallel()

	ps := NewPubSub(debugLogger)

	assert.NotPanics(t, func() { ps.Stop() })
}

func BenchmarkPubSub_DeleteSubscriber(b *testing.B) {
	ps := NewPubSub(emptyLogger)

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		sub := newMockSubscriber()
		ps.AddSubscriber(ctx, sub)

		_, ok := ps.subscribers.Load(sub.Topic())
		assert.True(b, ok)

		ps.stopTopic(sub.Topic())

		_, ok = ps.subscribers.Load(sub.Topic())
		assert.False(b, ok)
		assert.NotPanics(b, func() { ps.stopTopic("dont_exist") })

		ps.Stop()
	}
}

func TestPubSub_Send(t *testing.T) {
	t.Parallel()

	t.Run("multiple send", func(t *testing.T) {
		t.Parallel()

		topic := NewTopic("common_topic_name")

		subs := []*mockSubscriber{
			newMockSubscriber(withTopic(topic)),
			newMockSubscriber(withTopic(topic)),
			newMockSubscriber(withTopic(topic)),
		}
		ps := NewPubSub(debugLogger)
		for _, sub := range subs {
			ps.AddSubscriber(ctx, sub)
		}

		defer ps.Stop()

		data := internal.RandString(10)
		ps.Send(topic, data)

		var wg sync.WaitGroup
		wg.Add(len(subs))

		var handleTimes atomic.Int32
		for _, sub := range subs {
			go func(s *mockSubscriber) {
				defer wg.Done()

				<-s.received
				close(s.received)

				handleTimes.Add(1)
			}(sub)
		}

		wg.Wait()
		assert.Equal(t, int32(len(subs)), handleTimes.Load())
	})

	t.Run("multiple send with semaphore", func(t *testing.T) {
		t.Parallel()

		start := time.Now()
		topic := NewTopic("sem_topic_name")

		sub := newMockSubscriber(
			withTopic(topic),
			withDelay(1*time.Second),
		)
		ps := NewPubSub(debugLogger)
		ps.AddSubscriber(ctx, sub, WithMaxConcurrency(2))

		defer ps.Stop()

		n := 6
		var wg sync.WaitGroup
		wg.Add(n)
		for i := 0; i < n; i++ {
			go func() {
				ps.Send(topic, internal.RandString(10))
			}()
		}

		go func(s *mockSubscriber) {
			for range s.received {
				wg.Done()
			}
		}(sub)

		wg.Wait()
		close(sub.received)

		testDur := time.Since(start)
		delta := n / 2

		assert.GreaterOrEqual(t, testDur.Seconds(), (time.Duration(delta) * time.Second).Seconds())
		assert.LessOrEqual(t, testDur.Seconds(), (time.Duration(delta+1) * time.Second).Seconds())
	})
}

func newMockSubscriber(opts ...subOption) *mockSubscriber {
	sub := &mockSubscriber{
		topic:    Topic(internal.RandString(10)),
		data:     make(chan any),
		logger:   debugLogger,
		received: make(chan string),
	}

	for _, opt := range opts {
		opt(sub)
	}

	return sub
}

type mockSubscriber struct {
	topic    Topic
	data     chan any
	logger   logger.Logger
	received chan string
	delay    *time.Duration
}

func (m *mockSubscriber) Handle() func(_ context.Context, data any) error {
	return func(_ context.Context, data any) error {
		if d, ok := data.(string); ok {
			m.logger.Debugf("received data '%s' for topic '%s'", d, m.Topic())

			if m.delay != nil {
				time.Sleep(*m.delay)
			}

			m.received <- d
		}

		return nil
	}
}

func (m *mockSubscriber) Topic() Topic {
	return m.topic
}

func (m *mockSubscriber) Data() chan any {
	return m.data
}

type subOption func(subscriber *mockSubscriber)

func withTopic(topic Topic) subOption {
	return func(ms *mockSubscriber) {
		ms.topic = topic
	}
}

func withDelay(d time.Duration) subOption {
	return func(ms *mockSubscriber) {
		ms.delay = &d
	}
}
