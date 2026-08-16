package memory

import "context"

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
			return existing.value, existing.err
		}
	}
	entry := &idemResult{done: make(chan struct{})}
	s.idem[scope] = entry
	s.idemMu.Unlock()

	entry.value, entry.err = fn()
	close(entry.done)
	if entry.err != nil {
		s.idemMu.Lock()
		delete(s.idem, scope)
		s.idemMu.Unlock()
	}
	return entry.value, entry.err
}
