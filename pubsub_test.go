package inmemory

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"

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

		ps.AddSubscribers(ctx, sub1, sub2)

		assert.Len(t, ps.subscribers.m, 2)

		ps.Stop()

		assert.Len(t, ps.subscribers.m, 0)
	})

	t.Run("fail only one subscriber", func(t *testing.T) {
		t.Parallel()

		ps := NewPubSub(debugLogger)

		sub1 := newMockSubscriber(withTopicName("same_topic_name"))
		sub2 := newMockSubscriber(withTopicName("same_topic_name"))

		ps.AddSubscribers(ctx, sub1, sub2)

		assert.Len(t, ps.subscribers.m, 1)
	})
}

func TestPubSub_DeleteSubscriber(t *testing.T) {
	t.Parallel()

	ps := NewPubSub(debugLogger)

	sub := newMockSubscriber()
	ps.AddSubscribers(ctx, sub)

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
		ps.AddSubscribers(ctx, sub)

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
	key := "common_topic_name"

	subs := []*mockSubscriber{
		newMockSubscriber(withTopicName(key)),
		newMockSubscriber(withTopicName(key)),
		newMockSubscriber(withTopicName(key)),
	}
	ps := NewPubSub(debugLogger)
	for _, sub := range subs {
		ps.AddSubscribers(ctx, sub)
	}

	defer ps.Stop()

	data := internal.RandString(10)
	ps.Send(key, data)

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
}

func newMockSubscriber(opts ...subOption) *mockSubscriber {
	sub := &mockSubscriber{
		topic:    internal.RandString(10),
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
	topic    string
	data     chan any
	logger   logger.Logger
	received chan string
}

func (m *mockSubscriber) Handle() func(_ context.Context, data any) error {
	return func(_ context.Context, data any) error {
		if d, ok := data.(string); ok {
			m.logger.Debugf("received data '%s' for topic '%s'", d, m.Topic())
			m.received <- d
		}

		return nil
	}
}

func (m *mockSubscriber) Topic() string {
	return m.topic
}

func (m *mockSubscriber) Data() chan any {
	return m.data
}

type subOption func(subscriber *mockSubscriber)

func withTopicName(topicName string) subOption {
	return func(ms *mockSubscriber) {
		ms.topic = topicName
	}
}
