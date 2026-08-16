package application

import (
	"context"
	"time"

	"github.com/wyw14/cry013/internal/domain"
)

type AdminService struct {
	repo  Repository
	clock Clock
	ids   IDGenerator
}

func NewAdminService(repo Repository, clock Clock, ids IDGenerator) *AdminService {
	return &AdminService{repo: repo, clock: clock, ids: ids}
}

func (s *AdminService) SetUserStatus(ctx context.Context, actorID, userID string, status domain.UserStatus, meta RequestMeta) error {
	actor, err := s.repo.UserByID(ctx, actorID)
	if err != nil || actor.Role != domain.PlatformAdmin {
		return domain.ErrForbidden
	}
	user, err := s.repo.UserByID(ctx, userID)
	if err != nil {
		return err
	}
	user.Status = status
	user.UpdatedAt = s.clock.Now()
	if err := s.repo.UpdateUser(ctx, user); err != nil {
		return err
	}
	return s.repo.AppendAudit(ctx, domain.AuditEvent{ID: s.ids.New(), ActorID: actorID, Action: "user.status", TargetType: "user", TargetID: userID, Metadata: map[string]any{"status": status}, RequestID: meta.RequestID, CreatedAt: s.clock.Now()})
}

func (s *AdminService) PlatformStats(ctx context.Context, actorID string, now time.Time) (domain.Stats, error) {
	actor, err := s.repo.UserByID(ctx, actorID)
	if err != nil || actor.Role != domain.PlatformAdmin {
		return domain.Stats{}, domain.ErrForbidden
	}
	users, err := s.repo.ListUsers(ctx)
	if err != nil {
		return domain.Stats{}, err
	}
	vaults, err := s.repo.ListVaults(ctx)
	if err != nil {
		return domain.Stats{}, err
	}
	stats := domain.Stats{Users: len(users), Vaults: len(vaults), Recent30Days: map[string]int{}}
	for _, ws := range vaults {
		entries, err := s.repo.ListEntries(ctx, ws.ID)
		if err != nil {
			continue
		}
		for _, entry := range entries {
			if entry.DeletedAt != nil {
				stats.PrivateHidden++
				continue
			}
			stats.VisibleEntries++
			if entry.Status == domain.EntryPublished {
				stats.Published++
			}
			if entry.CreatedAt.After(now.AddDate(0, 0, -30)) {
				stats.Recent30Days[entry.CreatedAt.Format("2006-01-02")]++
			}
			comments, _ := s.repo.ListComments(ctx, entry.ID)
			stats.Comments += len(comments)
		}
	}
	return stats, nil
}
