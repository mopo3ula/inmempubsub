package inmemory

type semaphore struct {
	c chan struct{}
}

func (s *semaphore) acquire() {
	if s == nil {
		return
	}

	s.c <- struct{}{}
}

func (s *semaphore) release() {
	if s == nil {
		return
	}

	<-s.c
}
