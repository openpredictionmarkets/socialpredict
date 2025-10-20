package users

import (
	"context"
)

// Repository defines the interface for user data access
type Repository interface {
	GetByUsername(ctx context.Context, username string) (*User, error)
	UpdateBalance(ctx context.Context, username string, newBalance int64) error
	Create(ctx context.Context, user *User) error
	Update(ctx context.Context, user *User) error
	Delete(ctx context.Context, username string) error
	List(ctx context.Context, filters ListFilters) ([]*User, error)
}

// ListFilters represents filters for listing users
type ListFilters struct {
	UserType string
	Limit    int
	Offset   int
}

// Service implements the core user business logic
type Service struct {
	repo Repository
}

// NewService creates a new users service
func NewService(repo Repository) *Service {
	return &Service{repo: repo}
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
