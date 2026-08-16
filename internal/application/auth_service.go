package application

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/wyw14/cry013/internal/domain"
)

type AuthService struct {
	repo       Repository
	clock      Clock
	ids        IDGenerator
	passwords  PasswordHasher
	tokens     TokenManager
	refreshTTL time.Duration
}

type AuthTokens struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresAt    time.Time `json:"expires_at"`
}

func NewAuthService(repo Repository, clock Clock, ids IDGenerator, passwords PasswordHasher, tokens TokenManager, refreshTTL time.Duration) *AuthService {
	return &AuthService{repo: repo, clock: clock, ids: ids, passwords: passwords, tokens: tokens, refreshTTL: refreshTTL}
}

func (s *AuthService) Register(ctx context.Context, email, password, displayName string) (domain.User, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	if email == "" || len(password) < 10 || strings.TrimSpace(displayName) == "" {
		return domain.User{}, errors.New("invalid registration fields")
	}
	if _, err := s.repo.UserByEmail(ctx, email); err == nil {
		return domain.User{}, domain.ErrConflict
	}
	hash, err := s.passwords.Hash(password)
	if err != nil {
		return domain.User{}, err
	}
	now := s.clock.Now()
	user := domain.User{ID: s.ids.New(), Email: email, PasswordHash: hash, DisplayName: displayName, Role: domain.PlatformUser, Status: domain.UserActive, CreatedAt: now, UpdatedAt: now}
	if err := s.repo.CreateUser(ctx, user); err != nil {
		return domain.User{}, err
	}
	return user, nil
}

func (s *AuthService) Login(ctx context.Context, email, password string) (AuthTokens, error) {
	user, err := s.repo.UserByEmail(ctx, strings.ToLower(strings.TrimSpace(email)))
	if err != nil || user.Status != domain.UserActive {
		return AuthTokens{}, domain.ErrForbidden
	}
	if err := s.passwords.Compare(user.PasswordHash, password); err != nil {
		return AuthTokens{}, domain.ErrForbidden
	}
	return s.issue(ctx, user, "")
}

func (s *AuthService) Refresh(ctx context.Context, plain string) (AuthTokens, error) {
	now := s.clock.Now()
	hash := s.tokens.HashRefresh(plain)
	stored, err := s.repo.RefreshToken(ctx, hash)
	if err != nil {
		return AuthTokens{}, err
	}
	if err := stored.CanRotate(now); err != nil {
		return AuthTokens{}, err
	}
	stored.UsedAt = &now
	if err := s.repo.UpdateRefreshToken(ctx, stored); err != nil {
		return AuthTokens{}, err
	}
	user, err := s.repo.UserByID(ctx, stored.UserID)
	if err != nil || user.Status != domain.UserActive {
		_ = s.repo.RevokeTokenFamily(ctx, stored.FamilyID, now)
		return AuthTokens{}, domain.ErrForbidden
	}
	return s.issue(ctx, user, stored.FamilyID)
}

func (s *AuthService) Logout(ctx context.Context, plain string) error {
	hash := s.tokens.HashRefresh(plain)
	stored, err := s.repo.RefreshToken(ctx, hash)
	if err != nil {
		return nil
	}
	return s.repo.RevokeTokenFamily(ctx, stored.FamilyID, s.clock.Now())
}

func (s *AuthService) issue(ctx context.Context, user domain.User, familyID string) (AuthTokens, error) {
	access, err := s.tokens.IssueAccess(user)
	if err != nil {
		return AuthTokens{}, err
	}
	plain, hash, generatedFamily, err := s.tokens.NewRefreshToken()
	if err != nil {
		return AuthTokens{}, err
	}
	if familyID == "" {
		familyID = generatedFamily
	}
	expires := s.clock.Now().Add(s.refreshTTL)
	if err := s.repo.StoreRefreshToken(ctx, domain.RefreshToken{Hash: hash, UserID: user.ID, FamilyID: familyID, ExpiresAt: expires}); err != nil {
		return AuthTokens{}, err
	}
	return AuthTokens{AccessToken: access, RefreshToken: plain, ExpiresAt: expires}, nil
}
