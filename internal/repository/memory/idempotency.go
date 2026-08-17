package memory

import (
	"context"
)

func (s *Store) DoIdempotent(ctx context.Context, scope string, fn func() (string, error)) (string, error) {
	if err := checkContext(ctx); err != nil {
		return "", err
	}
	s.idemMu.Lock()
	if existing, ok := s.idem[scope]; ok {
		s.idemMu.Unlock()
		select {
		case <-ctx.Done():
			return "", ctx.Err()
		case <-existing.done:
		}
		return existing.value, existing.err
	}
	result := &idemResult{done: make(chan struct{})}
	s.idem[scope] = result
	s.idemMu.Unlock()

	value, err := fn()
	result.value = value
	result.err = err
	close(result.done)

	if err != nil {
		s.idemMu.Lock()
		if current, ok := s.idem[scope]; ok && current == result {
			delete(s.idem, scope)
		}
		s.idemMu.Unlock()
	}
	return value, err
}
