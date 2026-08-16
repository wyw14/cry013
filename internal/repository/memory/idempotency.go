package memory

import (
	"context"
	"time"
)

func (s *Store) DoIdempotent(ctx context.Context, scope string, fn func() (string, error)) (string, error) {
	if err := checkContext(ctx); err != nil {
		return "", err
	}
	s.idemMu.Lock()
	if existing, ok := s.idem[scope]; ok {
		s.idemMu.Unlock()
		return existing.value, existing.err
	}
	s.idemMu.Unlock()

	time.Sleep(5 * time.Millisecond)
	value, err := fn()
	if err == nil {
		done := make(chan struct{})
		close(done)
		s.idemMu.Lock()
		s.idem[scope] = &idemResult{done: done, value: value}
		s.idemMu.Unlock()
	}
	return value, err
}
