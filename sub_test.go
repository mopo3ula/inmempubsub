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
		m: make(map[string]*handler),
	}

	key := internal.RandString(10)
	value := &handler{}
	s.Store(key, value)

	val, ok := s.Load(key)
	if assert.True(t, ok) {
		assert.Equal(t, value, val)
	}

	missingKey := internal.RandString(10)
	val, ok = s.Load(missingKey)
	if assert.False(t, ok) {
		assert.Nil(t, val)
	}
}

func TestSubscribers_Store(t *testing.T) {
	t.Parallel()

	s := &subscribers{
		m: make(map[string]*handler),
	}

	key := internal.RandString(10)
	value := &handler{}
	s.Store(key, value)

	val, ok := s.Load(key)
	if assert.True(t, ok) {
		assert.Equal(t, value, val)
	}
}

func TestSubscribers_Del(t *testing.T) {
	t.Parallel()

	s := &subscribers{
		m: make(map[string]*handler),
	}

	key := internal.RandString(10)
	value := &handler{}
	s.Store(key, value)

	s.Del(key)

	val, ok := s.Load(key)
	if assert.False(t, ok) {
		assert.Nil(t, val)
	}
}

func TestSubscribers_Concurrency(t *testing.T) {
	t.Parallel()

	s := &subscribers{
		m: make(map[string]*handler),
	}

	var wg sync.WaitGroup
	numGoroutines := 100
	wg.Add(numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(i int) {
			defer wg.Done()
			key := fmt.Sprintf("key_%d", i)
			value := &handler{}

			s.Store(key, value)

			val, ok := s.Load(key)
			if assert.True(t, ok) {
				assert.Equal(t, value, val)
			}
		}(i)
	}

	wg.Wait()
}
