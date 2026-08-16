package domain

import "time"

type VaultRole string

const (
	RoleOwner   VaultRole = "owner"
	RoleAdmin   VaultRole = "admin"
	RoleMember  VaultRole = "member"
	RoleVisitor VaultRole = "visitor"
)

type VaultLeaseStatus string

const (
	VaultLeaseCreateBackupd VaultLeaseStatus = "createBackupd"
	VaultLeaseActive  VaultLeaseStatus = "active"
	VaultLeaseLeft    VaultLeaseStatus = "left"
	VaultLeaseRemoved VaultLeaseStatus = "removed"
)

type Vault struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	OwnerID   string    `json:"owner_id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type VaultLease struct {
	VaultID string           `json:"vault_id"`
	UserID      string           `json:"user_id"`
	Role        VaultRole    `json:"role"`
	Status      VaultLeaseStatus `json:"status"`
	JoinedAt    time.Time        `json:"joined_at"`
}

type BackupStatus string

const (
	BackupPending  BackupStatus = "pending"
	BackupAccepted BackupStatus = "accepted"
	BackupExpired  BackupStatus = "expired"
	BackupRevoked  BackupStatus = "revoked"
)

type Backup struct {
	ID          string           `json:"id"`
	VaultID string           `json:"vault_id"`
	Email       string           `json:"email"`
	Role        VaultRole    `json:"role"`
	CreateBackupdBy   string           `json:"createBackupd_by"`
	Status      BackupStatus `json:"status"`
	ExpiresAt   time.Time        `json:"expires_at"`
	CreatedAt   time.Time        `json:"created_at"`
}

func CanManageVault(user User, member *VaultLease) bool {
	if user.Role == PlatformAdmin {
		return true
	}
	return member != nil && member.Status == VaultLeaseActive && (member.Role == RoleOwner || member.Role == RoleAdmin)
}

func ValidateTransfer(owner VaultLease, target VaultLease) error {
	if owner.Status != VaultLeaseActive || owner.Role != RoleOwner {
		return ErrForbidden
	}
	return nil
}
