package users

import (
	"context"
	"fmt"
	"sort"

	"socialpredict/setup"

	"golang.org/x/crypto/bcrypt"
)

// ServiceInterface defines the behavior required by HTTP handlers and other consumers.
type ServiceInterface interface {
	GetPublicUser(ctx context.Context, username string) (*PublicUser, error)
	ApplyTransaction(ctx context.Context, username string, amount int64, transactionType string) error
	GetUserCredit(ctx context.Context, username string, maximumDebtAllowed int64) (int64, error)
	GetUserPortfolio(ctx context.Context, username string) (*Portfolio, error)
	GetUserFinancials(ctx context.Context, username string) (map[string]int64, error)
	ListUserMarkets(ctx context.Context, userID int64) ([]*UserMarket, error)
	UpdateDescription(ctx context.Context, username, description string) (*User, error)
	UpdateDisplayName(ctx context.Context, username, displayName string) (*User, error)
	UpdateEmoji(ctx context.Context, username, emoji string) (*User, error)
	UpdatePersonalLinks(ctx context.Context, username string, links PersonalLinks) (*User, error)
	ChangePassword(ctx context.Context, username, currentPassword, newPassword string) error
}

// Repository defines the interface for user data access
type Repository interface {
	GetByUsername(ctx context.Context, username string) (*User, error)
	UpdateBalance(ctx context.Context, username string, newBalance int64) error
	Create(ctx context.Context, user *User) error
	Update(ctx context.Context, user *User) error
	Delete(ctx context.Context, username string) error
	List(ctx context.Context, filters ListFilters) ([]*User, error)
	ListUserBets(ctx context.Context, username string) ([]*UserBet, error)
	GetMarketQuestion(ctx context.Context, marketID uint) (string, error)
	GetUserPositionInMarket(ctx context.Context, marketID int64, username string) (*MarketUserPosition, error)
	ComputeUserFinancials(ctx context.Context, username string, accountBalance int64, econ *setup.EconomicConfig) (map[string]int64, error)
	ListUserMarkets(ctx context.Context, userID int64) ([]*UserMarket, error)
	GetCredentials(ctx context.Context, username string) (*Credentials, error)
	UpdatePassword(ctx context.Context, username string, hashedPassword string, mustChange bool) error
}

// ListFilters represents filters for listing users
type ListFilters struct {
	UserType string
	Limit    int
	Offset   int
}

// Sanitizer defines the behavior needed to sanitize user profile inputs.
type Sanitizer interface {
	SanitizeDescription(string) (string, error)
	SanitizeDisplayName(string) (string, error)
	SanitizeEmoji(string) (string, error)
	SanitizePersonalLink(string) (string, error)
	SanitizePassword(string) (string, error)
}

// Service implements the core user business logic
type Service struct {
	repo      Repository
	config    *setup.EconomicConfig
	sanitizer Sanitizer
}

// NewService creates a new users service
func NewService(repo Repository, config *setup.EconomicConfig, sanitizer Sanitizer) *Service {
	return &Service{
		repo:      repo,
		config:    config,
		sanitizer: sanitizer,
	}
}

// ValidateUserExists checks if a user exists
func (s *Service) ValidateUserExists(ctx context.Context, username string) error {
	_, err := s.repo.GetByUsername(ctx, username)
	if err != nil {
		return ErrUserNotFound
	}
	return nil
}

// ValidateUserBalance validates if a user has sufficient balance for an operation
func (s *Service) ValidateUserBalance(ctx context.Context, username string, requiredAmount float64, maxDebt float64) error {
	user, err := s.repo.GetByUsername(ctx, username)
	if err != nil {
		return ErrUserNotFound
	}

	// Convert float64 amounts to int64 (assuming cents)
	requiredCents := int64(requiredAmount * 100)
	maxDebtCents := int64(maxDebt * 100)

	// Check if user would exceed maximum debt
	if user.AccountBalance-requiredCents < -maxDebtCents {
		return ErrInsufficientBalance
	}

	return nil
}

