package auth

import (
	"strings"
	"time"
)

// APIKey represents an issued API key for a user.
type APIKey struct {
	UserID    int64
	Key       string
	CreatedAt time.Time
	ExpiresAt time.Time
}

// Validate ensures the key shape and timestamps are sensible.
func (k APIKey) Validate() error {
	if k.UserID <= 0 {
		return ErrInvalidCredentials
	}
	if strings.TrimSpace(k.Key) == "" {
		return ErrInvalidCredentials
	}
	if !k.ExpiresAt.IsZero() && k.CreatedAt.After(k.ExpiresAt) {
		return ErrInvalidCredentials
	}
	return nil
}

// Rotate replaces the key material and timestamps.
func (k *APIKey) Rotate(newKey string, createdAt, expiresAt time.Time) error {
	newKey = strings.TrimSpace(newKey)
	if newKey == "" {
		return ErrInvalidCredentials
	}
	if createdAt.IsZero() {
		return ErrInvalidCredentials
	}
	if !expiresAt.IsZero() && createdAt.After(expiresAt) {
		return ErrInvalidCredentials
	}
	k.Key = newKey
	k.CreatedAt = createdAt
	k.ExpiresAt = expiresAt
	return nil
}

// IsExpired reports whether the key is past its expiry at the given time.
func (k APIKey) IsExpired(at time.Time) bool {
	if k.ExpiresAt.IsZero() {
		return false
	}
	if at.IsZero() {
		at = time.Now()
	}
	return !at.Before(k.ExpiresAt)
}
