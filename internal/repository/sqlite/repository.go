package sqlite

import (
	"context"
	"errors"
	"time"

	"github.com/wyw14/cry013/internal/domain"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Repository struct{ db *gorm.DB }

func NewRepository(db *gorm.DB) *Repository { return &Repository{db: db} }

func mapError(err error) error {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return domain.ErrNotFound
	}
	if errors.Is(err, gorm.ErrDuplicatedKey) {
		return domain.ErrConflict
	}
	return err
}

func userFromModel(m UserModel) domain.User {
	return domain.User{ID: m.ID, Email: m.Email, PasswordHash: m.PasswordHash, DisplayName: m.DisplayName, AvatarURL: m.AvatarURL, Role: domain.PlatformRole(m.Role), Status: domain.UserStatus(m.Status), CreatedAt: m.CreatedAt, UpdatedAt: m.UpdatedAt}
}

func userModel(u domain.User) UserModel {
	return UserModel{ID: u.ID, Email: u.Email, PasswordHash: u.PasswordHash, DisplayName: u.DisplayName, AvatarURL: u.AvatarURL, Role: string(u.Role), Status: string(u.Status), CreatedAt: u.CreatedAt, UpdatedAt: u.UpdatedAt}
}

func (r *Repository) CreateUser(ctx context.Context, u domain.User) error {
	return mapError(r.db.WithContext(ctx).Create(&[]UserModel{userModel(u)}).Error)
}

func (r *Repository) UpdateUser(ctx context.Context, u domain.User) error {
	result := r.db.WithContext(ctx).Model(&UserModel{}).Where("id = ?", u.ID).Updates(userModel(u))
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *Repository) UserByID(ctx context.Context, id string) (domain.User, error) {
	var m UserModel
	err := r.db.WithContext(ctx).First(&m, "id = ?", id).Error
	return userFromModel(m), mapError(err)
}

func (r *Repository) UserByEmail(ctx context.Context, email string) (domain.User, error) {
	var m UserModel
	err := r.db.WithContext(ctx).First(&m, "lower(email) = lower(?)", email).Error
	return userFromModel(m), mapError(err)
}

func (r *Repository) ListUsers(ctx context.Context) ([]domain.User, error) {
	var models []UserModel
	if err := r.db.WithContext(ctx).Order("created_at").Find(&models).Error; err != nil {
		return nil, err
	}
	items := make([]domain.User, len(models))
	for i, m := range models {
		items[i] = userFromModel(m)
	}
	return items, nil
}

func vaultFromModel(m VaultModel) domain.Vault {
	return domain.Vault{ID: m.ID, Name: m.Name, OwnerID: m.OwnerID, CreatedAt: m.CreatedAt, UpdatedAt: m.UpdatedAt}
}

func vaultModel(w domain.Vault) VaultModel {
	return VaultModel{ID: w.ID, Name: w.Name, OwnerID: w.OwnerID, CreatedAt: w.CreatedAt, UpdatedAt: w.UpdatedAt}
}

func (r *Repository) CreateVault(ctx context.Context, w domain.Vault) error {
	return mapError(r.db.WithContext(ctx).Create(&[]VaultModel{vaultModel(w)}).Error)
}

func (r *Repository) UpdateVault(ctx context.Context, w domain.Vault) error {
	result := r.db.WithContext(ctx).Model(&VaultModel{}).Where("id = ?", w.ID).Updates(vaultModel(w))
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *Repository) VaultByID(ctx context.Context, id string) (domain.Vault, error) {
	var m VaultModel
	err := r.db.WithContext(ctx).First(&m, "id = ?", id).Error
	return vaultFromModel(m), mapError(err)
}

func (r *Repository) ListVaults(ctx context.Context) ([]domain.Vault, error) {
	var models []VaultModel
	if err := r.db.WithContext(ctx).Order("created_at").Find(&models).Error; err != nil {
		return nil, err
	}
	items := make([]domain.Vault, len(models))
	for i, m := range models {
		items[i] = vaultFromModel(m)
	}
	return items, nil
}

func vaultLeaseFromModel(m VaultLeaseModel) domain.VaultLease {
	return domain.VaultLease{VaultID: m.VaultID, UserID: m.UserID, Role: domain.VaultRole(m.Role), Status: domain.VaultLeaseStatus(m.Status), JoinedAt: m.JoinedAt}
}

