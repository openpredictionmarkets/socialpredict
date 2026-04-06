package users_test

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"

	users "socialpredict/internal/domain/users"
	"socialpredict/security"

	"golang.org/x/crypto/bcrypt"
)

type fakeRepository struct {
	user         *users.User
	passwordHash string
	mustChange   bool

	getByUsernameFn     func(context.Context, string) (*users.User, error)
	updateBalanceFn     func(context.Context, string, int64) error
	createFn            func(context.Context, *users.User) error
	updateFn            func(context.Context, *users.User) error
	deleteFn            func(context.Context, string) error
	listFn              func(context.Context, users.ListFilters) ([]*users.User, error)
	listUserBetsFn      func(context.Context, string) ([]*users.UserBet, error)
	getMarketQuestionFn func(context.Context, uint) (string, error)
	getUserPositionFn   func(context.Context, int64, string) (*users.MarketUserPosition, error)
	listUserMarketsFn   func(context.Context, int64) ([]*users.UserMarket, error)
	getCredentialsFn    func(context.Context, string) (*users.Credentials, error)
	updatePasswordFn    func(context.Context, string, string, bool) error
}

var (
	_ users.UserReader              = (*fakeRepository)(nil)
	_ users.UserBalanceRepository   = (*fakeRepository)(nil)
	_ users.UserWriter              = (*fakeRepository)(nil)
	_ users.UserLister              = (*fakeRepository)(nil)
	_ users.UserPortfolioRepository = (*fakeRepository)(nil)
	_ users.UserMarketsRepository   = (*fakeRepository)(nil)
	_ users.CredentialsRepository   = (*fakeRepository)(nil)
)

const initialTestPassword = "CurrentPass123!"

func newFakeRepository(username string) *fakeRepository {
	hash, _ := bcrypt.GenerateFromPassword([]byte(initialTestPassword), users.PasswordHashCost())
	return &fakeRepository{
		user:         seededUser(username),
		passwordHash: string(hash),
		mustChange:   true,
	}
}

func newServiceDependencies(repo *fakeRepository) users.ServiceDependencies {
	return users.ServiceDependencies{
		Reader:      repo,
		BalanceRepo: repo,
		Writer:      repo,
		Lister:      repo,
		Portfolio:   repo,
		Markets:     repo,
		Credentials: repo,
	}
}

func seededUser(username string) *users.User {
	return &users.User{
		ID:                 1,
		Username:           username,
		DisplayName:        "Display " + username,
		Email:              username + "@example.com",
		UserType:           "regular",
		MustChangePassword: true,
	}
}

func cloneUser(user *users.User) *users.User {
	if user == nil {
		return nil
	}

	copy := *user
	return &copy
}

func (f *fakeRepository) storedUser(username string) (*users.User, error) {
	if f.user == nil || f.user.Username != username {
		return nil, users.ErrUserNotFound
	}
	copy := cloneUser(f.user)
	copy.MustChangePassword = f.mustChange
	return copy, nil
}

func (f *fakeRepository) persistUser(user *users.User) {
	f.user = cloneUser(user)
	if user != nil {
		f.mustChange = user.MustChangePassword
	}
}

func (f *fakeRepository) GetByUsername(ctx context.Context, username string) (*users.User, error) {
	if f.getByUsernameFn != nil {
		return f.getByUsernameFn(ctx, username)
	}
	return f.storedUser(username)
}

func (f *fakeRepository) UpdateBalance(ctx context.Context, username string, newBalance int64) error {
	if f.updateBalanceFn != nil {
		return f.updateBalanceFn(ctx, username, newBalance)
	}
	user, err := f.storedUser(username)
	if err != nil {
		return err
	}
	user.AccountBalance = newBalance
	f.persistUser(user)
	return nil
}

func (f *fakeRepository) Create(ctx context.Context, user *users.User) error {
	if f.createFn != nil {
		return f.createFn(ctx, user)
	}
	f.persistUser(user)
	return nil
}

func (f *fakeRepository) Update(ctx context.Context, user *users.User) error {
	if f.updateFn != nil {
		return f.updateFn(ctx, user)
	}
	f.persistUser(user)
	return nil
}

func (f *fakeRepository) Delete(ctx context.Context, username string) error {
	if f.deleteFn != nil {
		return f.deleteFn(ctx, username)
	}
	if _, err := f.storedUser(username); err != nil {
		return err
	}
	f.user = nil
	return nil
}

func (f *fakeRepository) List(ctx context.Context, filters users.ListFilters) ([]*users.User, error) {
	if f.listFn != nil {
		return f.listFn(ctx, filters)
	}
	return []*users.User{}, nil
}

func (f *fakeRepository) ListUserBets(ctx context.Context, username string) ([]*users.UserBet, error) {
	if f.listUserBetsFn != nil {
		return f.listUserBetsFn(ctx, username)
	}
	return []*users.UserBet{}, nil
}

