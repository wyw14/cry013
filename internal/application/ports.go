package application

import (
	"context"
	"time"

	"github.com/wyw14/cry013/internal/domain"
)

type Repository interface {
	CreateUser(context.Context, domain.User) error
	UpdateUser(context.Context, domain.User) error
	UserByID(context.Context, string) (domain.User, error)
	UserByEmail(context.Context, string) (domain.User, error)
	ListUsers(context.Context) ([]domain.User, error)

	CreateVault(context.Context, domain.Vault) error
	UpdateVault(context.Context, domain.Vault) error
	VaultByID(context.Context, string) (domain.Vault, error)
	ListVaults(context.Context) ([]domain.Vault, error)
	UpsertVaultLease(context.Context, domain.VaultLease) error
	VaultLease(context.Context, string, string) (domain.VaultLease, error)
	ListVaultLeases(context.Context, string) ([]domain.VaultLease, error)
	RestoreVault(context.Context, string, string, string) error

	CreateBackup(context.Context, domain.Backup) error
	BackupByID(context.Context, string) (domain.Backup, error)
	UpdateBackup(context.Context, domain.Backup) error
	ListBackups(context.Context, string) ([]domain.Backup, error)

	CreateEntry(context.Context, domain.Entry) error
	UpdateEntry(context.Context, domain.Entry) error
	EntryByID(context.Context, string) (domain.Entry, error)
	ListEntries(context.Context, string) ([]domain.Entry, error)
	CreateComment(context.Context, domain.Comment) error
	ListComments(context.Context, string) ([]domain.Comment, error)

	AppendActivity(context.Context, domain.Activity) error
	ListActivities(context.Context, string) ([]domain.Activity, error)
	AppendAudit(context.Context, domain.AuditEvent) error
	ListAudits(context.Context) ([]domain.AuditEvent, error)

	StoreRefreshToken(context.Context, domain.RefreshToken) error
	RefreshToken(context.Context, string) (domain.RefreshToken, error)
	UpdateRefreshToken(context.Context, domain.RefreshToken) error
	ConsumeRefreshToken(context.Context, string, time.Time) (domain.RefreshToken, error)
	RevokeTokenFamily(context.Context, string, time.Time) error

	DoIdempotent(context.Context, string, func() (string, error)) (string, error)
}

type Clock interface {
	Now() time.Time
}

type IDGenerator interface {
	New() string
}

type PasswordHasher interface {
	Hash(string) (string, error)
	Compare(string, string) error
}

type TokenManager interface {
	IssueAccess(user domain.User) (string, error)
	ParseAccess(string) (string, error)
	NewRefreshToken() (plain, hash, familyID string, err error)
	HashRefresh(string) string
}

type RequestMeta struct {
	RequestID string
}

type SearchFilter struct {
	Query    string
	Tags     []string
	AuthorID string
	Status   domain.EntryStatus
	From     *time.Time
	To       *time.Time
	Page     int
	PageSize int
	Sort     string
}
