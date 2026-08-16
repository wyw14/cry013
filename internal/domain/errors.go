package domain

import "errors"

var (
	ErrNotFound          = errors.New("not found")
	ErrForbidden         = errors.New("forbidden")
	ErrConflict          = errors.New("conflict")
	ErrInvalidTransition = errors.New("invalid state transition")
	ErrTokenReplayed     = errors.New("refresh token replayed")
	ErrTokenRevoked      = errors.New("refresh token revoked")
)
