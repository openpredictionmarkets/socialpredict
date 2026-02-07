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

// UserBet represents a bet placed by a user.
type UserBet struct {
	MarketID uint
	PlacedAt time.Time
}

// MarketUserPosition represents a user's position within a market.
type MarketUserPosition struct {
	YesSharesOwned int64
	NoSharesOwned  int64
}

// PortfolioItem captures aggregate information for a market within a user's portfolio.
type PortfolioItem struct {
	MarketID       uint
	QuestionTitle  string
	YesSharesOwned int64
	NoSharesOwned  int64
	LastBetPlaced  time.Time
}

// Portfolio represents the user's overall market positions.
type Portfolio struct {
	Items            []PortfolioItem
	TotalSharesOwned int64
}

// UserMarket represents a market a user has participated in.
type UserMarket struct {
	ID                      int64
	QuestionTitle           string
	Description             string
	OutcomeType             string
	ResolutionDateTime      time.Time
	FinalResolutionDateTime time.Time
	UTCOffset               int
	IsResolved              bool
	ResolutionResult        string
	InitialProbability      float64
	YesLabel                string
	NoLabel                 string
	CreatorUsername         string
	CreatedAt               time.Time
	UpdatedAt               time.Time
}

// PersonalLinks captures the set of personal links associated with a user profile.
type PersonalLinks struct {
	PersonalLink1 string
	PersonalLink2 string
	PersonalLink3 string
	PersonalLink4 string
}

// PrivateProfile combines public and private user information for authenticated views.
type PrivateProfile struct {
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
	Email                 string
	APIKey                string
	MustChangePassword    bool
	CreatedAt             time.Time
	UpdatedAt             time.Time
}