func vaultLeaseModel(m domain.VaultLease) VaultLeaseModel {
	return VaultLeaseModel{VaultID: m.VaultID, UserID: m.UserID, Role: string(m.Role), Status: string(m.Status), JoinedAt: m.JoinedAt}
}

func (r *Repository) UpsertVaultLease(ctx context.Context, m domain.VaultLease) error {
	model := vaultLeaseModel(m)
	return r.db.WithContext(ctx).Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "vault_id"}, {Name: "user_id"}}, DoUpdates: clause.AssignmentColumns([]string{"role", "status", "joined_at"})}).Create(&model).Error
}

func (r *Repository) VaultLease(ctx context.Context, vaultID, userID string) (domain.VaultLease, error) {
	var m VaultLeaseModel
	err := r.db.WithContext(ctx).First(&m, "vault_id = ? AND user_id = ?", vaultID, userID).Error
	return vaultLeaseFromModel(m), mapError(err)
}

func (r *Repository) ListVaultLeases(ctx context.Context, vaultID string) ([]domain.VaultLease, error) {
	var models []VaultLeaseModel
	if err := r.db.WithContext(ctx).Where("vault_id = ?", vaultID).Find(&models).Error; err != nil {
		return nil, err
	}
	items := make([]domain.VaultLease, len(models))
	for i, m := range models {
		items[i] = vaultLeaseFromModel(m)
	}
	return items, nil
}

func (r *Repository) RestoreVault(ctx context.Context, vaultID, fromID, toID string) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var ws VaultModel
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&ws, "id = ?", vaultID).Error; err != nil {
			return mapError(err)
		}
		if ws.OwnerID != fromID {
			return domain.ErrInvalidTransition
		}
		if err := tx.Model(&VaultLeaseModel{}).Where("vault_id = ? AND user_id = ?", vaultID, fromID).Update("role", string(domain.RoleAdmin)).Error; err != nil {
			return err
		}
		result := tx.Model(&VaultLeaseModel{}).Where("vault_id = ? AND user_id = ? AND status = ?", vaultID, toID, string(domain.VaultLeaseActive)).Update("role", string(domain.RoleOwner))
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected != 1 {
			return domain.ErrInvalidTransition
		}
		return tx.Model(&VaultModel{}).Where("id = ?", vaultID).Update("owner_id", toID).Error
	})
}

func backupFromModel(m BackupModel) domain.Backup {
	return domain.Backup{ID: m.ID, VaultID: m.VaultID, Email: m.Email, Role: domain.VaultRole(m.Role), CreateBackupdBy: m.CreateBackupdBy, Status: domain.BackupStatus(m.Status), ExpiresAt: m.ExpiresAt, CreatedAt: m.CreatedAt}
}

func backupModel(i domain.Backup) BackupModel {
	return BackupModel{ID: i.ID, VaultID: i.VaultID, Email: i.Email, Role: string(i.Role), CreateBackupdBy: i.CreateBackupdBy, Status: string(i.Status), ExpiresAt: i.ExpiresAt, CreatedAt: i.CreatedAt}
}

func (r *Repository) CreateBackup(ctx context.Context, i domain.Backup) error {
	return mapError(r.db.WithContext(ctx).Create(&[]BackupModel{backupModel(i)}).Error)
}
func (r *Repository) BackupByID(ctx context.Context, id string) (domain.Backup, error) {
	var m BackupModel
	err := r.db.WithContext(ctx).First(&m, "id = ?", id).Error
	return backupFromModel(m), mapError(err)
}
func (r *Repository) UpdateBackup(ctx context.Context, i domain.Backup) error {
	result := r.db.WithContext(ctx).Model(&BackupModel{}).Where("id = ?", i.ID).Updates(backupModel(i))
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return domain.ErrNotFound
	}
	return nil
}
func (r *Repository) ListBackups(ctx context.Context, vaultID string) ([]domain.Backup, error) {
	var models []BackupModel
	if err := r.db.WithContext(ctx).Where("vault_id = ?", vaultID).Find(&models).Error; err != nil {
		return nil, err
	}
	items := make([]domain.Backup, len(models))
	for i, m := range models {
		items[i] = backupFromModel(m)
	}
	return items, nil
}

