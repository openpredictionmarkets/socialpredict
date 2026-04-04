package users

import (
	"context"
	"errors"
	"fmt"
	"sort"

	analytics "socialpredict/internal/domain/analytics"

	"golang.org/x/crypto/bcrypt"
)

// ServiceInterface defines the behavior required by HTTP handlers and other consumers.
type ServiceInterface interface {
	GetPublicUser(ctx context.Context, username string) (*PublicUser, error)
	GetUser(ctx context.Context, username string) (*User, error)
	GetPrivateProfile(ctx context.Context, username string) (*PrivateProfile, error)
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

// AnalyticsService exposes the computations required from the analytics domain.
type AnalyticsService interface {
	ComputeUserFinancials(ctx context.Context, req analytics.FinancialSnapshotRequest) (*analytics.FinancialSnapshot, error)
}

// Service implements the core user business logic
type Service struct {
	repo      Repository
	analytics AnalyticsService
	sanitizer Sanitizer
}

type profileMutation func(*User) error
type profileFieldSpec struct {
	validate func(string) error
	sanitize func(*Service, string) (string, error)
	apply    func(*User, string)
}

type financialSnapshotField struct {
	key     string
	extract func(*analytics.FinancialSnapshot) int64
}

type passwordChangeValidator func(*Service, string, string, string) error

var profileFieldSpecs = map[string]profileFieldSpec{
	"description": {
		validate: validateMaximumLength("description", 2000),
		sanitize: (*Service).sanitizeDescription,
		apply: func(user *User, value string) {
			user.Description = value
		},
	},
	"display_name": {
		validate: validateRequiredLength("display name", 1, 50),
		sanitize: (*Service).sanitizeDisplayName,
		apply: func(user *User, value string) {
			user.DisplayName = value
		},
	},
	"emoji": {
		validate: validateRequiredLength("emoji", 1, 20),
		sanitize: (*Service).sanitizeEmoji,
		apply: func(user *User, value string) {
			user.PersonalEmoji = value
		},
	},
}

var financialSnapshotFields = []financialSnapshotField{
	{key: "accountBalance", extract: func(snapshot *analytics.FinancialSnapshot) int64 { return snapshot.AccountBalance }},
	{key: "maximumDebtAllowed", extract: func(snapshot *analytics.FinancialSnapshot) int64 { return snapshot.MaximumDebtAllowed }},
	{key: "amountInPlay", extract: func(snapshot *analytics.FinancialSnapshot) int64 { return snapshot.AmountInPlay }},
	{key: "amountBorrowed", extract: func(snapshot *analytics.FinancialSnapshot) int64 { return snapshot.AmountBorrowed }},
	{key: "retainedEarnings", extract: func(snapshot *analytics.FinancialSnapshot) int64 { return snapshot.RetainedEarnings }},
	{key: "equity", extract: func(snapshot *analytics.FinancialSnapshot) int64 { return snapshot.Equity }},
	{key: "tradingProfits", extract: func(snapshot *analytics.FinancialSnapshot) int64 { return snapshot.TradingProfits }},
	{key: "workProfits", extract: func(snapshot *analytics.FinancialSnapshot) int64 { return snapshot.WorkProfits }},
	{key: "totalProfits", extract: func(snapshot *analytics.FinancialSnapshot) int64 { return snapshot.TotalProfits }},
	{key: "amountInPlayActive", extract: func(snapshot *analytics.FinancialSnapshot) int64 { return snapshot.AmountInPlayActive }},
	{key: "totalSpent", extract: func(snapshot *analytics.FinancialSnapshot) int64 { return snapshot.TotalSpent }},
	{key: "totalSpentInPlay", extract: func(snapshot *analytics.FinancialSnapshot) int64 { return snapshot.TotalSpentInPlay }},
	{key: "realizedProfits", extract: func(snapshot *analytics.FinancialSnapshot) int64 { return snapshot.RealizedProfits }},
	{key: "potentialProfits", extract: func(snapshot *analytics.FinancialSnapshot) int64 { return snapshot.PotentialProfits }},
	{key: "realizedValue", extract: func(snapshot *analytics.FinancialSnapshot) int64 { return snapshot.RealizedValue }},
	{key: "potentialValue", extract: func(snapshot *analytics.FinancialSnapshot) int64 { return snapshot.PotentialValue }},
}

var passwordChangeValidators = []passwordChangeValidator{
	func(_ *Service, username, _, _ string) error { return validateUsername(username) },
	func(_ *Service, _, currentPassword, _ string) error {
		if currentPassword == "" {
			return fmt.Errorf("current password is required")
		}
		return nil
	},
	func(_ *Service, _, _, newPassword string) error {
		if newPassword == "" {
			return fmt.Errorf("new password is required")
		}
		return nil
	},
	func(s *Service, _, _, _ string) error {
		return s.requireSanitizer()
	},
}

func validateUsername(username string) error {
	if username == "" {
		return ErrInvalidUserData
	}
	return nil
}

func validateUserID(userID int64) error {
	if userID <= 0 {
		return ErrInvalidUserData
	}
	return nil
}

// NewService creates a new users service
func NewService(repo Repository, analyticsSvc AnalyticsService, sanitizer Sanitizer) *Service {
	service := &Service{
		repo:      repo,
		analytics: analyticsSvc,
		sanitizer: sanitizer,
	}
	return service
}

// ValidateUserExists checks if a user exists
func (s *Service) ValidateUserExists(ctx context.Context, username string) error {
	if err := validateUsername(username); err != nil {
		return err
	}
	_, err := s.requireUser(ctx, username)
	return err
}

// ValidateUserBalance validates if a user has sufficient balance for an operation
func (s *Service) ValidateUserBalance(ctx context.Context, username string, requiredAmount int64, maxDebt int64) error {
	user, err := s.requireUser(ctx, username)
	if err != nil {
		return err
	}

	// Check if user would exceed maximum debt
	if user.AccountBalance-requiredAmount < -maxDebt {
		return ErrInsufficientBalance
	}

	return nil
}

// DeductBalance deducts an amount from a user's balance
func (s *Service) DeductBalance(ctx context.Context, username string, amount int64) error {
	return s.ApplyTransaction(ctx, username, amount, TransactionBuy)
}

// GetUser retrieves a user by username
func (s *Service) GetUser(ctx context.Context, username string) (*User, error) {
	if err := validateUsername(username); err != nil {
		return nil, err
	}
	return s.requireUser(ctx, username)
}

// GetPublicUser retrieves the public view of a user
func (s *Service) GetPublicUser(ctx context.Context, username string) (*PublicUser, error) {
	user, err := s.GetUser(ctx, username)
	if err != nil {
		return nil, err
	}

	return user.ToPublicUser(), nil
}

// CreateUser creates a new user
func (s *Service) CreateUser(ctx context.Context, req UserCreateRequest) (*User, error) {
	if err := validateUsername(req.Username); err != nil {
		return nil, err
	}

	if _, err := s.repo.GetByUsername(ctx, req.Username); err == nil {
		return nil, ErrUserAlreadyExists
	} else if !errors.Is(err, ErrUserNotFound) {
		return nil, err
	}

	user := req.NewUser()

	if err := s.repo.Create(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

// UpdateUser updates user information
func (s *Service) UpdateUser(ctx context.Context, username string, req UserUpdateRequest) (*User, error) {
	return s.updateUserProfile(ctx, username, func(user *User) error {
		user.ApplyUpdate(req)
		return nil
	})
}

// ListUsers returns a list of users with filters
func (s *Service) ListUsers(ctx context.Context, filters ListFilters) ([]*User, error) {
	return s.repo.List(ctx, filters)
}

// DeleteUser removes a user
func (s *Service) DeleteUser(ctx context.Context, username string) error {
	if err := s.ValidateUserExists(ctx, username); err != nil {
		return err
	}

	return s.repo.Delete(ctx, username)
}

// ApplyTransaction adjusts the user's account balance based on the supplied transaction type.
func (s *Service) ApplyTransaction(ctx context.Context, username string, amount int64, transactionType string) error {
	return s.updateUserBalance(ctx, username, func(user *User) (int64, error) {
		return applyTransactionBalance(user.AccountBalance, amount, transactionType)
	})
}

// GetUserCredit returns the available credit for a user based on their balance and the maximum debt limit.
func (s *Service) GetUserCredit(ctx context.Context, username string, maximumDebtAllowed int64) (int64, error) {
	if err := validateUsername(username); err != nil {
		return 0, err
	}

	user, err := s.repo.GetByUsername(ctx, username)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
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
	if err := validateUserID(userID); err != nil {
		return nil, err
	}
	return s.repo.ListUserMarkets(ctx, userID)
}

// GetUserFinancials returns the user's comprehensive financial snapshot.
func (s *Service) GetUserFinancials(ctx context.Context, username string) (map[string]int64, error) {
	if s.analytics == nil {
		return nil, ErrInvalidUserData
	}

	user, err := s.requireUser(ctx, username)
	if err != nil {
		return nil, err
	}

	snapshot, err := s.analytics.ComputeUserFinancials(ctx, analytics.FinancialSnapshotRequest{
		Username:       username,
		AccountBalance: user.AccountBalance,
	})
	if err != nil {
		return nil, err
	}
	if snapshot == nil {
		return map[string]int64{}, nil
	}

	return financialSnapshotToMap(snapshot), nil
}

// UpdateDescription sanitizes and updates a user's description.
func (s *Service) UpdateDescription(ctx context.Context, username, description string) (*User, error) {
	return s.updateProfileField(ctx, username, description, profileFieldSpecs["description"])
}

// UpdateDisplayName sanitizes and updates a user's display name.
func (s *Service) UpdateDisplayName(ctx context.Context, username, displayName string) (*User, error) {
	return s.updateProfileField(ctx, username, displayName, profileFieldSpecs["display_name"])
}

// UpdateEmoji sanitizes and updates a user's personal emoji.
func (s *Service) UpdateEmoji(ctx context.Context, username, emoji string) (*User, error) {
	return s.updateProfileField(ctx, username, emoji, profileFieldSpecs["emoji"])
}

// UpdatePersonalLinks sanitizes and updates a user's personal links.
func (s *Service) UpdatePersonalLinks(ctx context.Context, username string, links PersonalLinks) (*User, error) {
	sanitized, err := s.sanitizePersonalLinks(links)
	if err != nil {
		return nil, err
	}

	return s.updateUserProfile(ctx, username, func(user *User) error {
		sanitized.ApplyTo(user)
		return nil
	})
}

func (s *Service) sanitizePersonalLinks(links PersonalLinks) (PersonalLinks, error) {
	if err := s.requireSanitizer(); err != nil {
		return PersonalLinks{}, err
	}

	values := links.Values()
	for _, link := range values {
		if len(link) > 200 {
			return PersonalLinks{}, fmt.Errorf("personal link exceeds maximum length of 200 characters")
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
			return PersonalLinks{}, fmt.Errorf("invalid personal link: %w", err)
		}
		sanitized[i] = clean
	}
	return NewPersonalLinks(sanitized), nil
}

func financialSnapshotToMap(snapshot *analytics.FinancialSnapshot) map[string]int64 {
	values := make(map[string]int64, len(financialSnapshotFields))
	for _, field := range financialSnapshotFields {
		values[field.key] = field.extract(snapshot)
	}
	return values
}

// GetPrivateProfile returns the combined private and public user information for the specified username.
func (s *Service) GetPrivateProfile(ctx context.Context, username string) (*PrivateProfile, error) {
	user, err := s.GetUser(ctx, username)
	if err != nil {
		return nil, err
	}

	return user.ToPrivateProfile(), nil
}

const passwordHashCost = 14

// PasswordHashCost exposes the bcrypt cost used for hashing user passwords.
func PasswordHashCost() int {
	return passwordHashCost
}

func (s *Service) validatePasswordChangeInputs(username, currentPassword, newPassword string) error {
	for _, validate := range passwordChangeValidators {
		if err := validate(s, username, currentPassword, newPassword); err != nil {
			return err
		}
	}
	return nil
}

func (s *Service) requireUser(ctx context.Context, username string) (*User, error) {
	user, err := s.repo.GetByUsername(ctx, username)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return user, nil
}

func (s *Service) updateUserBalance(ctx context.Context, username string, compute func(*User) (int64, error)) error {
	user, err := s.requireUser(ctx, username)
	if err != nil {
		return err
	}

	newBalance, err := compute(user)
	if err != nil {
		return err
	}
	return s.repo.UpdateBalance(ctx, username, newBalance)
}

func (s *Service) updateUserProfile(ctx context.Context, username string, mutate profileMutation) (*User, error) {
	user, err := s.requireUser(ctx, username)
	if err != nil {
		return nil, err
	}

	if err := mutate(user); err != nil {
		return nil, err
	}
	if err := s.repo.Update(ctx, user); err != nil {
		return nil, err
	}
	return user, nil
}

func (s *Service) updateProfileField(ctx context.Context, username, value string, spec profileFieldSpec) (*User, error) {
	if spec.validate != nil {
		if err := spec.validate(value); err != nil {
			return nil, err
		}
	}

	return s.updateUserProfile(ctx, username, func(user *User) error {
		sanitized, err := spec.sanitize(s, value)
		if err != nil {
			return err
		}
		spec.apply(user, sanitized)
		return nil
	})
}

func (s *Service) requireSanitizer() error {
	if s.sanitizer == nil {
		return ErrInvalidUserData
	}
	return nil
}

func (s *Service) sanitizeDescription(description string) (string, error) {
	if err := s.requireSanitizer(); err != nil {
		return "", err
	}

	sanitized, err := s.sanitizer.SanitizeDescription(description)
	if err != nil {
		return "", fmt.Errorf("invalid description: %w", err)
	}
	return sanitized, nil
}

func (s *Service) sanitizeDisplayName(displayName string) (string, error) {
	if err := s.requireSanitizer(); err != nil {
		return "", err
	}

	sanitized, err := s.sanitizer.SanitizeDisplayName(displayName)
	if err != nil {
		return "", fmt.Errorf("invalid display name: %w", err)
	}
	return sanitized, nil
}

func (s *Service) sanitizeEmoji(emoji string) (string, error) {
	if err := s.requireSanitizer(); err != nil {
		return "", err
	}

	sanitized, err := s.sanitizer.SanitizeEmoji(emoji)
	if err != nil {
		return "", fmt.Errorf("invalid emoji: %w", err)
	}
	return sanitized, nil
}

func validateMaximumLength(fieldName string, max int) func(string) error {
	return func(value string) error {
		if len(value) > max {
			return fmt.Errorf("%s exceeds maximum length of %d characters", fieldName, max)
		}
		return nil
	}
}

func validateRequiredLength(fieldName string, min int, max int) func(string) error {
	return func(value string) error {
		if len(value) < min {
			if fieldName == "emoji" {
				return fmt.Errorf("%s cannot be blank", fieldName)
			}
			return fmt.Errorf("%s must be between %d and %d characters", fieldName, min, max)
		}
		if len(value) > max {
			return fmt.Errorf("%s exceeds maximum length of %d characters", fieldName, max)
		}
		return nil
	}
}

func (s *Service) getCredentials(ctx context.Context, username string) (*Credentials, error) {
	if err := validateUsername(username); err != nil {
		return nil, err
	}
	return s.repo.GetCredentials(ctx, username)
}

func (s *Service) verifyCurrentPassword(creds *Credentials, currentPassword string) error {
	if err := bcrypt.CompareHashAndPassword([]byte(creds.PasswordHash), []byte(currentPassword)); err != nil {
		return ErrInvalidCredentials
	}
	return nil
}

func (s *Service) sanitizeNewPassword(newPassword string) (string, error) {
	if err := s.requireSanitizer(); err != nil {
		return "", err
	}

	sanitized, err := s.sanitizer.SanitizePassword(newPassword)
	if err != nil {
		return "", fmt.Errorf("new password does not meet security requirements: %w", err)
	}
	return sanitized, nil
}

func ensurePasswordChanged(currentHash string, newPassword string) error {
	if err := bcrypt.CompareHashAndPassword([]byte(currentHash), []byte(newPassword)); err == nil {
		return fmt.Errorf("new password must differ from the current password")
	}
	return nil
}

func hashPassword(password string) (string, error) {
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), passwordHashCost)
	if err != nil {
		return "", fmt.Errorf("failed to hash new password: %w", err)
	}
	return string(hashed), nil
}

// ChangePassword validates credentials and persists a new hashed password.
func (s *Service) ChangePassword(ctx context.Context, username, currentPassword, newPassword string) error {
	if err := s.validatePasswordChangeInputs(username, currentPassword, newPassword); err != nil {
		return err
	}

	creds, err := s.getCredentials(ctx, username)
	if err != nil {
		return err
	}

	if err := s.verifyCurrentPassword(creds, currentPassword); err != nil {
		return err
	}

	sanitized, err := s.sanitizeNewPassword(newPassword)
	if err != nil {
		return err
	}

	if err := ensurePasswordChanged(creds.PasswordHash, sanitized); err != nil {
		return err
	}

	hashed, err := hashPassword(sanitized)
	if err != nil {
		return err
	}

	return s.repo.UpdatePassword(ctx, username, hashed, false)
}

var _ ServiceInterface = (*Service)(nil)
