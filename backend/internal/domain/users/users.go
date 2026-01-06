package users

import (
	"errors"
	"strings"
)

// User holds identity and profile-only fields. Balance, auth, and market data live in their own domains.
type User struct {
	ID            int64
	Username      string
	DisplayName   string
	PersonalLink  string
	PersonalEmoji string
}

// NewUser constructs a User with the minimum required fields.
func NewUser(id int64, username, displayName, personalLink, personalEmoji string) User {
	return User{
		ID:            id,
		Username:      username,
		DisplayName:   displayName,
		PersonalLink:  personalLink,
		PersonalEmoji: personalEmoji,
	}
}

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