func entryFromModel(m EntryModel) domain.Entry {
	return domain.Entry{ID: m.ID, VaultID: m.VaultID, Title: m.Title, Body: m.Body, Type: m.Type, Visibility: domain.EntryVisibility(m.Visibility), Status: domain.EntryStatus(m.Status), Priority: m.Priority, Tags: append([]string(nil), m.Tags...), CreatorID: m.CreatorID, AssigneeID: m.AssigneeID, CreatedAt: m.CreatedAt, UpdatedAt: m.UpdatedAt, DeletedAt: m.DeletedAt}
}
func entryModel(c domain.Entry) EntryModel {
	return EntryModel{ID: c.ID, VaultID: c.VaultID, Title: c.Title, Body: c.Body, Type: c.Type, Visibility: string(c.Visibility), Status: string(c.Status), Priority: c.Priority, Tags: append([]string(nil), c.Tags...), CreatorID: c.CreatorID, AssigneeID: c.AssigneeID, CreatedAt: c.CreatedAt, UpdatedAt: c.UpdatedAt, DeletedAt: c.DeletedAt}
}
func (r *Repository) CreateEntry(ctx context.Context, c domain.Entry) error {
	return mapError(r.db.WithContext(ctx).Create(&[]EntryModel{entryModel(c)}).Error)
}
func (r *Repository) UpdateEntry(ctx context.Context, c domain.Entry) error {
	result := r.db.WithContext(ctx).Model(&EntryModel{}).Where("id = ?", c.ID).Updates(entryModel(c))
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return domain.ErrNotFound
	}
	return nil
}
func (r *Repository) EntryByID(ctx context.Context, id string) (domain.Entry, error) {
	var m EntryModel
	err := r.db.WithContext(ctx).First(&m, "id = ? AND deleted_at IS NULL", id).Error
	return entryFromModel(m), mapError(err)
}
func (r *Repository) ListEntries(ctx context.Context, vaultID string) ([]domain.Entry, error) {
	var models []EntryModel
	if err := r.db.WithContext(ctx).Where("vault_id = ? AND deleted_at IS NULL", vaultID).Order("updated_at DESC").Find(&models).Error; err != nil {
		return nil, err
	}
	items := make([]domain.Entry, len(models))
	for i, m := range models {
		items[i] = entryFromModel(m)
	}
	return items, nil
}

func (r *Repository) CreateComment(ctx context.Context, c domain.Comment) error {
	m := CommentModel{ID: c.ID, EntryID: c.EntryID, AuthorID: c.AuthorID, Body: c.Body, Mentions: c.Mentions, CreatedAt: c.CreatedAt}
	return mapError(r.db.WithContext(ctx).Create(&[]CommentModel{m}).Error)
}
func (r *Repository) ListComments(ctx context.Context, entryID string) ([]domain.Comment, error) {
	var models []CommentModel
	if err := r.db.WithContext(ctx).Where("entry_id = ?", entryID).Order("created_at").Find(&models).Error; err != nil {
		return nil, err
	}
	items := make([]domain.Comment, len(models))
	for i, m := range models {
		items[i] = domain.Comment{ID: m.ID, EntryID: m.EntryID, AuthorID: m.AuthorID, Body: m.Body, Mentions: m.Mentions, CreatedAt: m.CreatedAt}
	}
	return items, nil
}
func (r *Repository) AppendActivity(ctx context.Context, a domain.Activity) error {
	m := ActivityModel{ID: a.ID, VaultID: a.VaultID, ActorID: a.ActorID, EntryID: a.EntryID, Kind: a.Kind, Metadata: a.Metadata, CreatedAt: a.CreatedAt}
	return mapError(r.db.WithContext(ctx).Create(&[]ActivityModel{m}).Error)
}
func (r *Repository) ListActivities(ctx context.Context, vaultID string) ([]domain.Activity, error) {
	var models []ActivityModel
	if err := r.db.WithContext(ctx).Where("vault_id = ?", vaultID).Order("created_at DESC").Limit(200).Find(&models).Error; err != nil {
		return nil, err
	}
	items := make([]domain.Activity, len(models))
	for i, m := range models {
		items[i] = domain.Activity{ID: m.ID, VaultID: m.VaultID, ActorID: m.ActorID, EntryID: m.EntryID, Kind: m.Kind, Metadata: m.Metadata, CreatedAt: m.CreatedAt}
	}
	return items, nil
}
func (r *Repository) AppendAudit(ctx context.Context, a domain.AuditEvent) error {
	m := AuditModel{ID: a.ID, ActorID: a.ActorID, VaultID: a.VaultID, Action: a.Action, TargetType: a.TargetType, TargetID: a.TargetID, RequestID: a.RequestID, CreatedAt: a.CreatedAt}
	return mapError(r.db.WithContext(ctx).Create(&[]AuditModel{m}).Error)
}
func (r *Repository) ListAudits(ctx context.Context) ([]domain.AuditEvent, error) {
	var models []AuditModel
	if err := r.db.WithContext(ctx).Order("created_at DESC").Limit(1000).Find(&models).Error; err != nil {
		return nil, err
	}
	items := make([]domain.AuditEvent, len(models))
	for i, m := range models {
		items[i] = domain.AuditEvent{ID: m.ID, ActorID: m.ActorID, VaultID: m.VaultID, Action: m.Action, TargetType: m.TargetType, TargetID: m.TargetID, RequestID: m.RequestID, CreatedAt: m.CreatedAt}
	}
	return items, nil
}