// DeductBalance deducts an amount from a user's balance
func (s *Service) DeductBalance(ctx context.Context, username string, amount float64) error {
	user, err := s.repo.GetByUsername(ctx, username)
	if err != nil {
		return ErrUserNotFound
	}

	// Convert float64 amount to int64 (assuming cents)
	amountCents := int64(amount * 100)

	newBalance := user.AccountBalance - amountCents
	return s.repo.UpdateBalance(ctx, username, newBalance)
}

// GetUser retrieves a user by username
func (s *Service) GetUser(ctx context.Context, username string) (*User, error) {
	return s.repo.GetByUsername(ctx, username)
}

// GetPublicUser retrieves the public view of a user
func (s *Service) GetPublicUser(ctx context.Context, username string) (*PublicUser, error) {
	user, err := s.repo.GetByUsername(ctx, username)
	if err != nil {
		return nil, ErrUserNotFound
	}

	return &PublicUser{
		ID:                    user.ID,
		Username:              user.Username,
		DisplayName:           user.DisplayName,
		UserType:              user.UserType,
		InitialAccountBalance: user.InitialAccountBalance,
		AccountBalance:        user.AccountBalance,
		PersonalEmoji:         user.PersonalEmoji,
		Description:           user.Description,
		PersonalLink1:         user.PersonalLink1,
		PersonalLink2:         user.PersonalLink2,
		PersonalLink3:         user.PersonalLink3,
		PersonalLink4:         user.PersonalLink4,
	}, nil
}

