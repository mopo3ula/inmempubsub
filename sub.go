package inmemory

import (
	"sync"
)

type subscribers struct {
	sync.RWMutex
	m map[string][]*handler
}

func (s *subscribers) Load(key string) ([]*handler, bool) {
	s.RLock()
	val, ok := s.m[key]
	s.RUnlock()

	return val, ok
}

func (s *subscribers) Add(key string, handlers ...*handler) {
	for _, h := range handlers {
		s.Lock()
		s.m[key] = append(s.m[key], h)
		s.Unlock()
	}
}

func (s *subscribers) Del(key string) {
	s.Lock()
	delete(s.m, key)
	s.Unlock()
}

func (s *subscribers) Len() int {
	return len(s.m)
}
