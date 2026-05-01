package users

import (
	"context"
	"fmt"
	"math/rand"

	"github.com/brianvoe/gofakeit"
	"github.com/google/uuid"
)

// AdminManagedUserCreateRequest contains admin-owned defaults for creating a regular user.
type AdminManagedUserCreateRequest struct {
	Username              string
	InitialAccountBalance int64
}

// AdminManagedUserCreateResult contains the created user fields returned to the admin.
type AdminManagedUserCreateResult struct {
	Username string
	Password string
	UserType string
}

// CreateAdminManagedUser creates a regular user with generated private identity fields.
func (s *Service) CreateAdminManagedUser(ctx context.Context, req AdminManagedUserCreateRequest) (*AdminManagedUserCreateResult, error) {
	if err := validateUsername(req.Username); err != nil {
		return nil, err
	}

	uniqueness, err := s.userUniquenessRepository()
	if err != nil {
		return nil, err
	}
	if exists, err := uniqueness.UsernameExists(ctx, req.Username); err != nil {
		return nil, err
	} else if exists {
		return nil, ErrUserAlreadyExists
	}

	displayName, err := uniqueDisplayName(ctx, uniqueness)
	if err != nil {
		return nil, err
	}
	email, err := uniqueEmail(ctx, uniqueness)
	if err != nil {
		return nil, err
	}
	apiKey, err := uniqueAPIKey(ctx, uniqueness)
	if err != nil {
		return nil, err
	}

	password := gofakeit.Password(true, true, true, false, false, 12)
	passwordHash, err := hashPassword(password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	user := &User{
		Username:              req.Username,
		DisplayName:           displayName,
		Email:                 email,
		APIKey:                apiKey,
		PasswordHash:          passwordHash,
		UserType:              "REGULAR",
		InitialAccountBalance: req.InitialAccountBalance,
		AccountBalance:        req.InitialAccountBalance,
		PersonalEmoji:         randomEmoji(),
		MustChangePassword:    true,
	}

	if exists, err := uniqueness.AnyUserIdentityExists(ctx, user.Username, user.DisplayName, user.Email, user.APIKey); err != nil {
		return nil, err
	} else if exists {
		return nil, ErrUserAlreadyExists
	}

	writer, err := s.userWriter()
	if err != nil {
		return nil, err
	}
	if err := writer.Create(ctx, user); err != nil {
		return nil, err
	}

	return &AdminManagedUserCreateResult{
		Username: user.Username,
		Password: password,
		UserType: user.UserType,
	}, nil
}

func uniqueDisplayName(ctx context.Context, repo UserUniquenessRepository) (string, error) {
	for {
		name := gofakeit.Name()
		exists, err := repo.DisplayNameExists(ctx, name)
		if err != nil {
			return "", err
		}
		if !exists {
			return name, nil
		}
	}
}

func uniqueEmail(ctx context.Context, repo UserUniquenessRepository) (string, error) {
	for {
		email := gofakeit.Email()
		exists, err := repo.EmailExists(ctx, email)
		if err != nil {
			return "", err
		}
		if !exists {
			return email, nil
		}
	}
}

func uniqueAPIKey(ctx context.Context, repo UserUniquenessRepository) (string, error) {
	for {
		apiKey := uuid.NewString()
		exists, err := repo.APIKeyExists(ctx, apiKey)
		if err != nil {
			return "", err
		}
		if !exists {
			return apiKey, nil
		}
	}
}

func randomEmoji() string {
	emojis := []string{"😀", "😃", "😄", "😁", "😆"}
	return emojis[rand.Intn(len(emojis))]
}
