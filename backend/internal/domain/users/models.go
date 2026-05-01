package users

import (
	"time"
)

const personalLinkFieldCount = 4

// PersonalLinksTarget exposes the minimal contract needed to receive personal links.
type PersonalLinksTarget interface {
	ApplyPersonalLinks(PersonalLinks)
}

// User represents the core user domain model
type User struct {
	ID                    int64
	Username              string
	DisplayName           string
	Email                 string
	PasswordHash          string
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

// ToPublicUser returns the public-safe projection of the user model.
func (u *User) ToPublicUser() *PublicUser {
	return new(PublicUser).FromUser(u)
}

// ToPrivateProfile returns the authenticated private projection of the user model.
func (u *User) ToPrivateProfile() *PrivateProfile {
	return new(PrivateProfile).FromUser(u)
}

func copyPublicUser(target *PublicUser, user *User) *PublicUser {
	if user == nil {
		return nil
	}

	if target == nil {
		target = &PublicUser{}
	}

	*target = PublicUser{
		ID:                    user.ID,
		Username:              user.Username,
		DisplayName:           user.DisplayName,
		UserType:              user.UserType,
		InitialAccountBalance: user.InitialAccountBalance,
		AccountBalance:        user.AccountBalance,
		PersonalEmoji:         user.PersonalEmoji,
		Description:           user.Description,
	}
	user.PersonalLinks().ApplyTo(target)
	return target
}

// FromUser copies the public-safe projection of the user model.
func (u *PublicUser) FromUser(user *User) *PublicUser {
	return copyPublicUser(u, user)
}

func copyPrivateProfile(target *PrivateProfile, user *User) *PrivateProfile {
	if user == nil {
		return nil
	}

	if target == nil {
		target = &PrivateProfile{}
	}

	*target = PrivateProfile{
		ID:                    user.ID,
		Username:              user.Username,
		DisplayName:           user.DisplayName,
		UserType:              user.UserType,
		InitialAccountBalance: user.InitialAccountBalance,
		AccountBalance:        user.AccountBalance,
		PersonalEmoji:         user.PersonalEmoji,
		Description:           user.Description,
		Email:                 user.Email,
		APIKey:                user.APIKey,
		MustChangePassword:    user.MustChangePassword,
		CreatedAt:             user.CreatedAt,
		UpdatedAt:             user.UpdatedAt,
	}
	user.PersonalLinks().ApplyTo(target)
	return target
}

// FromUser copies the authenticated private projection of the user model.
func (p *PrivateProfile) FromUser(user *User) *PrivateProfile {
	return copyPrivateProfile(p, user)
}

// ApplyUpdate mutates the user with the fields supported by UserUpdateRequest.
func (u *User) ApplyUpdate(req UserUpdateRequest) {
	if u == nil {
		return
	}

	u.DisplayName = req.DisplayName
	u.Description = req.Description
	u.PersonalEmoji = req.PersonalEmoji
	req.PersonalLinks().ApplyTo(u)
}

// PersonalLinks returns the personal-link fields currently stored on the user.
func (u *User) PersonalLinks() PersonalLinks {
	if u == nil {
		return PersonalLinks{}
	}

	return PersonalLinks{
		PersonalLink1: u.PersonalLink1,
		PersonalLink2: u.PersonalLink2,
		PersonalLink3: u.PersonalLink3,
		PersonalLink4: u.PersonalLink4,
	}
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

// NewUser builds the default domain user entity for the request.
func (r UserCreateRequest) NewUser() *User {
	return &User{
		Username:              r.Username,
		DisplayName:           r.DisplayName,
		Email:                 r.Email,
		UserType:              r.UserType,
		InitialAccountBalance: 0,
		AccountBalance:        0,
		MustChangePassword:    true,
	}
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

// PersonalLinks returns the request's personal-link values in the domain storage shape.
func (r UserUpdateRequest) PersonalLinks() PersonalLinks {
	return PersonalLinks{
		PersonalLink1: r.PersonalLink1,
		PersonalLink2: r.PersonalLink2,
		PersonalLink3: r.PersonalLink3,
		PersonalLink4: r.PersonalLink4,
	}
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

// Values returns the links in storage order for bulk validation or sanitization.
func (p PersonalLinks) Values() []string {
	return []string{
		p.PersonalLink1,
		p.PersonalLink2,
		p.PersonalLink3,
		p.PersonalLink4,
	}
}

// NewPersonalLinks rebuilds personal links from an ordered value slice.
func NewPersonalLinks(values []string) PersonalLinks {
	if len(values) > personalLinkFieldCount {
		values = values[:personalLinkFieldCount]
	}

	links := PersonalLinks{}
	if len(values) > 0 {
		links.PersonalLink1 = values[0]
	}
	if len(values) > 1 {
		links.PersonalLink2 = values[1]
	}
	if len(values) > 2 {
		links.PersonalLink3 = values[2]
	}
	if len(values) > 3 {
		links.PersonalLink4 = values[3]
	}
	return links
}

// ApplyPersonalLinks updates the mutable personal-link fields on the user.
func (u *User) ApplyPersonalLinks(p PersonalLinks) {
	if u == nil {
		return
	}

	u.PersonalLink1 = p.PersonalLink1
	u.PersonalLink2 = p.PersonalLink2
	u.PersonalLink3 = p.PersonalLink3
	u.PersonalLink4 = p.PersonalLink4
}

// ApplyPersonalLinks updates the public personal-link fields.
func (u *PublicUser) ApplyPersonalLinks(p PersonalLinks) {
	if u == nil {
		return
	}

	u.PersonalLink1 = p.PersonalLink1
	u.PersonalLink2 = p.PersonalLink2
	u.PersonalLink3 = p.PersonalLink3
	u.PersonalLink4 = p.PersonalLink4
}

// ApplyPersonalLinks updates the private-profile personal-link fields.
func (p *PrivateProfile) ApplyPersonalLinks(links PersonalLinks) {
	if p == nil {
		return
	}

	p.PersonalLink1 = links.PersonalLink1
	p.PersonalLink2 = links.PersonalLink2
	p.PersonalLink3 = links.PersonalLink3
	p.PersonalLink4 = links.PersonalLink4
}

// ApplyTo updates any compatible personal-links target.
func (p PersonalLinks) ApplyTo(target PersonalLinksTarget) {
	if target == nil {
		return
	}
	target.ApplyPersonalLinks(p)
}

// Credentials represents the sensitive authentication fields associated with a user.
type Credentials struct {
	PasswordHash       string
	MustChangePassword bool
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
