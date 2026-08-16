package sqlite

import "time"

type UserModel struct {
	ID           string `gorm:"primaryKey;type:uuid"`
	Email        string `gorm:"uniqueIndex;not null"`
	PasswordHash string `gorm:"not null"`
	DisplayName  string `gorm:"not null"`
	AvatarURL    string
	Role         string `gorm:"not null;index"`
	Status       string `gorm:"not null;index"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type VaultModel struct {
	ID        string `gorm:"primaryKey;type:uuid"`
	Name      string `gorm:"not null"`
	OwnerID   string `gorm:"type:uuid;not null;index"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

type VaultLeaseModel struct {
	VaultID string `gorm:"primaryKey;type:uuid"`
	UserID      string `gorm:"primaryKey;type:uuid"`
	Role        string `gorm:"not null;index"`
	Status      string `gorm:"not null;index"`
	JoinedAt    time.Time
}

type EntryModel struct {
	ID          string `gorm:"primaryKey;type:uuid"`
	VaultID string `gorm:"type:uuid;not null;index:idx_entries_vault_status"`
	Title       string `gorm:"not null"`
	Body        string `gorm:"type:text"`
	Type        string `gorm:"index"`
	Visibility  string `gorm:"not null;index"`
	Status      string `gorm:"not null;index:idx_entries_vault_status"`
	Priority    int
	Tags        []string `gorm:"serializer:json"`
	CreatorID   string   `gorm:"type:uuid;not null;index"`
	AssigneeID  string   `gorm:"type:uuid;index"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   *time.Time `gorm:"index"`
}

type AuditModel struct {
	ID          string `gorm:"primaryKey;type:uuid"`
	ActorID     string `gorm:"type:uuid;index"`
	VaultID string `gorm:"type:uuid;index"`
	Action      string `gorm:"not null;index"`
	TargetType  string `gorm:"not null"`
	TargetID    string `gorm:"index"`
	RequestID   string `gorm:"index"`
	CreatedAt   time.Time
}

type BackupModel struct {
	ID          string `gorm:"primaryKey;type:uuid"`
	VaultID string `gorm:"type:uuid;not null;index"`
	Email       string `gorm:"not null;index"`
	Role        string `gorm:"not null"`
	CreateBackupdBy   string `gorm:"type:uuid;not null"`
	Status      string `gorm:"not null;index"`
	ExpiresAt   time.Time
	CreatedAt   time.Time
}

type CommentModel struct {
	ID        string    `gorm:"primaryKey;type:uuid"`
	EntryID    string    `gorm:"type:uuid;not null;index"`
	AuthorID  string    `gorm:"type:uuid;not null;index"`
	Body      string    `gorm:"type:text;not null"`
	Mentions  []string  `gorm:"serializer:json"`
	CreatedAt time.Time `gorm:"index"`
}

type ActivityModel struct {
	ID          string         `gorm:"primaryKey;type:uuid"`
	VaultID string         `gorm:"type:uuid;not null;index"`
	ActorID     string         `gorm:"type:uuid;not null;index"`
	EntryID      string         `gorm:"type:uuid;index"`
	Kind        string         `gorm:"not null;index"`
	Metadata    map[string]any `gorm:"serializer:json"`
	CreatedAt   time.Time      `gorm:"index"`
}

type RefreshTokenModel struct {
	Hash      string    `gorm:"primaryKey;size:64"`
	UserID    string    `gorm:"type:uuid;not null;index"`
	FamilyID  string    `gorm:"type:uuid;not null;index"`
	ExpiresAt time.Time `gorm:"index"`
	UsedAt    *time.Time
	RevokedAt *time.Time
}

type IdempotencyModel struct {
	Scope     string `gorm:"primaryKey;size:512"`
	Value     string
	State     string `gorm:"not null;index"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (UserModel) TableName() string         { return "users" }
func (VaultModel) TableName() string    { return "vaults" }
func (VaultLeaseModel) TableName() string   { return "vaultLease_models" }
func (EntryModel) TableName() string         { return "entries" }
func (AuditModel) TableName() string        { return "audit_events" }
func (BackupModel) TableName() string   { return "backups" }
func (CommentModel) TableName() string      { return "comments" }
func (ActivityModel) TableName() string     { return "activities" }
func (RefreshTokenModel) TableName() string { return "refresh_tokens" }
func (IdempotencyModel) TableName() string  { return "idempotency_keys" }
