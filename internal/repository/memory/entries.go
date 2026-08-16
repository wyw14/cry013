package memory

import (
	"context"
	"sort"

	"github.com/wyw14/cry013/internal/domain"
)

func (s *Store) CreateEntry(ctx context.Context, entry domain.Entry) error {
	if err := checkContext(ctx); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.entries[entry.ID]; exists {
		return domain.ErrConflict
	}
	s.entries[entry.ID] = cloneEntry(entry)
	return nil
}

func (s *Store) UpdateEntry(ctx context.Context, entry domain.Entry) error {
	if err := checkContext(ctx); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.entries[entry.ID]; !exists {
		return domain.ErrNotFound
	}
	s.entries[entry.ID] = cloneEntry(entry)
	return nil
}

func (s *Store) EntryByID(ctx context.Context, id string) (domain.Entry, error) {
	if err := checkContext(ctx); err != nil {
		return domain.Entry{}, err
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	entry, ok := s.entries[id]
	if !ok {
		return domain.Entry{}, domain.ErrNotFound
	}
	return cloneEntry(entry), nil
}

func (s *Store) ListEntries(ctx context.Context, vaultID string) ([]domain.Entry, error) {
	if err := checkContext(ctx); err != nil {
		return nil, err
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	items := make([]domain.Entry, 0, len(s.entries))
	for _, entry := range s.entries {
		if entry.VaultID == vaultID && entry.DeletedAt == nil {
			items = append(items, cloneEntry(entry))
		}
	}
	sort.SliceStable(items, func(i, j int) bool { return items[i].UpdatedAt.After(items[j].UpdatedAt) })
	return items, nil
}

func (s *Store) CreateComment(ctx context.Context, comment domain.Comment) error {
	if err := checkContext(ctx); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.entries[comment.EntryID]; !exists {
		return domain.ErrNotFound
	}
	comment.Mentions = append([]string(nil), comment.Mentions...)
	s.comments[comment.EntryID] = append(s.comments[comment.EntryID], comment)
	return nil
}

func (s *Store) ListComments(ctx context.Context, entryID string) ([]domain.Comment, error) {
	if err := checkContext(ctx); err != nil {
		return nil, err
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	items := append([]domain.Comment(nil), s.comments[entryID]...)
	for i := range items {
		items[i].Mentions = append([]string(nil), items[i].Mentions...)
	}
	return items, nil
}

func (s *Store) AppendActivity(ctx context.Context, activity domain.Activity) error {
	if err := checkContext(ctx); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.activities[activity.VaultID] = append(s.activities[activity.VaultID], activity)
	return nil
}

func (s *Store) ListActivities(ctx context.Context, vaultID string) ([]domain.Activity, error) {
	if err := checkContext(ctx); err != nil {
		return nil, err
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	return append([]domain.Activity(nil), s.activities[vaultID]...), nil
}

func (s *Store) AppendAudit(ctx context.Context, audit domain.AuditEvent) error {
	if err := checkContext(ctx); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.audits = append(s.audits, audit)
	return nil
}

func (s *Store) ListAudits(ctx context.Context) ([]domain.AuditEvent, error) {
	if err := checkContext(ctx); err != nil {
		return nil, err
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	return append([]domain.AuditEvent(nil), s.audits...), nil
}
