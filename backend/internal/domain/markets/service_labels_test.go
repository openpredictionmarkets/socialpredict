package markets_test

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	markets "socialpredict/internal/domain/markets"
)

type labelsRepo struct {
	market       *markets.Market
	getByIDErr   error
	updatedID    int64
	updatedYes   string
	updatedNo    string
	updateLabErr error
}

func (r *labelsRepo) Create(context.Context, *markets.Market) error { panic("unexpected call") }

func (r *labelsRepo) GetByID(_ context.Context, id int64) (*markets.Market, error) {
	if r.getByIDErr != nil {
		return nil, r.getByIDErr
	}
	if r.market == nil || r.market.ID != id {
		return nil, markets.ErrMarketNotFound
	}
	return r.market, nil
}

func (r *labelsRepo) UpdateLabels(_ context.Context, id int64, yesLabel, noLabel string) error {
	if r.updateLabErr != nil {
		return r.updateLabErr
	}
	r.updatedID = id
	r.updatedYes = yesLabel
	r.updatedNo = noLabel
	return nil
}

func (r *labelsRepo) List(context.Context, markets.ListFilters) ([]*markets.Market, error) {
	panic("unexpected call")
}
func (r *labelsRepo) ListByStatus(context.Context, string, markets.Page) ([]*markets.Market, error) {
	panic("unexpected call")
}
func (r *labelsRepo) Search(context.Context, string, markets.SearchFilters) ([]*markets.Market, error) {
	panic("unexpected call")
}
func (r *labelsRepo) Delete(context.Context, int64) error { panic("unexpected call") }
func (r *labelsRepo) ResolveMarket(context.Context, int64, string) error {
	panic("unexpected call")
}
func (r *labelsRepo) GetUserPosition(context.Context, int64, string) (*markets.UserPosition, error) {
	panic("unexpected call")
}
func (r *labelsRepo) ListMarketPositions(context.Context, int64) (markets.MarketPositions, error) {
	panic("unexpected call")
}
func (r *labelsRepo) ListBetsForMarket(context.Context, int64) ([]*markets.Bet, error) {
	panic("unexpected call")
}
func (r *labelsRepo) CalculatePayoutPositions(context.Context, int64) ([]*markets.PayoutPosition, error) {
	panic("unexpected call")
}
func (r *labelsRepo) GetPublicMarket(context.Context, int64) (*markets.PublicMarket, error) {
	panic("unexpected call")
}

type labelsClock struct{ now time.Time }

func (c labelsClock) Now() time.Time { return c.now }

func TestSetCustomLabels_Success(t *testing.T) {
	repo := &labelsRepo{market: &markets.Market{ID: 1}}
	svc := markets.NewServiceWithWallet(repo, noOpCreatorProfile{}, noOpWallet{}, labelsClock{}, markets.Config{})

	err := svc.SetCustomLabels(context.Background(), 1, "Agree", "Disagree")
	if err != nil {
		t.Fatalf("SetCustomLabels returned error: %v", err)
	}
	if repo.updatedID != 1 {
		t.Fatalf("expected UpdateLabels called with ID 1, got %d", repo.updatedID)
	}
	if repo.updatedYes != "Agree" || repo.updatedNo != "Disagree" {
		t.Fatalf("expected labels (Agree, Disagree), got (%s, %s)", repo.updatedYes, repo.updatedNo)
	}
}

func TestSetCustomLabels_MarketNotFound(t *testing.T) {
	repo := &labelsRepo{} // no market set
	svc := markets.NewServiceWithWallet(repo, noOpCreatorProfile{}, noOpWallet{}, labelsClock{}, markets.Config{})

	err := svc.SetCustomLabels(context.Background(), 999, "Yes", "No")
	if err != markets.ErrMarketNotFound {
		t.Fatalf("expected ErrMarketNotFound, got %v", err)
	}
}

func TestSetCustomLabels_InvalidLabelTooLong(t *testing.T) {
	repo := &labelsRepo{market: &markets.Market{ID: 1}}
	svc := markets.NewServiceWithWallet(repo, noOpCreatorProfile{}, noOpWallet{}, labelsClock{}, markets.Config{})

	longLabel := strings.Repeat("x", 21) // MaxLabelLength is 20
	err := svc.SetCustomLabels(context.Background(), 1, longLabel, "No")
	if err != markets.ErrInvalidLabel {
		t.Fatalf("expected ErrInvalidLabel for long yes label, got %v", err)
	}

	err = svc.SetCustomLabels(context.Background(), 1, "Yes", longLabel)
	if err != markets.ErrInvalidLabel {
		t.Fatalf("expected ErrInvalidLabel for long no label, got %v", err)
	}
}

func TestSetCustomLabels_EmptyLabelsAllowed(t *testing.T) {
	repo := &labelsRepo{market: &markets.Market{ID: 1}}
	svc := markets.NewServiceWithWallet(repo, noOpCreatorProfile{}, noOpWallet{}, labelsClock{}, markets.Config{})

	err := svc.SetCustomLabels(context.Background(), 1, "", "")
	if err != nil {
		t.Fatalf("expected empty labels to be allowed, got %v", err)
	}
}

