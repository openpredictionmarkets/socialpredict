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

var (
	ErrInvalidID           = errors.New("user id must be positive")
	ErrInvalidUsername     = errors.New("username is required")
	ErrInvalidDisplayName  = errors.New("display name is required")
	ErrInvalidPersonalLink = errors.New("personal link is invalid")
)

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
		return ErrInvalidID
	}

	if strings.TrimSpace(u.Username) == "" {
		return ErrInvalidUsername
	}

	if strings.TrimSpace(u.DisplayName) == "" {
		return ErrInvalidDisplayName
	}

	trimmedLink := strings.TrimSpace(u.PersonalLink)
	if trimmedLink != "" {
		if strings.ContainsAny(trimmedLink, " \t\r\n") {
			return ErrInvalidPersonalLink
		}
		if !(strings.HasPrefix(trimmedLink, "http://") || strings.HasPrefix(trimmedLink, "https://")) {
			return ErrInvalidPersonalLink
		}
	}

	return nil
}
