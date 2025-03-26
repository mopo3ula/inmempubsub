package inmemory

const ConcurrencyUnlimited = 0

type Option func(*handler)

func WithMaxConcurrency(maxConcurrency int) Option {
	return func(h *handler) {
		h.maxConcurrency = maxConcurrency
	}
}
