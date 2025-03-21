package inmemory

import "sync"

type subscribers struct {
	sync.RWMutex
	m map[string]*handler
}

func (s *subscribers) Load(key string) (*handler, bool) {
	s.RLock()
	val, ok := s.m[key]
	s.RUnlock()

	return val, ok
}

func (s *subscribers) Store(key string, val *handler) {
	s.Lock()
	defer s.Unlock()

	s.m[key] = val
}

func (s *subscribers) Del(key string) {
	s.Lock()
	delete(s.m, key)
	s.Unlock()
}
