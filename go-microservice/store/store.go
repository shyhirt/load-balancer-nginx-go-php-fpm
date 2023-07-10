package store

import (
	"golang.org/x/time/rate"
	"sync"
	"time"
)

type visitor struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

type Store struct {
	mut sync.Mutex
	m   map[string]*visitor
	lim int
}

func New(limitPerSec int) *Store {
	return &Store{
		mut: sync.Mutex{},
		m:   make(map[string]*visitor),
		lim: limitPerSec,
	}
}

func (s *Store) Allow(token string) bool {
	s.mut.Lock()
	t, ok := s.m[token]
	s.mut.Unlock()
	if !ok {
		s.mut.Lock()
		limiter := rate.NewLimiter(1, s.lim)
		s.m[token] = &visitor{limiter, time.Now()}
		s.mut.Unlock()
		return true
	}
	t.lastSeen = time.Now()
	return t.limiter.Allow()
}