func TestSetCustomLabels_RepoUpdateError(t *testing.T) {
	repoErr := errors.New("database unavailable")
	repo := &labelsRepo{market: &markets.Market{ID: 1}, updateLabErr: repoErr}
	svc := markets.NewServiceWithWallet(repo, noOpCreatorProfile{}, noOpWallet{}, labelsClock{}, markets.Config{})

	err := svc.SetCustomLabels(context.Background(), 1, "Yes", "No")
	if !errors.Is(err, repoErr) {
		t.Fatalf("expected repo error, got %v", err)
	}
}

func TestSetCustomLabels_LabelExactlyAtMaxLength(t *testing.T) {
	repo := &labelsRepo{market: &markets.Market{ID: 1}}
	svc := markets.NewServiceWithWallet(repo, noOpCreatorProfile{}, noOpWallet{}, labelsClock{}, markets.Config{})

	exactLabel := strings.Repeat("x", 20) // MaxLabelLength is 20
	err := svc.SetCustomLabels(context.Background(), 1, exactLabel, "No")
	if err != nil {
		t.Fatalf("expected label at max length to succeed, got %v", err)
	}

	err = svc.SetCustomLabels(context.Background(), 1, "Yes", exactLabel)
	if err != nil {
		t.Fatalf("expected label at max length to succeed, got %v", err)
	}
}

func TestSetCustomLabels_WhitespaceTrimmedForLengthCheck(t *testing.T) {
	repo := &labelsRepo{market: &markets.Market{ID: 1}}
	svc := markets.NewServiceWithWallet(repo, noOpCreatorProfile{}, noOpWallet{}, labelsClock{}, markets.Config{})

	// " YES " trims to "YES" (3 chars) — should pass validation
	err := svc.SetCustomLabels(context.Background(), 1, " YES ", " NO ")
	if err != nil {
		t.Fatalf("expected padded labels to pass after trimming, got %v", err)
	}
	// But the original (untrimmed) string is written to repo
	if repo.updatedYes != " YES " {
		t.Fatalf("expected repo to receive original string ' YES ', got %q", repo.updatedYes)
	}
	if repo.updatedNo != " NO " {
		t.Fatalf("expected repo to receive original string ' NO ', got %q", repo.updatedNo)
	}
}

func TestSetCustomLabels_WhitespaceOnlyLabel(t *testing.T) {
	repo := &labelsRepo{market: &markets.Market{ID: 1}}
	svc := markets.NewServiceWithWallet(repo, noOpCreatorProfile{}, noOpWallet{}, labelsClock{}, markets.Config{})

	// " " is non-empty so validation kicks in, trims to "" which is < MinLabelLength
	err := svc.SetCustomLabels(context.Background(), 1, " ", "No")
	if err != markets.ErrInvalidLabel {
		t.Fatalf("expected ErrInvalidLabel for whitespace-only label, got %v", err)
	}
}

func TestSetCustomLabels_OneEmptyOneNonEmpty(t *testing.T) {
	repo := &labelsRepo{market: &markets.Market{ID: 1}}
	svc := markets.NewServiceWithWallet(repo, noOpCreatorProfile{}, noOpWallet{}, labelsClock{}, markets.Config{})

	err := svc.SetCustomLabels(context.Background(), 1, "Agree", "")
	if err != nil {
		t.Fatalf("expected yes-only label to succeed, got %v", err)
	}
	if repo.updatedYes != "Agree" || repo.updatedNo != "" {
		t.Fatalf("expected (Agree, ''), got (%s, %s)", repo.updatedYes, repo.updatedNo)
	}

	repo.updatedYes = ""
	repo.updatedNo = ""
	err = svc.SetCustomLabels(context.Background(), 1, "", "Disagree")
	if err != nil {
		t.Fatalf("expected no-only label to succeed, got %v", err)
	}
	if repo.updatedYes != "" || repo.updatedNo != "Disagree" {
		t.Fatalf("expected ('', Disagree), got (%s, %s)", repo.updatedYes, repo.updatedNo)
	}
}

func TestSetCustomLabels_GetByIDNonNotFoundError(t *testing.T) {
	// Any GetByID error gets mapped to ErrMarketNotFound by the implementation
	dbErr := errors.New("database connection lost")
	repo := &labelsRepo{market: &markets.Market{ID: 1}, getByIDErr: dbErr}
	svc := markets.NewServiceWithWallet(repo, noOpCreatorProfile{}, noOpWallet{}, labelsClock{}, markets.Config{})

	err := svc.SetCustomLabels(context.Background(), 1, "Yes", "No")
	if err != markets.ErrMarketNotFound {
		t.Fatalf("expected ErrMarketNotFound for non-not-found GetByID error, got %v", err)
	}
}
