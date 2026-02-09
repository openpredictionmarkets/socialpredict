package auth

import (
	"context"
	"fmt"

	dauth "socialpredict/internal/domain/auth"
	dusers "socialpredict/internal/domain/users"
	usermodels "socialpredict/internal/domain/users/models"

	"golang.org/x/crypto/bcrypt"
)

// PasswordService handles credential validation and password rotation rules.
type PasswordService struct {
	repo      CredentialRepository
	sanitizer PasswordSanitizer
}

// NewPasswordService constructs the credential/password service.
func NewPasswordService(repo CredentialRepository, sanitizer PasswordSanitizer) *PasswordService {
	return &PasswordService{
		repo:      repo,
		sanitizer: sanitizer,
	}
}

// MustChangePassword reports whether the specified user is required to change their password.
func (s *PasswordService) MustChangePassword(ctx context.Context, username string) (bool, error) {
	if s.repo == nil {
		return false, dusers.ErrInvalidUserData
	}

	creds, err := s.repo.GetCredentials(ctx, username)
	if err != nil {
		return false, err
	}
	return creds.MustChangePassword, nil
}

// ChangePassword validates credentials and persists a new hashed password.
func (s *PasswordService) ChangePassword(ctx context.Context, username, currentPassword, newPassword string) error {
	if username == "" {
		return dusers.ErrInvalidUserData
	}
	if currentPassword == "" {
		return fmt.Errorf("current password is required")
	}
	if newPassword == "" {
		return fmt.Errorf("new password is required")
	}
	if s.repo == nil || s.sanitizer == nil {
		return dusers.ErrInvalidUserData
	}

	creds, err := s.repo.GetCredentials(ctx, username)
	if err != nil {
		return err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(creds.PasswordHash), []byte(currentPassword)); err != nil {
		return dauth.ErrInvalidCredentials
	}

	sanitized, err := s.sanitizer.SanitizePassword(newPassword)
	if err != nil {
		return fmt.Errorf("new password does not meet security requirements: %w", err)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(creds.PasswordHash), []byte(sanitized)); err == nil {
		return fmt.Errorf("new password must differ from the current password")
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(sanitized), usermodels.PasswordHashCost())
	if err != nil {
		return fmt.Errorf("failed to hash new password: %w", err)
	}

	return s.repo.UpdatePassword(ctx, username, string(hashed), false)
}

var _ PasswordManager = (*PasswordService)(nil)