func (f *fakeRepository) GetMarketQuestion(ctx context.Context, marketID uint) (string, error) {
	if f.getMarketQuestionFn != nil {
		return f.getMarketQuestionFn(ctx, marketID)
	}
	return fmt.Sprintf("market-%d", marketID), nil
}

func (f *fakeRepository) GetUserPositionInMarket(ctx context.Context, marketID int64, username string) (*users.MarketUserPosition, error) {
	if f.getUserPositionFn != nil {
		return f.getUserPositionFn(ctx, marketID, username)
	}
	return &users.MarketUserPosition{}, nil
}

func (f *fakeRepository) ListUserMarkets(ctx context.Context, userID int64) ([]*users.UserMarket, error) {
	if f.listUserMarketsFn != nil {
		return f.listUserMarketsFn(ctx, userID)
	}
	return []*users.UserMarket{}, nil
}

func (f *fakeRepository) GetCredentials(ctx context.Context, username string) (*users.Credentials, error) {
	if f.getCredentialsFn != nil {
		return f.getCredentialsFn(ctx, username)
	}
	if _, err := f.storedUser(username); err != nil {
		return nil, err
	}
	return &users.Credentials{
		PasswordHash:       f.passwordHash,
		MustChangePassword: f.mustChange,
	}, nil
}

func (f *fakeRepository) UpdatePassword(ctx context.Context, username string, hashedPassword string, mustChange bool) error {
	if f.updatePasswordFn != nil {
		return f.updatePasswordFn(ctx, username, hashedPassword, mustChange)
	}
	user, err := f.storedUser(username)
	if err != nil {
		return err
	}
	f.passwordHash = hashedPassword
	f.mustChange = mustChange
	user.MustChangePassword = mustChange
	f.persistUser(user)
	return nil
}

func newServiceWithUser(t *testing.T) (string, users.ServiceInterface, *fakeRepository, context.Context) {
	t.Helper()

	username := fmt.Sprintf("profile_%s", strings.ToLower(t.Name()))
	repo := newFakeRepository(username)
	var service users.ServiceInterface = users.NewServiceWithDependencies(newServiceDependencies(repo), nil, security.NewSecurityService().Sanitizer)

	return username, service, repo, context.Background()
}

func TestServiceUpdateDescription(t *testing.T) {
	username, service, _, ctx := newServiceWithUser(t)

	updated, err := service.UpdateDescription(ctx, username, "   Friendly <b>description</b>   ")
	if err != nil {
		t.Fatalf("UpdateDescription returned error: %v", err)
	}
	if updated.Description == "" {
		t.Fatalf("expected sanitized description, got empty string")
	}
	if strings.Contains(updated.Description, "<script>") {
		t.Fatalf("expected script tags removed, got %q", updated.Description)
	}
	public, err := service.GetPublicUser(ctx, username)
	if err != nil {
		t.Fatalf("GetPublicUser returned error: %v", err)
	}
	if public.Description != updated.Description {
		t.Fatalf("expected persisted description %q, got %q", updated.Description, public.Description)
	}

	if _, err := service.UpdateDescription(ctx, username, strings.Repeat("a", 2001)); err == nil {
		t.Fatal("expected error for overlong description")
	}
	if _, err := service.UpdateDescription(ctx, username, "bad<script>alert(1)</script>"); err == nil {
		t.Fatal("expected error for unsafe description content")
	}
}

func TestServiceUpdateDisplayName(t *testing.T) {
	username, service, _, ctx := newServiceWithUser(t)

	updated, err := service.UpdateDisplayName(ctx, username, "  New Name  ")
	if err != nil {
		t.Fatalf("UpdateDisplayName returned error: %v", err)
	}
	if updated.DisplayName != "New Name" {
		t.Fatalf("expected trimmed display name, got %q", updated.DisplayName)
	}
	public, err := service.GetPublicUser(ctx, username)
	if err != nil {
		t.Fatalf("GetPublicUser returned error: %v", err)
	}
	if public.DisplayName != updated.DisplayName {
		t.Fatalf("expected persisted display name %q, got %q", updated.DisplayName, public.DisplayName)
	}

	if _, err := service.UpdateDisplayName(ctx, username, ""); err == nil {
		t.Fatal("expected error for empty display name")
	}
	if _, err := service.UpdateDisplayName(ctx, username, strings.Repeat("b", 51)); err == nil {
		t.Fatal("expected error for overlong display name")
	}
	if _, err := service.UpdateDisplayName(ctx, username, "bad<script>alert(1)</script>"); err == nil {
		t.Fatal("expected error for unsafe display name content")
	}
}

