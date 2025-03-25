package inmemory

import (
	"fmt"
	"sync"
	"testing"

	"github.com/mopo3ula/inmempubsub/internal"
	"github.com/stretchr/testify/assert"
)

func TestSubscribers_Load(t *testing.T) {
	t.Parallel()

	s := subscribers{
		m: make(map[Topic][]*handler),
	}

	key := NewTopic(internal.RandString(10))
	value := &handler{}
	s.Add(key, value)

	handlers, ok := s.Load(key)
	if assert.True(t, ok) {
		assert.Equal(t, value, handlers[0])
		assert.Len(t, handlers, 1)
	}

	missingKey := NewTopic(internal.RandString(10))
	handlers, ok = s.Load(missingKey)
	if assert.False(t, ok) {
		assert.Nil(t, handlers)
	}
}

func TestSubscribers_Del(t *testing.T) {
	t.Parallel()

	s := &subscribers{
		m: make(map[Topic][]*handler),
	}

	key := NewTopic(internal.RandString(10))
	value := &handler{}
	s.Add(key, value)

	s.Del(key)

	val, ok := s.Load(key)
	if assert.False(t, ok) {
		assert.Nil(t, val)
	}
}

func TestSubscribers_Add(t *testing.T) {
	t.Parallel()

	s := &subscribers{
		m: make(map[Topic][]*handler),
	}

	key := NewTopic(internal.RandString(10))
	h1, h2 := &handler{}, &handler{}
	s.Add(key, h1, h2)

	handlers, ok := s.Load(key)
	if assert.True(t, ok) {
		assert.Equal(t, h1, handlers[0])
		assert.Equal(t, h2, handlers[1])
		assert.Len(t, handlers, 2)
	}

	assert.Equal(t, 1, s.Len())
}

func TestSubscribers_Concurrency(t *testing.T) {
	t.Parallel()

	s := &subscribers{
		m: make(map[Topic][]*handler),
	}

	var wg sync.WaitGroup
	numGoroutines := 100
	wg.Add(numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(i int) {
			defer wg.Done()
			key := NewTopic(fmt.Sprintf("key_%d", i))
			value := &handler{}

			s.Add(key, value)

			handlers, ok := s.Load(key)
			if assert.True(t, ok) {
				assert.Equal(t, value, handlers[0])
				assert.Len(t, handlers, 1)
			}
		}(i)
	}

	wg.Wait()
}

func TestSubscribers_Len(t *testing.T) {
	t.Parallel()

	s := &subscribers{
		m: make(map[Topic][]*handler),
	}

	key := NewTopic(internal.RandString(10))
	h1, h2, h3 := &handler{}, &handler{}, &handler{}
	s.Add(key, h1, h2, h3)
	assert.Equal(t, 1, s.Len())
}