// CreateUser creates a new user
func (s *Service) CreateUser(ctx context.Context, req UserCreateRequest) (*User, error) {
	// Check if user already exists
	if _, err := s.repo.GetByUsername(ctx, req.Username); err == nil {
		return nil, ErrUserAlreadyExists
	}

	user := &User{
		Username:              req.Username,
		DisplayName:           req.DisplayName,
		Email:                 req.Email,
		UserType:              req.UserType,
		InitialAccountBalance: 0,
		AccountBalance:        0,
		MustChangePassword:    true,
	}

	if err := s.repo.Create(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

// UpdateUser updates user information
func (s *Service) UpdateUser(ctx context.Context, username string, req UserUpdateRequest) (*User, error) {
	user, err := s.repo.GetByUsername(ctx, username)
	if err != nil {
		return nil, ErrUserNotFound
	}

	// Update fields
	user.DisplayName = req.DisplayName
	user.Description = req.Description
	user.PersonalEmoji = req.PersonalEmoji
	user.PersonalLink1 = req.PersonalLink1
	user.PersonalLink2 = req.PersonalLink2
	user.PersonalLink3 = req.PersonalLink3
	user.PersonalLink4 = req.PersonalLink4

	if err := s.repo.Update(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

// ListUsers returns a list of users with filters
func (s *Service) ListUsers(ctx context.Context, filters ListFilters) ([]*User, error) {
	return s.repo.List(ctx, filters)
}

// DeleteUser removes a user
func (s *Service) DeleteUser(ctx context.Context, username string) error {
	// Check if user exists
	if err := s.ValidateUserExists(ctx, username); err != nil {
		return err
	}

	return s.repo.Delete(ctx, username)
}

// ApplyTransaction adjusts the user's account balance based on the supplied transaction type.
func (s *Service) ApplyTransaction(ctx context.Context, username string, amount int64, transactionType string) error {
	user, err := s.repo.GetByUsername(ctx, username)
	if err != nil {
		return ErrUserNotFound
	}

	newBalance := user.AccountBalance
	switch transactionType {
	case TransactionWin, TransactionRefund, TransactionSale:
		newBalance += amount
	case TransactionBuy, TransactionFee:
		newBalance -= amount
	default:
		return ErrInvalidTransactionType
	}

	return s.repo.UpdateBalance(ctx, username, newBalance)
}

// GetUserCredit returns the available credit for a user based on their balance and the maximum debt limit.
func (s *Service) GetUserCredit(ctx context.Context, username string, maximumDebtAllowed int64) (int64, error) {
	user, err := s.repo.GetByUsername(ctx, username)
	if err != nil {
		if err == ErrUserNotFound {
			return maximumDebtAllowed, nil
		}
		return 0, err
	}

	return maximumDebtAllowed + user.AccountBalance, nil
}

// GetUserPortfolio returns the user's portfolio across markets.
func (s *Service) GetUserPortfolio(ctx context.Context, username string) (*Portfolio, error) {
	bets, err := s.repo.ListUserBets(ctx, username)
	if err != nil {
		return nil, err
	}

	marketMap := make(map[uint]*PortfolioItem)
	for _, bet := range bets {
		item, exists := marketMap[bet.MarketID]
		if !exists {
			item = &PortfolioItem{
				MarketID:      bet.MarketID,
				LastBetPlaced: bet.PlacedAt,
			}
			marketMap[bet.MarketID] = item
		}
		if bet.PlacedAt.After(item.LastBetPlaced) {
			item.LastBetPlaced = bet.PlacedAt
		}
	}

	var items []PortfolioItem
	var totalShares int64
	for marketID, item := range marketMap {
		position, err := s.repo.GetUserPositionInMarket(ctx, int64(marketID), username)
		if err != nil {
			return nil, err
		}

		title, err := s.repo.GetMarketQuestion(ctx, marketID)
		if err != nil {
			return nil, err
		}

		item.YesSharesOwned = position.YesSharesOwned
		item.NoSharesOwned = position.NoSharesOwned
		item.QuestionTitle = title
		totalShares += position.YesSharesOwned + position.NoSharesOwned

		items = append(items, *item)
	}

	sort.Slice(items, func(i, j int) bool {
		return items[i].LastBetPlaced.After(items[j].LastBetPlaced)
	})

	return &Portfolio{
		Items:            items,
		TotalSharesOwned: totalShares,
	}, nil
}

// ListUserMarkets returns markets the specified user has participated in.
func (s *Service) ListUserMarkets(ctx context.Context, userID int64) ([]*UserMarket, error) {
	if userID <= 0 {
		return nil, ErrInvalidUserData
	}
	return s.repo.ListUserMarkets(ctx, userID)
}

// GetUserFinancials returns the user's comprehensive financial snapshot.
func (s *Service) GetUserFinancials(ctx context.Context, username string) (map[string]int64, error) {
	if s.config == nil {
		return nil, ErrInvalidUserData
	}

	user, err := s.repo.GetByUsername(ctx, username)
	if err != nil {
		return nil, ErrUserNotFound
	}

	return s.repo.ComputeUserFinancials(ctx, username, user.AccountBalance, s.config)
}

// UpdateDescription sanitizes and updates a user's description.
func (s *Service) UpdateDescription(ctx context.Context, username, description string) (*User, error) {
	if len(description) > 2000 {
		return nil, fmt.Errorf("description exceeds maximum length of 2000 characters")
	}

	user, err := s.repo.GetByUsername(ctx, username)
	if err != nil {
		return nil, err
	}

	if s.sanitizer == nil {
		return nil, ErrInvalidUserData
	}

	sanitized, err := s.sanitizer.SanitizeDescription(description)
	if err != nil {
		return nil, fmt.Errorf("invalid description: %w", err)
	}

	user.Description = sanitized
	if err := s.repo.Update(ctx, user); err != nil {
		return nil, err
	}
	return user, nil
}

// UpdateDisplayName sanitizes and updates a user's display name.
func (s *Service) UpdateDisplayName(ctx context.Context, username, displayName string) (*User, error) {
	if len(displayName) < 1 || len(displayName) > 50 {
		return nil, fmt.Errorf("display name must be between 1 and 50 characters")
	}

	user, err := s.repo.GetByUsername(ctx, username)
	if err != nil {
		return nil, err
	}

	if s.sanitizer == nil {
		return nil, ErrInvalidUserData
	}

	sanitized, err := s.sanitizer.SanitizeDisplayName(displayName)
	if err != nil {
		return nil, fmt.Errorf("invalid display name: %w", err)
	}

	user.DisplayName = sanitized
	if err := s.repo.Update(ctx, user); err != nil {
		return nil, err
	}
	return user, nil
}

// UpdateEmoji sanitizes and updates a user's personal emoji.
func (s *Service) UpdateEmoji(ctx context.Context, username, emoji string) (*User, error) {
	if emoji == "" {
		return nil, fmt.Errorf("emoji cannot be blank")
	}
	if len(emoji) > 20 {
		return nil, fmt.Errorf("emoji exceeds maximum length of 20 characters")
	}

	user, err := s.repo.GetByUsername(ctx, username)
	if err != nil {
		return nil, err
	}

	if s.sanitizer == nil {
		return nil, ErrInvalidUserData
	}

	sanitized, err := s.sanitizer.SanitizeEmoji(emoji)
	if err != nil {
		return nil, fmt.Errorf("invalid emoji: %w", err)
	}

	user.PersonalEmoji = sanitized
	if err := s.repo.Update(ctx, user); err != nil {
		return nil, err
	}
	return user, nil
}

// UpdatePersonalLinks sanitizes and updates a user's personal links.
func (s *Service) UpdatePersonalLinks(ctx context.Context, username string, links PersonalLinks) (*User, error) {
	if s.sanitizer == nil {
		return nil, ErrInvalidUserData
	}

	values := []string{
		links.PersonalLink1,
		links.PersonalLink2,
		links.PersonalLink3,
		links.PersonalLink4,
	}

	for _, link := range values {
		if len(link) > 200 {
			return nil, fmt.Errorf("personal link exceeds maximum length of 200 characters")
		}
	}

	sanitized := make([]string, len(values))
	for i, link := range values {
		if link == "" {
			sanitized[i] = ""
			continue
		}
		clean, err := s.sanitizer.SanitizePersonalLink(link)
		if err != nil {
			return nil, fmt.Errorf("invalid personal link: %w", err)
		}
		sanitized[i] = clean
	}

	user, err := s.repo.GetByUsername(ctx, username)
	if err != nil {
		return nil, err
	}

	user.PersonalLink1 = sanitized[0]
	user.PersonalLink2 = sanitized[1]
	user.PersonalLink3 = sanitized[2]
	user.PersonalLink4 = sanitized[3]

	if err := s.repo.Update(ctx, user); err != nil {
		return nil, err
	}
	return user, nil
}

const passwordHashCost = 14

// PasswordHashCost exposes the bcrypt cost used for hashing user passwords.
func PasswordHashCost() int {
	return passwordHashCost
}

// ChangePassword validates credentials and persists a new hashed password.
func (s *Service) ChangePassword(ctx context.Context, username, currentPassword, newPassword string) error {
	if username == "" {
		return ErrInvalidUserData
	}
	if currentPassword == "" {
		return fmt.Errorf("current password is required")
	}
	if newPassword == "" {
		return fmt.Errorf("new password is required")
	}
	if s.sanitizer == nil {
		return ErrInvalidUserData
	}

	creds, err := s.repo.GetCredentials(ctx, username)
	if err != nil {
		return err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(creds.PasswordHash), []byte(currentPassword)); err != nil {
		return ErrInvalidCredentials
	}

	sanitized, err := s.sanitizer.SanitizePassword(newPassword)
	if err != nil {
		return fmt.Errorf("new password does not meet security requirements: %w", err)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(creds.PasswordHash), []byte(sanitized)); err == nil {
		return fmt.Errorf("new password must differ from the current password")
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(sanitized), passwordHashCost)
	if err != nil {
		return fmt.Errorf("failed to hash new password: %w", err)
	}

	return s.repo.UpdatePassword(ctx, username, string(hashed), false)
}

var _ ServiceInterface = (*Service)(nil)
