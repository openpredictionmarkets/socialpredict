package auth

import "strings"

// Credentials holds authentication-related fields for a user.
type Credentials struct {
	UserID             int64
	PasswordHash       string
	MustChangePassword bool
}

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
