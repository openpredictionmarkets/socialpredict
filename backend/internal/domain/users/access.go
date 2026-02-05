package users

// Credentials holds authentication-related fields for a user.
type Credentials struct {
	UserID             int64
	PasswordHash       string
	MustChangePassword bool
}
