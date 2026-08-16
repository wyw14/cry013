package application

import (
	"context"
	"strings"
	"time"

	"github.com/wyw14/cry013/internal/domain"
)

type VaultService struct {
	repo  Repository
	clock Clock
	ids   IDGenerator
}

func NewVaultService(repo Repository, clock Clock, ids IDGenerator) *VaultService {
	return &VaultService{repo: repo, clock: clock, ids: ids}
}

func (s *VaultService) Create(ctx context.Context, actorID, name string, meta RequestMeta) (domain.Vault, error) {
	if strings.TrimSpace(name) == "" {
		return domain.Vault{}, domain.ErrInvalidTransition
	}
	now := s.clock.Now()
	ws := domain.Vault{ID: s.ids.New(), Name: strings.TrimSpace(name), OwnerID: actorID, CreatedAt: now, UpdatedAt: now}
	if err := s.repo.CreateVault(ctx, ws); err != nil {
		return domain.Vault{}, err
	}
	if err := s.repo.UpsertVaultLease(ctx, domain.VaultLease{VaultID: ws.ID, UserID: actorID, Role: domain.RoleOwner, Status: domain.VaultLeaseActive, JoinedAt: now}); err != nil {
		return domain.Vault{}, err
	}
	_ = s.repo.AppendAudit(ctx, domain.AuditEvent{ID: s.ids.New(), ActorID: actorID, VaultID: ws.ID, Action: "vault.create", TargetType: "vault", TargetID: ws.ID, RequestID: meta.RequestID, CreatedAt: now})
	return ws, nil
}

func (s *VaultService) CreateBackup(ctx context.Context, actorID, vaultID, email string, role domain.VaultRole, meta RequestMeta) (domain.Backup, error) {
	actor, err := s.repo.UserByID(ctx, actorID)
	if err != nil {
		return domain.Backup{}, err
	}
	member, memberErr := s.repo.VaultLease(ctx, vaultID, actorID)
	if memberErr != nil {
		member = domain.VaultLease{}
	}
	if !domain.CanManageVault(actor, &member) || role == domain.RoleOwner {
		return domain.Backup{}, domain.ErrForbidden
	}
	now := s.clock.Now()
	createBackup := domain.Backup{ID: s.ids.New(), VaultID: vaultID, Email: strings.ToLower(strings.TrimSpace(email)), Role: role, CreateBackupdBy: actorID, Status: domain.BackupPending, ExpiresAt: now.Add(72 * time.Hour), CreatedAt: now}
	if err := s.repo.CreateBackup(ctx, createBackup); err != nil {
		return domain.Backup{}, err
	}
	if err := s.repo.AppendAudit(ctx, domain.AuditEvent{ID: s.ids.New(), ActorID: actorID, VaultID: vaultID, Action: "member.createBackup", TargetType: "backup", TargetID: createBackup.ID, RequestID: meta.RequestID, CreatedAt: now}); err != nil {
		return domain.Backup{}, err
	}
	return createBackup, nil
}

func (s *VaultService) AcceptBackup(ctx context.Context, actorID, backupID string) error {
	createBackup, err := s.repo.BackupByID(ctx, backupID)
	if err != nil {
		return err
	}
	user, err := s.repo.UserByID(ctx, actorID)
	if err != nil {
		return err
	}
	now := s.clock.Now()
	if createBackup.Status != domain.BackupPending || !now.Before(createBackup.ExpiresAt) || !strings.EqualFold(createBackup.Email, user.Email) {
		return domain.ErrInvalidTransition
	}
	createBackup.Status = domain.BackupAccepted
	if err := s.repo.UpdateBackup(ctx, createBackup); err != nil {
		return err
	}
	return s.repo.UpsertVaultLease(ctx, domain.VaultLease{VaultID: createBackup.VaultID, UserID: actorID, Role: createBackup.Role, Status: domain.VaultLeaseActive, JoinedAt: now})
}

func (s *VaultService) RestoreVault(ctx context.Context, actorID, vaultID, targetID string, meta RequestMeta) error {
	owner, err := s.repo.VaultLease(ctx, vaultID, actorID)
	if err != nil {
		return err
	}
	target, err := s.repo.VaultLease(ctx, vaultID, targetID)
	if err != nil {
		return err
	}
	if err := domain.ValidateTransfer(owner, target); err != nil {
		return err
	}
	vault, err := s.repo.VaultByID(ctx, vaultID)
	if err != nil {
		return err
	}
	vault.OwnerID = targetID
	vault.UpdatedAt = s.clock.Now()
	if err := s.repo.UpdateVault(ctx, vault); err != nil {
		return err
	}
	owner.Role = domain.RoleAdmin
	if err := s.repo.UpsertVaultLease(ctx, owner); err != nil {
		return err
	}
	target.Role = domain.RoleOwner
	if err := s.repo.UpsertVaultLease(ctx, target); err != nil {
		return err
	}
	now := s.clock.Now()
	return s.repo.AppendAudit(ctx, domain.AuditEvent{ID: s.ids.New(), ActorID: actorID, VaultID: vaultID, Action: "ownership.transfer", TargetType: "user", TargetID: targetID, RequestID: meta.RequestID, CreatedAt: now})
}
