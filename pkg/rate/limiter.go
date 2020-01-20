package rate

import (
	"sync"
)

type Limiter struct {
	mux     sync.Mutex
	limit   int
	current int
}

func NewLimiter(limit int) *Limiter {
	return &Limiter{
		limit: limit,
	}
}

func (l *Limiter) In() bool {
	l.mux.Lock()
	defer l.mux.Unlock()
	if l.current < l.limit {
		l.current++
		return true
	}
	return false
}

func (l *Limiter) Out() {
	l.mux.Lock()
	defer l.mux.Unlock()
	l.current--
}
