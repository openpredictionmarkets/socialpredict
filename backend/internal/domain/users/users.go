package users

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
