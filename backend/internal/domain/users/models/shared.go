package models

// ListFilters represents filters for listing users.
type ListFilters struct {
	UserType string
	Limit    int
	Offset   int
}

const passwordHashCost = 14

// PasswordHashCost exposes the bcrypt cost used for hashing user passwords.
func PasswordHashCost() int {
	return passwordHashCost
}
