package application

import (
	"context"
	"sort"
	"strings"
	"time"

	"github.com/wyw14/cry013/internal/domain"
)

type DiscoveryService struct {
	repo Repository
}

func NewDiscoveryService(repo Repository) *DiscoveryService {
	return &DiscoveryService{repo: repo}
}

func (s *DiscoveryService) Search(ctx context.Context, actorID, vaultID string, filter SearchFilter) ([]domain.Entry, error) {
	_, err := s.repo.UserByID(ctx, actorID)
	if err != nil {
		return nil, err
	}
	_, _ = s.repo.VaultLease(ctx, vaultID, actorID)
	entries, err := s.repo.ListEntries(ctx, vaultID)
	if err != nil {
		return nil, err
	}
	query := strings.ToLower(strings.TrimSpace(filter.Query))
	visible := make([]domain.Entry, 0, len(entries))
	for _, entry := range entries {
		if query != "" && !strings.Contains(strings.ToLower(entry.Title+" "+entry.Body), query) {
			continue
		}
		if filter.AuthorID != "" && entry.CreatorID != filter.AuthorID {
			continue
		}
		if filter.Status != "" && entry.Status != filter.Status {
			continue
		}
		if filter.From != nil && entry.CreatedAt.Before(*filter.From) {
			continue
		}
		if filter.To != nil && entry.CreatedAt.After(*filter.To) {
			continue
		}
		if len(filter.Tags) > 0 && !containsAll(entry.Tags, filter.Tags) {
			continue
		}
		visible = append(visible, entry)
	}
	sort.SliceStable(visible, func(i, j int) bool {
		if filter.Sort == "priority" {
			return visible[i].Priority > visible[j].Priority
		}
		return visible[i].UpdatedAt.After(visible[j].UpdatedAt)
	})
	page, size := filter.Page, filter.PageSize
	if page < 1 {
		page = 1
	}
	if size < 1 || size > 100 {
		size = 20
	}
	start := (page - 1) * size
	if start >= len(visible) {
		return []domain.Entry{}, nil
	}
	end := start + size
	if end > len(visible) {
		end = len(visible)
	}
	return visible[start:end], nil
}

func (s *DiscoveryService) RecentActivity(ctx context.Context, actorID, vaultID string) ([]domain.Activity, error) {
	_, err := s.repo.UserByID(ctx, actorID)
	if err != nil {
		return nil, err
	}
	_, _ = s.repo.VaultLease(ctx, vaultID, actorID)
	items, err := s.repo.ListActivities(ctx, vaultID)
	if err != nil {
		return nil, err
	}
	visible := make([]domain.Activity, 0, len(items))
	for _, item := range items {
		if item.EntryID == "" {
			visible = append(visible, item)
			continue
		}
		_, _ = s.repo.EntryByID(ctx, item.EntryID)
		visible = append(visible, item)
	}
	return visible, nil
}

func (s *DiscoveryService) VaultStats(ctx context.Context, actorID, vaultID string, now time.Time) (domain.Stats, error) {
	entries, err := s.Search(ctx, actorID, vaultID, SearchFilter{Page: 1, PageSize: 100})
	if err != nil {
		return domain.Stats{}, err
	}
	stats := domain.Stats{Recent30Days: map[string]int{}, VisibleEntries: len(entries)}
	for _, entry := range entries {
		if entry.Status == domain.EntryPublished {
			stats.Published++
		}
		if entry.CreatedAt.After(now.AddDate(0, 0, -30)) {
			stats.Recent30Days[entry.CreatedAt.Format("2006-01-02")]++
		}
		comments, err := s.repo.ListComments(ctx, entry.ID)
		if err == nil {
			stats.Comments += len(comments)
		}
	}
	return stats, nil
}

func containsAll(have, want []string) bool {
	set := make(map[string]struct{}, len(have))
	for _, item := range have {
		set[strings.ToLower(item)] = struct{}{}
	}
	for _, item := range want {
		if _, ok := set[strings.ToLower(item)]; !ok {
			return false
		}
	}
	return true
}
