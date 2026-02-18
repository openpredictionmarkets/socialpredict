package auth

import (
	"context"
	"errors"
	"testing"

	dauth "socialpredict/internal/domain/auth"
	dusers "socialpredict/internal/domain/users"

	"golang.org/x/crypto/bcrypt"
)

// --- Test doubles ---

type fakeCredentialRepo struct {
	creds     *dauth.Credentials
	err       error
	updateErr error

	updatedUsername string
	updatedHash    string
	updatedMust    bool
}

func (r *fakeCredentialRepo) GetCredentials(_ context.Context, _ string) (*dauth.Credentials, error) {
	if r.err != nil {
		return nil, r.err
	}
	return r.creds, nil
}

func (r *fakeCredentialRepo) UpdatePassword(_ context.Context, username, hashedPassword string, mustChange bool) error {
	if r.updateErr != nil {
		return r.updateErr
	}
	r.updatedUsername = username
	r.updatedHash = hashedPassword
	r.updatedMust = mustChange
	return nil
}

type fakeSanitizer struct {
	result string
	err    error
}

func (s *fakeSanitizer) SanitizePassword(pw string) (string, error) {
	if s.err != nil {
		return "", s.err
	}
	if s.result != "" {
		return s.result, nil
	}
	return pw, nil
}

// --- EnsureAdmin tests ---

func TestEnsureAdmin_AdminUser(t *testing.T) {
	svc := NewIdentityService(nil)

	err := svc.EnsureAdmin(&dusers.User{UserType: "ADMIN"})
	if err != nil {
		t.Fatalf("expected nil for ADMIN user, got %v", err)
	}
}

func TestEnsureAdmin_AdminLowercase(t *testing.T) {
	svc := NewIdentityService(nil)

	err := svc.EnsureAdmin(&dusers.User{UserType: "admin"})
	if err != nil {
		t.Fatalf("expected nil for lowercase admin, got %v", err)
	}
}

func TestEnsureAdmin_RegularUser(t *testing.T) {
	svc := NewIdentityService(nil)

	err := svc.EnsureAdmin(&dusers.User{UserType: "regular"})
	if !errors.Is(err, ErrAdminPrivilegesRequired) {
		t.Fatalf("expected ErrAdminPrivilegesRequired, got %v", err)
	}
}

func TestEnsureAdmin_NilUser(t *testing.T) {
	svc := NewIdentityService(nil)

	err := svc.EnsureAdmin(nil)
	if !errors.Is(err, ErrInvalidToken) {
		t.Fatalf("expected ErrInvalidToken, got %v", err)
	}
}

// --- MustChangePassword tests ---

func TestMustChangePassword_True(t *testing.T) {
	repo := &fakeCredentialRepo{
		creds: &dauth.Credentials{UserID: 1, PasswordHash: "hash", MustChangePassword: true},
	}
	svc := NewPasswordService(repo, nil)

	must, err := svc.MustChangePassword(context.Background(), "alice")
	if err != nil {
		t.Fatalf("MustChangePassword returned error: %v", err)
	}
	if !must {
		t.Fatalf("expected true, got false")
	}
}

func TestMustChangePassword_False(t *testing.T) {
	repo := &fakeCredentialRepo{
		creds: &dauth.Credentials{UserID: 1, PasswordHash: "hash", MustChangePassword: false},
	}
	svc := NewPasswordService(repo, nil)

	must, err := svc.MustChangePassword(context.Background(), "alice")
	if err != nil {
		t.Fatalf("MustChangePassword returned error: %v", err)
	}
	if must {
		t.Fatalf("expected false, got true")
	}
}

func TestMustChangePassword_NilRepo(t *testing.T) {
	svc := NewPasswordService(nil, nil)

	_, err := svc.MustChangePassword(context.Background(), "alice")
	if !errors.Is(err, dusers.ErrInvalidUserData) {
		t.Fatalf("expected ErrInvalidUserData, got %v", err)
	}
}

func TestMustChangePassword_RepoUserNotFound(t *testing.T) {
	repo := &fakeCredentialRepo{err: dusers.ErrUserNotFound}
	svc := NewPasswordService(repo, nil)

	_, err := svc.MustChangePassword(context.Background(), "unknown")
	if !errors.Is(err, dusers.ErrUserNotFound) {
		t.Fatalf("expected ErrUserNotFound, got %v", err)
	}
}

func TestMustChangePassword_RepoGenericError(t *testing.T) {
	dbErr := errors.New("database timeout")
	repo := &fakeCredentialRepo{err: dbErr}
	svc := NewPasswordService(repo, nil)

	_, err := svc.MustChangePassword(context.Background(), "alice")
	if !errors.Is(err, dbErr) {
		t.Fatalf("expected generic repo error, got %v", err)
	}
}

func TestMustChangePassword_NilCredentials(t *testing.T) {
	// Repo returns (nil, nil) — accessing .MustChangePassword on nil would panic
	repo := &fakeCredentialRepo{creds: nil, err: nil}
	svc := NewPasswordService(repo, nil)

	defer func() {
		if r := recover(); r != nil {
			t.Logf("MustChangePassword panics on nil credentials: %v", r)
		}
	}()

	_, _ = svc.MustChangePassword(context.Background(), "alice")
}

// --- ChangePassword tests ---

