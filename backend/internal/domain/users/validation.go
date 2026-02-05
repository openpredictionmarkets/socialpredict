// This file exists to put all of the validation functions in the package in one place.

// We declare the package

package users

// imports go here

import (
	"strings"
	"time"
)

// Validate ensures credential data is well-formed.
func (c Credentials) Validate() error {
	if c.UserID <= 0 {
		return ErrInvalidCredentials
	}
	if strings.TrimSpace(c.PasswordHash) == "" {
		return ErrInvalidCredentials
	}
	return nil
}

// RotateHash updates the stored hash and clears the forced-change flag.
func (c *Credentials) RotateHash(newHash string) error {
	newHash = strings.TrimSpace(newHash)
	if newHash == "" {
		return ErrInvalidCredentials
	}
	c.PasswordHash = newHash
	c.MustChangePassword = false
	return nil
}

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

// User validation

// Validate performs basic sanity checks on required fields and simple link hygiene.
func (u User) Validate() error {
	if u.ID <= 0 {
		return ErrInvalidUserData
	}

	if strings.TrimSpace(u.Username) == "" {
		return ErrInvalidUserData
	}

	if strings.TrimSpace(u.DisplayName) == "" {
		return ErrInvalidUserData
	}

	trimmedLink := strings.TrimSpace(u.PersonalLink)
	if trimmedLink != "" {
		if strings.ContainsAny(trimmedLink, " \t\r\n") {
			return ErrInvalidUserData
		}
		if !(strings.HasPrefix(trimmedLink, "http://") || strings.HasPrefix(trimmedLink, "https://")) {
			return ErrInvalidUserData
		}
	}

	return nil
}
