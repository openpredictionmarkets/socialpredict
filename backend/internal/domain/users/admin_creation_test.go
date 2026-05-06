package users

import (
	"context"
	"errors"
	"testing"

	"golang.org/x/crypto/bcrypt"
)

func TestServiceCreateAdminManagedUserCreatesRegularUser(t *testing.T) {
	repo := &adminCreateRepo{}
	service := NewServiceWithDependencies(ServiceDependencies{
		Writer:     repo,
		Uniqueness: repo,
	}, nil, nil)

	result, err := service.CreateAdminManagedUser(context.Background(), AdminManagedUserCreateRequest{
		Username:              "freshuser",
		InitialAccountBalance: 250,
	})
	if err != nil {
		t.Fatalf("CreateAdminManagedUser returned error: %v", err)
	}

	if result.Username != "freshuser" || result.UserType != "REGULAR" || result.Password == "" {
		t.Fatalf("unexpected create result: %+v", result)
	}

	created := repo.created
	if created == nil {
		t.Fatalf("expected repository create call")
	}
	if created.Username != "freshuser" || created.UserType != "REGULAR" {
		t.Fatalf("unexpected created user identity: %+v", created)
	}
	if created.InitialAccountBalance != 250 || created.AccountBalance != 250 {
		t.Fatalf("expected seeded balances of 250, got initial=%d account=%d", created.InitialAccountBalance, created.AccountBalance)
	}
	if created.DisplayName == "" || created.Email == "" || created.APIKey == "" || created.PersonalEmoji == "" {
		t.Fatalf("expected generated identity fields, got %+v", created)
	}
	if !created.MustChangePassword {
		t.Fatalf("expected created user to require password change")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(created.PasswordHash), []byte(result.Password)); err != nil {
		t.Fatalf("expected stored password hash to match returned password: %v", err)
	}
}

func TestServiceCreateAdminManagedUserRejectsExistingUsername(t *testing.T) {
	repo := &adminCreateRepo{existingUsernames: map[string]bool{"freshuser": true}}
	service := NewServiceWithDependencies(ServiceDependencies{
		Writer:     repo,
		Uniqueness: repo,
	}, nil, nil)

	_, err := service.CreateAdminManagedUser(context.Background(), AdminManagedUserCreateRequest{
		Username:              "freshuser",
		InitialAccountBalance: 250,
	})
	if !errors.Is(err, ErrUserAlreadyExists) {
		t.Fatalf("expected ErrUserAlreadyExists, got %v", err)
	}
	if repo.created != nil {
		t.Fatalf("expected no repository create call")
	}
}

type adminCreateRepo struct {
	existingUsernames map[string]bool
	created           *User
}

func (r *adminCreateRepo) Create(_ context.Context, user *User) error {
	copy := *user
	r.created = &copy
	return nil
}

func (r *adminCreateRepo) Update(context.Context, *User) error {
	return nil
}

func (r *adminCreateRepo) Delete(context.Context, string) error {
	return nil
}

func (r *adminCreateRepo) UsernameExists(_ context.Context, username string) (bool, error) {
	return r.existingUsernames[username], nil
}

func (r *adminCreateRepo) DisplayNameExists(context.Context, string) (bool, error) {
	return false, nil
}

func (r *adminCreateRepo) EmailExists(context.Context, string) (bool, error) {
	return false, nil
}

func (r *adminCreateRepo) APIKeyExists(context.Context, string) (bool, error) {
	return false, nil
}

func (r *adminCreateRepo) AnyUserIdentityExists(context.Context, string, string, string, string) (bool, error) {
	return false, nil
}
