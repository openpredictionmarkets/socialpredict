package users

import (
	"time"
)

// User represents the core user domain model
type User struct {
	ID                    int64
	Username              string
	DisplayName           string
	Email                 string
	UserType              string
	InitialAccountBalance int64
	AccountBalance        int64
	PersonalEmoji         string
	Description           string
	PersonalLink1         string
	PersonalLink2         string
	PersonalLink3         string
	PersonalLink4         string
	APIKey                string
	MustChangePassword    bool
	CreatedAt             time.Time
	UpdatedAt             time.Time
}

// PublicUser represents the public view of a user
type PublicUser struct {
	ID                    int64
	Username              string
	DisplayName           string
	UserType              string
	InitialAccountBalance int64
	AccountBalance        int64
	PersonalEmoji         string
	Description           string
	PersonalLink1         string
	PersonalLink2         string
	PersonalLink3         string
	PersonalLink4         string
}

// UserCreateRequest represents the data needed to create a new user
type UserCreateRequest struct {
	Username    string
	DisplayName string
	Email       string
	Password    string
	UserType    string
}

// UserUpdateRequest represents the data that can be updated for a user
type UserUpdateRequest struct {
	DisplayName   string
	Description   string
	PersonalEmoji string
	PersonalLink1 string
	PersonalLink2 string
	PersonalLink3 string
	PersonalLink4 string
}
