package service

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/wyw14/cry013/internal/domain"
)

type JWTManager struct {
	Secret    []byte
	AccessTTL time.Duration
	Clock     func() time.Time
}

func (m JWTManager) now() time.Time {
	if m.Clock != nil {
		return m.Clock()
	}
	return time.Now().UTC()
}

func (m JWTManager) IssueAccess(user domain.User) (string, error) {
	now := m.now()
	claims := jwt.MapClaims{
		"sub":  user.ID,
		"role": user.Role,
		"iat":  now.Unix(),
		"exp":  now.Add(m.AccessTTL).Unix(),
		"jti":  uuid.NewString(),
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(m.Secret)
}

func (m JWTManager) ParseAccess(raw string) (string, error) {
	token, err := jwt.Parse(raw, func(token *jwt.Token) (any, error) {
		if token.Method != jwt.SigningMethodHS256 {
			return nil, errors.New("unexpected signing method")
		}
		return m.Secret, nil
	}, jwt.WithExpirationRequired())
	if err != nil || !token.Valid {
		return "", domain.ErrForbidden
	}
	subject, err := token.Claims.GetSubject()
	if err != nil || subject == "" {
		return "", domain.ErrForbidden
	}
	return subject, nil
}

func (m JWTManager) NewRefreshToken() (plain, hash, familyID string, err error) {
	buf := make([]byte, 32)
	if _, err = rand.Read(buf); err != nil {
		return "", "", "", err
	}
	plain = base64.RawURLEncoding.EncodeToString(buf)
	return plain, m.HashRefresh(plain), uuid.NewString(), nil
}

func (JWTManager) HashRefresh(plain string) string {
	sum := sha256.Sum256([]byte(plain))
	return hex.EncodeToString(sum[:])
}
