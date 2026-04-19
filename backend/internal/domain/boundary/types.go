package boundary

import (
	"time"

	"golang.org/x/crypto/bcrypt"
)

// Bet captures the persistence-neutral wager fields used by domain and math code.
type Bet struct {
	ID        uint
	Username  string
	MarketID  uint
	Amount    int64
	Outcome   string
	PlacedAt  time.Time
	CreatedAt time.Time
}

// Market captures the persistence-neutral market fields used by analytics flows.
type Market struct {
	ID               uint
	CreatedAt        time.Time
	IsResolved       bool
	ResolutionResult string
}

// User captures the persistence-neutral user fields used by analytics flows.
type User struct {
	Username       string
	AccountBalance int64
}

// AuthenticatedUser captures the auth-facing fields needed during login.
type AuthenticatedUser struct {
	Username           string
	UserType           string
	PasswordHash       string
	MustChangePassword bool
}

// CheckPasswordHash validates the supplied password against the stored hash.
func (u AuthenticatedUser) CheckPasswordHash(password string) bool {
	return bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password)) == nil
}