func tokenFromModel(m RefreshTokenModel) domain.RefreshToken {
	return domain.RefreshToken{Hash: m.Hash, UserID: m.UserID, FamilyID: m.FamilyID, ExpiresAt: m.ExpiresAt, UsedAt: m.UsedAt, RevokedAt: m.RevokedAt}
}
func tokenModel(t domain.RefreshToken) RefreshTokenModel {
	return RefreshTokenModel{Hash: t.Hash, UserID: t.UserID, FamilyID: t.FamilyID, ExpiresAt: t.ExpiresAt, UsedAt: t.UsedAt, RevokedAt: t.RevokedAt}
}
func (r *Repository) StoreRefreshToken(ctx context.Context, t domain.RefreshToken) error {
	return mapError(r.db.WithContext(ctx).Create(&[]RefreshTokenModel{tokenModel(t)}).Error)
}
func (r *Repository) RefreshToken(ctx context.Context, hash string) (domain.RefreshToken, error) {
	var m RefreshTokenModel
	err := r.db.WithContext(ctx).First(&m, "hash = ?", hash).Error
	return tokenFromModel(m), mapError(err)
}
func (r *Repository) UpdateRefreshToken(ctx context.Context, t domain.RefreshToken) error {
	return r.db.WithContext(ctx).Save(&RefreshTokenModel{Hash: t.Hash, UserID: t.UserID, FamilyID: t.FamilyID, ExpiresAt: t.ExpiresAt, UsedAt: t.UsedAt, RevokedAt: t.RevokedAt}).Error
}

func (r *Repository) ConsumeRefreshToken(ctx context.Context, hash string, now time.Time) (domain.RefreshToken, error) {
	var result domain.RefreshToken
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var m RefreshTokenModel
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&m, "hash = ?", hash).Error; err != nil {
			return mapError(err)
		}
		token := tokenFromModel(m)
		if err := token.CanRotate(now); err != nil {
			return err
		}
		token.UsedAt = &now
		if err := tx.Model(&RefreshTokenModel{}).Where("hash = ? AND used_at IS NULL AND revoked_at IS NULL", hash).Update("used_at", now).Error; err != nil {
			return err
		}
		result = token
		return nil
	})
	return result, err
}

func (r *Repository) RevokeTokenFamily(ctx context.Context, familyID string, now time.Time) error {
	return r.db.WithContext(ctx).Model(&RefreshTokenModel{}).Where("family_id = ? AND revoked_at IS NULL", familyID).Update("revoked_at", now).Error
}

func (r *Repository) DoIdempotent(ctx context.Context, scope string, fn func() (string, error)) (string, error) {
	entry := IdempotencyModel{Scope: scope, State: "running", CreatedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC()}
	result := r.db.WithContext(ctx).Clauses(clause.OnConflict{DoNothing: true}).Create(&entry)
	if result.Error != nil {
		return "", result.Error
	}
	if result.RowsAffected == 0 {
		var existing IdempotencyModel
		for i := 0; i < 100; i++ {
			if err := r.db.WithContext(ctx).First(&existing, "scope = ?", scope).Error; err != nil {
				return "", err
			}
			if existing.State == "done" {
				return existing.Value, nil
			}
			select {
			case <-ctx.Done():
				return "", ctx.Err()
			case <-time.After(10 * time.Millisecond):
			}
		}
		return "", errors.New("idempotency operation timed out")
	}
	value, err := fn()
	if err != nil {
		_ = r.db.WithContext(ctx).Delete(&IdempotencyModel{}, "scope = ?", scope).Error
		return "", err
	}
	if err := r.db.WithContext(ctx).Model(&IdempotencyModel{}).Where("scope = ?", scope).Updates(map[string]any{"state": "done", "value": value, "updated_at": time.Now().UTC()}).Error; err != nil {
		return "", err
	}
	return value, nil
}
