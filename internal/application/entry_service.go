package application

import (
	"context"
	"strings"

	"github.com/wyw14/cry013/internal/domain"
)

type EntryService struct {
	repo  Repository
	clock Clock
	ids   IDGenerator
}

type CreateEntryInput struct {
	VaultID string
	Title       string
	Body        string
	Type        string
	Visibility  domain.EntryVisibility
	Priority    int
	Tags        []string
	AssigneeID  string
}

func NewEntryService(repo Repository, clock Clock, ids IDGenerator) *EntryService {
	return &EntryService{repo: repo, clock: clock, ids: ids}
}

func (s *EntryService) Create(ctx context.Context, actorID, idempotencyKey string, in CreateEntryInput, meta RequestMeta) (domain.Entry, error) {
	member, err := s.repo.VaultLease(ctx, in.VaultID, actorID)
	if err != nil || member.Status != domain.VaultLeaseActive || member.Role == domain.RoleVisitor {
		return domain.Entry{}, domain.ErrForbidden
	}
	if strings.TrimSpace(in.Title) == "" || idempotencyKey == "" {
		return domain.Entry{}, domain.ErrInvalidTransition
	}
	scope := domain.IdempotencyScope(in.VaultID, actorID, "entry.create", meta.RequestID)
	entryID, err := s.repo.DoIdempotent(ctx, scope, func() (string, error) {
		now := s.clock.Now()
		entry := domain.Entry{ID: s.ids.New(), VaultID: in.VaultID, Title: strings.TrimSpace(in.Title), Body: in.Body, Type: in.Type, Visibility: in.Visibility, Status: domain.EntryDraft, Priority: in.Priority, Tags: append([]string(nil), in.Tags...), CreatorID: actorID, AssigneeID: in.AssigneeID, CreatedAt: now, UpdatedAt: now}
		if err := s.repo.CreateEntry(ctx, entry); err != nil {
			return "", err
		}
		if err := s.repo.AppendActivity(ctx, domain.Activity{ID: s.ids.New(), VaultID: in.VaultID, ActorID: actorID, EntryID: entry.ID, Kind: "entry.created", CreatedAt: now}); err != nil {
			return "", err
		}
		_ = s.repo.AppendAudit(ctx, domain.AuditEvent{ID: s.ids.New(), ActorID: actorID, VaultID: in.VaultID, Action: "entry.create", TargetType: "entry", TargetID: entry.ID, RequestID: meta.RequestID, CreatedAt: now})
		return entry.ID, nil
	})
	if err != nil {
		return domain.Entry{}, err
	}
	return s.repo.EntryByID(ctx, entryID)
}

func (s *EntryService) Transition(ctx context.Context, actorID, entryID string, to domain.EntryStatus, meta RequestMeta) (domain.Entry, error) {
	entry, err := s.repo.EntryByID(ctx, entryID)
	if err != nil {
		return domain.Entry{}, err
	}
	actor, err := s.repo.UserByID(ctx, actorID)
	if err != nil {
		return domain.Entry{}, err
	}
	member, _ := s.repo.VaultLease(ctx, entry.VaultID, actorID)
	if !domain.CanEditEntry(actor, entry, &member) || !entry.CanTransition(to) {
		return domain.Entry{}, domain.ErrForbidden
	}
	entry.Status = to
	entry.UpdatedAt = s.clock.Now()
	if err := s.repo.UpdateEntry(ctx, entry); err != nil {
		return domain.Entry{}, err
	}
	_ = s.repo.AppendActivity(ctx, domain.Activity{ID: s.ids.New(), VaultID: entry.VaultID, ActorID: actorID, EntryID: entry.ID, Kind: "entry." + string(to), CreatedAt: entry.UpdatedAt})
	_ = s.repo.AppendAudit(ctx, domain.AuditEvent{ID: s.ids.New(), ActorID: actorID, VaultID: entry.VaultID, Action: "entry.transition", TargetType: "entry", TargetID: entry.ID, Metadata: map[string]any{"status": to}, RequestID: meta.RequestID, CreatedAt: entry.UpdatedAt})
	return entry, nil
}
