package domain

import "time"

type PlatformRole string

const (
	PlatformAdmin PlatformRole = "platform_admin"
	PlatformUser  PlatformRole = "user"
)

type UserStatus string

const (
	UserActive   UserStatus = "active"
	UserDisabled UserStatus = "disabled"
)

type User struct {
	ID           string       `json:"id"`
	Email        string       `json:"email"`
	PasswordHash string       `json:"-"`
	DisplayName  string       `json:"display_name"`
	AvatarURL    string       `json:"avatar_url,omitempty"`
	Role         PlatformRole `json:"role"`
	Status       UserStatus   `json:"status"`
	CreatedAt    time.Time    `json:"created_at"`
	UpdatedAt    time.Time    `json:"updated_at"`
}

type RefreshToken struct {
	Hash      string
	UserID    string
	FamilyID  string
	ExpiresAt time.Time
	UsedAt    *time.Time
	RevokedAt *time.Time
}

func (t RefreshToken) CanRotate(now time.Time) error {
	if t.RevokedAt != nil {
		return ErrTokenRevoked
	}
	if !now.Before(t.ExpiresAt) {
		return ErrTokenRevoked
	}
	if t.UsedAt != nil {
		return ErrTokenReplayed
	}
	return nil
}
