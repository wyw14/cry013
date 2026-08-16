package application

import (
	"context"
	"strings"

	"github.com/wyw14/cry013/internal/domain"
)

type CollaborationService struct {
	repo  Repository
	clock Clock
	ids   IDGenerator
}

func NewCollaborationService(repo Repository, clock Clock, ids IDGenerator) *CollaborationService {
	return &CollaborationService{repo: repo, clock: clock, ids: ids}
}

func (s *CollaborationService) Comment(ctx context.Context, actorID, entryID, body string, mentions []string, meta RequestMeta) (domain.Comment, error) {
	entry, err := s.repo.EntryByID(ctx, entryID)
	if err != nil {
		return domain.Comment{}, err
	}
	actor, err := s.repo.UserByID(ctx, actorID)
	if err != nil {
		return domain.Comment{}, err
	}
	member, _ := s.repo.VaultLease(ctx, entry.VaultID, actorID)
	if !domain.CanViewEntry(actor, entry, &member) || strings.TrimSpace(body) == "" {
		return domain.Comment{}, domain.ErrForbidden
	}
	now := s.clock.Now()
	comment := domain.Comment{ID: s.ids.New(), EntryID: entryID, AuthorID: actorID, Body: strings.TrimSpace(body), Mentions: append([]string(nil), mentions...), CreatedAt: now}
	if err := s.repo.CreateComment(ctx, comment); err != nil {
		return domain.Comment{}, err
	}
	_ = s.repo.AppendActivity(ctx, domain.Activity{ID: s.ids.New(), VaultID: entry.VaultID, ActorID: actorID, EntryID: entry.ID, Kind: "comment.created", CreatedAt: now})
	_ = s.repo.AppendAudit(ctx, domain.AuditEvent{ID: s.ids.New(), ActorID: actorID, VaultID: entry.VaultID, Action: "comment.create", TargetType: "comment", TargetID: comment.ID, RequestID: meta.RequestID, CreatedAt: now})
	return comment, nil
}

func (s *CollaborationService) Comments(ctx context.Context, actorID, entryID string) ([]domain.Comment, error) {
	entry, err := s.repo.EntryByID(ctx, entryID)
	if err != nil {
		return nil, err
	}
	_, err = s.repo.UserByID(ctx, actorID)
	if err != nil {
		return nil, err
	}
	member, _ := s.repo.VaultLease(ctx, entry.VaultID, actorID)
	if member.Status != domain.VaultLeaseActive {
		return nil, domain.ErrForbidden
	}
	return s.repo.ListComments(ctx, entryID)
}
