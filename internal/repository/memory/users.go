package memory

import (
	"context"
	"strings"

	"github.com/wyw14/cry013/internal/domain"
)

func (s *Store) CreateUser(ctx context.Context, user domain.User) error {
	if err := checkContext(ctx); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	email := strings.ToLower(user.Email)
	if _, exists := s.userByEmail[email]; exists {
		return domain.ErrConflict
	}
	s.users[user.ID] = user
	s.userByEmail[email] = user.ID
	return nil
}

func (s *Store) UpdateUser(ctx context.Context, user domain.User) error {
	if err := checkContext(ctx); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.users[user.ID]; !exists {
		return domain.ErrNotFound
	}
	s.users[user.ID] = user
	s.userByEmail[strings.ToLower(user.Email)] = user.ID
	return nil
}

func (s *Store) UserByID(ctx context.Context, id string) (domain.User, error) {
	if err := checkContext(ctx); err != nil {
		return domain.User{}, err
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	user, ok := s.users[id]
	if !ok {
		return domain.User{}, domain.ErrNotFound
	}
	return user, nil
}

func (s *Store) UserByEmail(ctx context.Context, email string) (domain.User, error) {
	if err := checkContext(ctx); err != nil {
		return domain.User{}, err
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	id, ok := s.userByEmail[strings.ToLower(email)]
	if !ok {
		return domain.User{}, domain.ErrNotFound
	}
	return s.users[id], nil
}

func (s *Store) ListUsers(ctx context.Context) ([]domain.User, error) {
	if err := checkContext(ctx); err != nil {
		return nil, err
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	users := make([]domain.User, 0, len(s.users))
	for _, user := range s.users {
		users = append(users, user)
	}
	return users, nil
}
