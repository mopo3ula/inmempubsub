package inmemory

import (
	"context"
	"testing"

	"github.com/mopo3ula/inmempubsub/internal"
	"github.com/stretchr/testify/assert"
)

var ctx = context.Background()
var logger = StdDebugLogger{}

func TestPubSub_AddSubscriber(t *testing.T) {
	t.Parallel()

	ps := NewPubSub(logger)

	sub1 := newMockSubscriber()
	sub2 := newMockSubscriber()

	ps.AddSubscribers(ctx, sub1, sub2)

	assert.Len(t, ps.subscribers.m, 2)

	ps.Stop()

	assert.Len(t, ps.subscribers.m, 0)
}

func TestPubSub_DeleteSubscriber(t *testing.T) {
	t.Parallel()

	ps := NewPubSub(logger)

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

	ps := NewPubSub(logger)

	assert.NotPanics(t, func() { ps.Stop() })
}

func BenchmarkPubSub_DeleteSubscriber(b *testing.B) {
	ps := NewPubSub(logger)

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

func newMockSubscriber() mockSubscriber {
	return mockSubscriber{
		topic: internal.RandString(15),
		data:  make(chan any),
	}
}

type mockSubscriber struct {
	topic string
	data  chan any
}

func (m mockSubscriber) Handle() func(_ context.Context, _ any) error {
	return nil
}

func (m mockSubscriber) Topic() string {
	return m.topic
}

func (m mockSubscriber) Data() chan any {
	return m.data
}
