package inmemory

import (
	"sync"
)

type subscribers struct {
	sync.RWMutex
	m map[Topic][]*handler
}

func (s *subscribers) Load(topic Topic) ([]*handler, bool) {
	s.RLock()
	val, ok := s.m[topic]
	s.RUnlock()

	return val, ok
}

func (s *subscribers) Add(topic Topic, handlers ...*handler) {
	for _, h := range handlers {
		s.Lock()
		s.m[topic] = append(s.m[topic], h)
		s.Unlock()
	}
}

func (s *subscribers) Del(topic Topic) {
	s.Lock()
	delete(s.m, topic)
	s.Unlock()
}

func (s *subscribers) Len() int {
	return len(s.m)
}
