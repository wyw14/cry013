package memory

import (
	"context"
	"time"

	"github.com/wyw14/cry013/internal/domain"
)

func (s *Store) StoreRefreshToken(ctx context.Context, token domain.RefreshToken) error {
	if err := checkContext(ctx); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.tokens[token.Hash]; exists {
		return domain.ErrConflict
	}
	s.tokens[token.Hash] = token
	return nil
}

func (s *Store) RefreshToken(ctx context.Context, hash string) (domain.RefreshToken, error) {
	if err := checkContext(ctx); err != nil {
		return domain.RefreshToken{}, err
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	token, ok := s.tokens[hash]
	if !ok {
		return domain.RefreshToken{}, domain.ErrNotFound
	}
	return token, nil
}

func (s *Store) UpdateRefreshToken(ctx context.Context, token domain.RefreshToken) error {
	if err := checkContext(ctx); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.tokens[token.Hash] = token
	return nil
}

func (s *Store) ConsumeRefreshToken(ctx context.Context, hash string, now time.Time) (domain.RefreshToken, error) {
	if err := checkContext(ctx); err != nil {
		return domain.RefreshToken{}, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	token, ok := s.tokens[hash]
	if !ok {
		return domain.RefreshToken{}, domain.ErrNotFound
	}
	if err := token.CanRotate(now); err != nil {
		return domain.RefreshToken{}, err
	}
	token.UsedAt = &now
	s.tokens[hash] = token
	return token, nil
}

func (s *Store) RevokeTokenFamily(ctx context.Context, familyID string, now time.Time) error {
	if err := checkContext(ctx); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	for hash, token := range s.tokens {
		if token.FamilyID == familyID {
			token.RevokedAt = &now
			s.tokens[hash] = token
		}
	}
	return nil
}