func TestServiceUpdateEmoji(t *testing.T) {
	username, service, _, ctx := newServiceWithUser(t)

	updated, err := service.UpdateEmoji(ctx, username, "😊")
	if err != nil {
		t.Fatalf("UpdateEmoji returned error: %v", err)
	}
	if updated.PersonalEmoji != "😊" {
		t.Fatalf("expected emoji to persist, got %q", updated.PersonalEmoji)
	}
	public, err := service.GetPublicUser(ctx, username)
	if err != nil {
		t.Fatalf("GetPublicUser returned error: %v", err)
	}
	if public.PersonalEmoji != updated.PersonalEmoji {
		t.Fatalf("expected persisted emoji %q, got %q", updated.PersonalEmoji, public.PersonalEmoji)
	}

	if _, err := service.UpdateEmoji(ctx, username, ""); err == nil {
		t.Fatal("expected error for blank emoji")
	}
	if _, err := service.UpdateEmoji(ctx, username, strings.Repeat("😀", 21)); err == nil {
		t.Fatal("expected error for overlong emoji")
	}
}

func TestServiceUpdatePersonalLinks(t *testing.T) {
	username, service, _, ctx := newServiceWithUser(t)

	links := users.PersonalLinks{
		PersonalLink1: "example.com",
		PersonalLink2: "",
		PersonalLink3: "https://valid.example",
		PersonalLink4: "http://valid.example/path",
	}

	updated, err := service.UpdatePersonalLinks(ctx, username, links)
	if err != nil {
		t.Fatalf("UpdatePersonalLinks returned error: %v", err)
	}
	if updated.PersonalLink1 == "" || !strings.HasPrefix(updated.PersonalLink1, "https://") {
		t.Fatalf("expected sanitized link with https prefix, got %q", updated.PersonalLink1)
	}
	if updated.PersonalLink2 != "" {
		t.Fatalf("expected empty link to remain empty, got %q", updated.PersonalLink2)
	}
	public, err := service.GetPublicUser(ctx, username)
	if err != nil {
		t.Fatalf("GetPublicUser returned error: %v", err)
	}
	if public.PersonalLink1 != updated.PersonalLink1 || public.PersonalLink4 != updated.PersonalLink4 {
		t.Fatalf("expected persisted links to match updates: %+v vs %+v", public, updated)
	}

	longLink := strings.Repeat("a", 201)
	if _, err := service.UpdatePersonalLinks(ctx, username, users.PersonalLinks{PersonalLink1: longLink}); err == nil {
		t.Fatal("expected error for overly long personal link")
	}
	if _, err := service.UpdatePersonalLinks(ctx, username, users.PersonalLinks{PersonalLink1: "javascript:alert('xss')"}); err == nil {
		t.Fatal("expected error for unsafe personal link")
	}
}

func TestServiceChangePassword(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		username, service, repo, ctx := newServiceWithUser(t)

		if err := service.ChangePassword(ctx, username, initialTestPassword, "NewPassword456!"); err != nil {
			t.Fatalf("ChangePassword returned error: %v", err)
		}

		if err := bcrypt.CompareHashAndPassword([]byte(repo.passwordHash), []byte("NewPassword456!")); err != nil {
			t.Fatalf("expected password hash to update: %v", err)
		}
		if repo.mustChange {
			t.Fatalf("expected mustChangePassword to be cleared")
		}
	})

	t.Run("invalid current password", func(t *testing.T) {
		username, service, _, ctx := newServiceWithUser(t)

		err := service.ChangePassword(ctx, username, "wrong", "AnotherPass789!")
		if !errors.Is(err, users.ErrInvalidCredentials) {
			t.Fatalf("expected ErrInvalidCredentials, got %v", err)
		}
	})

	t.Run("weak new password", func(t *testing.T) {
		username, service, _, ctx := newServiceWithUser(t)

		if err := service.ChangePassword(ctx, username, initialTestPassword, "short"); err == nil {
			t.Fatal("expected error for weak password")
		}
	})

	t.Run("same password", func(t *testing.T) {
		username, service, _, ctx := newServiceWithUser(t)

		if err := service.ChangePassword(ctx, username, initialTestPassword, initialTestPassword); err == nil {
			t.Fatal("expected error when new password matches current password")
		}
	})
}

func TestServiceGetPrivateProfile(t *testing.T) {
	username, service, repo, ctx := newServiceWithUser(t)

	profile, err := service.GetPrivateProfile(ctx, username)
	if err != nil {
		t.Fatalf("GetPrivateProfile returned error: %v", err)
	}

	if profile.Username != username {
		t.Fatalf("expected username %q, got %q", username, profile.Username)
	}
	if profile.Email == "" {
		t.Fatalf("expected email to be populated")
	}

	// simulate missing user
	repo.user = nil
	if _, err := service.GetPrivateProfile(ctx, username); err == nil {
		t.Fatal("expected error for missing user")
	}
}