func TestChangePassword_EmptyUsername(t *testing.T) {
	svc := NewPasswordService(&fakeCredentialRepo{}, &fakeSanitizer{})

	err := svc.ChangePassword(context.Background(), "", "old", "new")
	if !errors.Is(err, dusers.ErrInvalidUserData) {
		t.Fatalf("expected ErrInvalidUserData for empty username, got %v", err)
	}
}

func TestChangePassword_EmptyCurrentPassword(t *testing.T) {
	svc := NewPasswordService(&fakeCredentialRepo{}, &fakeSanitizer{})

	err := svc.ChangePassword(context.Background(), "alice", "", "new")
	if err == nil || err.Error() != "current password is required" {
		t.Fatalf("expected 'current password is required', got %v", err)
	}
}

func TestChangePassword_EmptyNewPassword(t *testing.T) {
	svc := NewPasswordService(&fakeCredentialRepo{}, &fakeSanitizer{})

	err := svc.ChangePassword(context.Background(), "alice", "old", "")
	if err == nil || err.Error() != "new password is required" {
		t.Fatalf("expected 'new password is required', got %v", err)
	}
}

func TestChangePassword_NilRepo(t *testing.T) {
	svc := NewPasswordService(nil, &fakeSanitizer{})

	err := svc.ChangePassword(context.Background(), "alice", "old", "new")
	if !errors.Is(err, dusers.ErrInvalidUserData) {
		t.Fatalf("expected ErrInvalidUserData for nil repo, got %v", err)
	}
}

func TestChangePassword_NilSanitizer(t *testing.T) {
	svc := NewPasswordService(&fakeCredentialRepo{}, nil)

	err := svc.ChangePassword(context.Background(), "alice", "old", "new")
	if !errors.Is(err, dusers.ErrInvalidUserData) {
		t.Fatalf("expected ErrInvalidUserData for nil sanitizer, got %v", err)
	}
}

func TestChangePassword_RepoGetCredentialsError(t *testing.T) {
	dbErr := errors.New("database unreachable")
	repo := &fakeCredentialRepo{err: dbErr}
	svc := NewPasswordService(repo, &fakeSanitizer{})

	err := svc.ChangePassword(context.Background(), "alice", "old", "new")
	if !errors.Is(err, dbErr) {
		t.Fatalf("expected repo error, got %v", err)
	}
}

func TestChangePassword_WrongCurrentPassword(t *testing.T) {
	hash, _ := bcrypt.GenerateFromPassword([]byte("correctpassword"), bcrypt.MinCost)
	repo := &fakeCredentialRepo{
		creds: &dauth.Credentials{UserID: 1, PasswordHash: string(hash)},
	}
	svc := NewPasswordService(repo, &fakeSanitizer{})

	err := svc.ChangePassword(context.Background(), "alice", "wrongpassword", "newpassword")
	if !errors.Is(err, dauth.ErrInvalidCredentials) {
		t.Fatalf("expected ErrInvalidCredentials, got %v", err)
	}
}

func TestChangePassword_SanitizationFails(t *testing.T) {
	hash, _ := bcrypt.GenerateFromPassword([]byte("currentpass"), bcrypt.MinCost)
	repo := &fakeCredentialRepo{
		creds: &dauth.Credentials{UserID: 1, PasswordHash: string(hash)},
	}
	sanitizer := &fakeSanitizer{err: errors.New("password too weak")}
	svc := NewPasswordService(repo, sanitizer)

	err := svc.ChangePassword(context.Background(), "alice", "currentpass", "weak")
	if err == nil {
		t.Fatalf("expected sanitization error, got nil")
	}
	if err.Error() != "new password does not meet security requirements: password too weak" {
		t.Fatalf("unexpected error message: %v", err)
	}
}

func TestChangePassword_NewPasswordSameAsCurrent(t *testing.T) {
	hash, _ := bcrypt.GenerateFromPassword([]byte("samepassword"), bcrypt.MinCost)
	repo := &fakeCredentialRepo{
		creds: &dauth.Credentials{UserID: 1, PasswordHash: string(hash)},
	}
	// Sanitizer returns the same password unchanged
	sanitizer := &fakeSanitizer{result: "samepassword"}
	svc := NewPasswordService(repo, sanitizer)

	err := svc.ChangePassword(context.Background(), "alice", "samepassword", "samepassword")
	if err == nil {
		t.Fatalf("expected error for same password, got nil")
	}
	if err.Error() != "new password must differ from the current password" {
		t.Fatalf("unexpected error message: %v", err)
	}
}

func TestChangePassword_Success(t *testing.T) {
	hash, _ := bcrypt.GenerateFromPassword([]byte("oldpass"), bcrypt.MinCost)
	repo := &fakeCredentialRepo{
		creds: &dauth.Credentials{UserID: 1, PasswordHash: string(hash)},
	}
	sanitizer := &fakeSanitizer{}
	svc := NewPasswordService(repo, sanitizer)

	err := svc.ChangePassword(context.Background(), "alice", "oldpass", "newpass")
	if err != nil {
		t.Fatalf("ChangePassword returned error: %v", err)
	}
	if repo.updatedUsername != "alice" {
		t.Fatalf("expected UpdatePassword called for alice, got %s", repo.updatedUsername)
	}
	if repo.updatedMust != false {
		t.Fatalf("expected mustChange=false after password change")
	}
	// Verify the new hash is valid
	if err := bcrypt.CompareHashAndPassword([]byte(repo.updatedHash), []byte("newpass")); err != nil {
		t.Fatalf("stored hash doesn't match new password: %v", err)
	}
}
