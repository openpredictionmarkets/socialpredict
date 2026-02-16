package auth

import (
	"context"
	"errors"
	"testing"

	dauth "socialpredict/internal/domain/auth"
	dusers "socialpredict/internal/domain/users"
)

// --- Test doubles ---

type fakeCredentialRepo struct {
	creds *dauth.Credentials
	err   error
}

func (r *fakeCredentialRepo) GetCredentials(_ context.Context, _ string) (*dauth.Credentials, error) {
	if r.err != nil {
		return nil, r.err
	}
	return r.creds, nil
}

func (r *fakeCredentialRepo) UpdatePassword(_ context.Context, _, _ string, _ bool) error {
	return nil
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
