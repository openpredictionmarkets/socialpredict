package markets_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"socialpredict/internal/domain/boundary"
	markets "socialpredict/internal/domain/markets"
	"socialpredict/internal/domain/math/probabilities/wpam"
)

type projectionRepo struct {
	createFunc                                      func(context.Context, *markets.Market) error
	updateLabelsFunc                                func(context.Context, int64, string, string) error
	listFunc                                        func(context.Context, markets.ListFilters) ([]*markets.Market, error)
	listByStatusFunc                                func(context.Context, string, markets.Page) ([]*markets.Market, error)
	searchFunc                                      func(context.Context, string, markets.SearchFilters) ([]*markets.Market, error)
	deleteFunc                                      func(context.Context, int64) error
	resolveMarketFunc                               func(context.Context, int64, string) error
	getUserPositionFunc                             func(context.Context, int64, string) (*markets.UserPosition, error)
	listMarketPositionsFunc                         func(context.Context, int64) (markets.MarketPositions, error)
	listBetsForMarketFunc                           func(context.Context, int64) ([]*markets.Bet, error)
	getByIDFunc                                     func(context.Context, int64) (*markets.Market, error)
	calculatePayoutPositionsFunc                    func(context.Context, int64) ([]*markets.PayoutPosition, error)
	getPublicMarketFunc                             func(context.Context, int64) (*markets.PublicMarket, error)
	approveMarketFunc                               func(context.Context, int64, string, time.Time) error
	rejectMarketFunc                                func(context.Context, int64, string, time.Time, string) error
	reassignMarketStewardFunc                       func(context.Context, int64, string, string, string, string, time.Time) error
	listMarketTagsFunc                              func(context.Context, bool) ([]markets.MarketTag, error)
	createMarketTagFunc                             func(context.Context, markets.MarketTag) (*markets.MarketTag, error)
	updateMarketTagFunc                             func(context.Context, string, markets.MarketTagRequest) (*markets.MarketTag, error)
	setMarketTagsFunc                               func(context.Context, int64, []string, string, string, time.Time) ([]markets.MarketTag, error)
	setMarketGroupTagsFunc                          func(context.Context, int64, []string, string, string, time.Time) ([]markets.MarketTag, error)
	getMarketGovernanceSettingsFunc                 func(context.Context) (*markets.MarketGovernanceSettings, error)
	updateMarketGovernanceSettingsFunc              func(context.Context, markets.MarketGovernanceSettingsUpdate) (*markets.MarketGovernanceSettings, error)
	listAdminMarketReviewRowsFunc                   func(context.Context, markets.AdminMarketReviewFilters) (*markets.AdminMarketReviewPage, error)
	listLifecycleMarketDiscoveryFunc                func(context.Context, markets.ListFilters) (*markets.MarketDiscoveryPage, error)
	listDescriptionAmendmentReviewCandidatesFunc    func(context.Context, markets.AdminDescriptionAmendmentReviewFilters) ([]markets.MarketDescriptionAmendment, int, error)
	listAnswerAdditionsForAdminReviewFunc           func(context.Context, markets.AdminAnswerAdditionReviewFilters) ([]markets.MarketGroupAnswerAddition, int, error)
	reviewGroupedDescriptionAmendmentsFunc          func(context.Context, []int64, string, string, string, time.Time) ([]markets.MarketDescriptionAmendment, error)
	createMarketGroupFunc                           func(context.Context, *markets.MarketGroup, []markets.MarketGroupMember) error
	getMarketGroupFunc                              func(context.Context, int64) (*markets.MarketGroup, error)
	listMarketGroupMembersFunc                      func(context.Context, int64) ([]markets.MarketGroupMember, error)
	getMarketGroupForMarketFunc                     func(context.Context, int64) (*markets.MarketGroup, error)
	markMarketGroupResolvedFunc                     func(context.Context, int64, time.Time) error
	updateMarketGroupAnswerAdditionAutoApprovalFunc func(context.Context, int64, bool, time.Time) (*markets.MarketGroup, error)
}

func newProjectionRepo(opts ...func(*projectionRepo)) *projectionRepo {
	repo := &projectionRepo{
		createFunc:       func(context.Context, *markets.Market) error { return errUnexpectedMarketsTestCall },
		updateLabelsFunc: func(context.Context, int64, string, string) error { return errUnexpectedMarketsTestCall },
		listFunc: func(context.Context, markets.ListFilters) ([]*markets.Market, error) {
			return nil, errUnexpectedMarketsTestCall
		},
		listByStatusFunc: func(context.Context, string, markets.Page) ([]*markets.Market, error) {
			return nil, errUnexpectedMarketsTestCall
		},
		searchFunc: func(context.Context, string, markets.SearchFilters) ([]*markets.Market, error) {
			return nil, errUnexpectedMarketsTestCall
		},
		deleteFunc:        func(context.Context, int64) error { return errUnexpectedMarketsTestCall },
		resolveMarketFunc: func(context.Context, int64, string) error { return errUnexpectedMarketsTestCall },
		getUserPositionFunc: func(context.Context, int64, string) (*markets.UserPosition, error) {
			return nil, errUnexpectedMarketsTestCall
		},
		listMarketPositionsFunc: func(context.Context, int64) (markets.MarketPositions, error) {
			return nil, errUnexpectedMarketsTestCall
		},
		listBetsForMarketFunc: func(context.Context, int64) ([]*markets.Bet, error) {
			return nil, errUnexpectedMarketsTestCall
		},
		getByIDFunc: func(context.Context, int64) (*markets.Market, error) {
			return nil, errUnexpectedMarketsTestCall
		},
		calculatePayoutPositionsFunc: func(context.Context, int64) ([]*markets.PayoutPosition, error) {
			return nil, errUnexpectedMarketsTestCall
		},
		getPublicMarketFunc: func(context.Context, int64) (*markets.PublicMarket, error) {
			return nil, errUnexpectedMarketsTestCall
		},
		approveMarketFunc: func(context.Context, int64, string, time.Time) error {
			return errUnexpectedMarketsTestCall
		},
		rejectMarketFunc: func(context.Context, int64, string, time.Time, string) error {
			return errUnexpectedMarketsTestCall
		},
		reassignMarketStewardFunc: func(context.Context, int64, string, string, string, string, time.Time) error {
			return errUnexpectedMarketsTestCall
		},
		listMarketTagsFunc: func(context.Context, bool) ([]markets.MarketTag, error) {
			return nil, errUnexpectedMarketsTestCall
		},
		createMarketTagFunc: func(context.Context, markets.MarketTag) (*markets.MarketTag, error) {
			return nil, errUnexpectedMarketsTestCall
		},
		updateMarketTagFunc: func(context.Context, string, markets.MarketTagRequest) (*markets.MarketTag, error) {
			return nil, errUnexpectedMarketsTestCall
		},
		setMarketTagsFunc: func(context.Context, int64, []string, string, string, time.Time) ([]markets.MarketTag, error) {
			return nil, errUnexpectedMarketsTestCall
		},
		setMarketGroupTagsFunc: func(context.Context, int64, []string, string, string, time.Time) ([]markets.MarketTag, error) {
			return nil, errUnexpectedMarketsTestCall
		},
		getMarketGovernanceSettingsFunc: func(context.Context) (*markets.MarketGovernanceSettings, error) {
			return nil, errUnexpectedMarketsTestCall
		},
		updateMarketGovernanceSettingsFunc: func(context.Context, markets.MarketGovernanceSettingsUpdate) (*markets.MarketGovernanceSettings, error) {
			return nil, errUnexpectedMarketsTestCall
		},
		listAdminMarketReviewRowsFunc: func(context.Context, markets.AdminMarketReviewFilters) (*markets.AdminMarketReviewPage, error) {
			return nil, errUnexpectedMarketsTestCall
		},
		listLifecycleMarketDiscoveryFunc: func(context.Context, markets.ListFilters) (*markets.MarketDiscoveryPage, error) {
			return nil, errUnexpectedMarketsTestCall
		},
		listDescriptionAmendmentReviewCandidatesFunc: func(context.Context, markets.AdminDescriptionAmendmentReviewFilters) ([]markets.MarketDescriptionAmendment, int, error) {
			return nil, 0, errUnexpectedMarketsTestCall
		},
		listAnswerAdditionsForAdminReviewFunc: func(context.Context, markets.AdminAnswerAdditionReviewFilters) ([]markets.MarketGroupAnswerAddition, int, error) {
			return nil, 0, errUnexpectedMarketsTestCall
		},
		reviewGroupedDescriptionAmendmentsFunc: func(context.Context, []int64, string, string, string, time.Time) ([]markets.MarketDescriptionAmendment, error) {
			return nil, errUnexpectedMarketsTestCall
		},
		createMarketGroupFunc: func(context.Context, *markets.MarketGroup, []markets.MarketGroupMember) error {
			return errUnexpectedMarketsTestCall
		},
		getMarketGroupFunc: func(context.Context, int64) (*markets.MarketGroup, error) {
			return nil, errUnexpectedMarketsTestCall
		},
		listMarketGroupMembersFunc: func(context.Context, int64) ([]markets.MarketGroupMember, error) {
			return nil, errUnexpectedMarketsTestCall
		},
		getMarketGroupForMarketFunc: func(context.Context, int64) (*markets.MarketGroup, error) {
			return nil, errUnexpectedMarketsTestCall
		},
		markMarketGroupResolvedFunc: func(context.Context, int64, time.Time) error {
			return errUnexpectedMarketsTestCall
		},
		updateMarketGroupAnswerAdditionAutoApprovalFunc: func(context.Context, int64, bool, time.Time) (*markets.MarketGroup, error) {
			return nil, errUnexpectedMarketsTestCall
		},
	}
	for _, opt := range opts {
		opt(repo)
	}
	return repo
}

func withProjectionRepoMarket(market *markets.Market) func(*projectionRepo) {
	return func(repo *projectionRepo) {
		repo.getByIDFunc = func(context.Context, int64) (*markets.Market, error) {
			if market == nil {
				return nil, markets.ErrMarketNotFound
			}
			return market, nil
		}
	}
}

func withProjectionRepoBets(bets []*markets.Bet) func(*projectionRepo) {
	return func(repo *projectionRepo) {
		repo.listBetsForMarketFunc = func(context.Context, int64) ([]*markets.Bet, error) {
			return bets, nil
		}
	}
}

func (r *projectionRepo) Create(ctx context.Context, market *markets.Market) error {
	if r.createFunc == nil {
		return errUnexpectedMarketsTestCall
	}
	return r.createFunc(ctx, market)
}
func (r *projectionRepo) UpdateLabels(ctx context.Context, id int64, yesLabel string, noLabel string) error {
	if r.updateLabelsFunc == nil {
		return errUnexpectedMarketsTestCall
	}
	return r.updateLabelsFunc(ctx, id, yesLabel, noLabel)
}
func (r *projectionRepo) List(ctx context.Context, filters markets.ListFilters) ([]*markets.Market, error) {
	if r.listFunc == nil {
		return nil, errUnexpectedMarketsTestCall
	}
	return r.listFunc(ctx, filters)
}
func (r *projectionRepo) ListByStatus(ctx context.Context, status string, page markets.Page) ([]*markets.Market, error) {
	if r.listByStatusFunc == nil {
		return nil, errUnexpectedMarketsTestCall
	}
	return r.listByStatusFunc(ctx, status, page)
}
func (r *projectionRepo) Search(ctx context.Context, query string, filters markets.SearchFilters) ([]*markets.Market, error) {
	if r.searchFunc == nil {
		return nil, errUnexpectedMarketsTestCall
	}
	return r.searchFunc(ctx, query, filters)
}
func (r *projectionRepo) Delete(ctx context.Context, id int64) error {
	if r.deleteFunc == nil {
		return errUnexpectedMarketsTestCall
	}
	return r.deleteFunc(ctx, id)
}
func (r *projectionRepo) ResolveMarket(ctx context.Context, id int64, outcome string) error {
	if r.resolveMarketFunc == nil {
		return errUnexpectedMarketsTestCall
	}
	return r.resolveMarketFunc(ctx, id, outcome)
}

func (r *projectionRepo) ApproveMarket(ctx context.Context, id int64, actorUsername string, approvedAt time.Time) error {
	if r.approveMarketFunc == nil {
		return errUnexpectedMarketsTestCall
	}
	return r.approveMarketFunc(ctx, id, actorUsername, approvedAt)
}

func (r *projectionRepo) RejectMarket(ctx context.Context, id int64, actorUsername string, rejectedAt time.Time, reason string) error {
	if r.rejectMarketFunc == nil {
		return errUnexpectedMarketsTestCall
	}
	return r.rejectMarketFunc(ctx, id, actorUsername, rejectedAt, reason)
}

func (r *projectionRepo) ReassignMarketSteward(ctx context.Context, id int64, fromStewardUsername string, toStewardUsername string, actorUsername string, reason string, changedAt time.Time) error {
	if r.reassignMarketStewardFunc == nil {
		return errUnexpectedMarketsTestCall
	}
	return r.reassignMarketStewardFunc(ctx, id, fromStewardUsername, toStewardUsername, actorUsername, reason, changedAt)
}

func (r *projectionRepo) ListMarketTags(ctx context.Context, includeInactive bool) ([]markets.MarketTag, error) {
	if r.listMarketTagsFunc == nil {
		return nil, errUnexpectedMarketsTestCall
	}
	return r.listMarketTagsFunc(ctx, includeInactive)
}

func (r *projectionRepo) CreateMarketTag(ctx context.Context, tag markets.MarketTag) (*markets.MarketTag, error) {
	if r.createMarketTagFunc == nil {
		return nil, errUnexpectedMarketsTestCall
	}
	return r.createMarketTagFunc(ctx, tag)
}

func (r *projectionRepo) UpdateMarketTag(ctx context.Context, slug string, update markets.MarketTagRequest) (*markets.MarketTag, error) {
	if r.updateMarketTagFunc == nil {
		return nil, errUnexpectedMarketsTestCall
	}
	return r.updateMarketTagFunc(ctx, slug, update)
}

func (r *projectionRepo) SetMarketTags(ctx context.Context, marketID int64, tagSlugs []string, assignedBy string, source string, assignedAt time.Time) ([]markets.MarketTag, error) {
	if r.setMarketTagsFunc == nil {
		return nil, errUnexpectedMarketsTestCall
	}
	return r.setMarketTagsFunc(ctx, marketID, tagSlugs, assignedBy, source, assignedAt)
}

func (r *projectionRepo) SetMarketGroupTags(ctx context.Context, groupID int64, tagSlugs []string, assignedBy string, source string, assignedAt time.Time) ([]markets.MarketTag, error) {
	if r.setMarketGroupTagsFunc == nil {
		return nil, errUnexpectedMarketsTestCall
	}
	return r.setMarketGroupTagsFunc(ctx, groupID, tagSlugs, assignedBy, source, assignedAt)
}

func (r *projectionRepo) GetMarketGovernanceSettings(ctx context.Context) (*markets.MarketGovernanceSettings, error) {
	if r.getMarketGovernanceSettingsFunc == nil {
		return nil, errUnexpectedMarketsTestCall
	}
	return r.getMarketGovernanceSettingsFunc(ctx)
}

func (r *projectionRepo) UpdateMarketGovernanceSettings(ctx context.Context, update markets.MarketGovernanceSettingsUpdate) (*markets.MarketGovernanceSettings, error) {
	if r.updateMarketGovernanceSettingsFunc == nil {
		return nil, errUnexpectedMarketsTestCall
	}
	return r.updateMarketGovernanceSettingsFunc(ctx, update)
}

func (r *projectionRepo) ListMarketDiscovery(ctx context.Context, filters markets.ListFilters) (*markets.MarketDiscoveryPage, error) {
	if r.listLifecycleMarketDiscoveryFunc == nil {
		return nil, errUnexpectedMarketsTestCall
	}
	return r.listLifecycleMarketDiscoveryFunc(ctx, filters)
}

func (r *projectionRepo) ListLifecycleMarketDiscovery(ctx context.Context, filters markets.ListFilters) (*markets.MarketDiscoveryPage, error) {
	if r.listLifecycleMarketDiscoveryFunc == nil {
		return nil, errUnexpectedMarketsTestCall
	}
	return r.listLifecycleMarketDiscoveryFunc(ctx, filters)
}

func (r *projectionRepo) ListAdminMarketReviewRows(ctx context.Context, filters markets.AdminMarketReviewFilters) (*markets.AdminMarketReviewPage, error) {
	if r.listAdminMarketReviewRowsFunc == nil {
		return nil, errUnexpectedMarketsTestCall
	}
	return r.listAdminMarketReviewRowsFunc(ctx, filters)
}

func (r *projectionRepo) SearchMarketDiscovery(context.Context, string, markets.SearchFilters) (*markets.MarketDiscoveryPage, error) {
	return nil, errUnexpectedMarketsTestCall
}

func (r *projectionRepo) ListMarketDescriptionAmendmentReviewCandidates(ctx context.Context, filters markets.AdminDescriptionAmendmentReviewFilters) ([]markets.MarketDescriptionAmendment, int, error) {
	if r.listDescriptionAmendmentReviewCandidatesFunc == nil {
		return nil, 0, errUnexpectedMarketsTestCall
	}
	return r.listDescriptionAmendmentReviewCandidatesFunc(ctx, filters)
}

func (r *projectionRepo) ReviewGroupedMarketDescriptionAmendments(ctx context.Context, ids []int64, status string, actorUsername string, reason string, reviewedAt time.Time) ([]markets.MarketDescriptionAmendment, error) {
	if r.reviewGroupedDescriptionAmendmentsFunc == nil {
		return nil, errUnexpectedMarketsTestCall
	}
	return r.reviewGroupedDescriptionAmendmentsFunc(ctx, ids, status, actorUsername, reason, reviewedAt)
}

func (r *projectionRepo) ListMarketGroupAnswerAdditionsForAdminReview(ctx context.Context, filters markets.AdminAnswerAdditionReviewFilters) ([]markets.MarketGroupAnswerAddition, int, error) {
	if r.listAnswerAdditionsForAdminReviewFunc == nil {
		return nil, 0, errUnexpectedMarketsTestCall
	}
	return r.listAnswerAdditionsForAdminReviewFunc(ctx, filters)
}

func (r *projectionRepo) CreateMarketGroup(ctx context.Context, group *markets.MarketGroup, members []markets.MarketGroupMember) error {
	if r.createMarketGroupFunc == nil {
		return errUnexpectedMarketsTestCall
	}
	return r.createMarketGroupFunc(ctx, group, members)
}

func (r *projectionRepo) GetMarketGroup(ctx context.Context, groupID int64) (*markets.MarketGroup, error) {
	if r.getMarketGroupFunc == nil {
		return nil, errUnexpectedMarketsTestCall
	}
	return r.getMarketGroupFunc(ctx, groupID)
}

func (r *projectionRepo) ListMarketGroupMembers(ctx context.Context, groupID int64) ([]markets.MarketGroupMember, error) {
	if r.listMarketGroupMembersFunc == nil {
		return nil, errUnexpectedMarketsTestCall
	}
	return r.listMarketGroupMembersFunc(ctx, groupID)
}

func (r *projectionRepo) GetMarketGroupForMarket(ctx context.Context, marketID int64) (*markets.MarketGroup, error) {
	if r.getMarketGroupForMarketFunc == nil {
		return nil, errUnexpectedMarketsTestCall
	}
	return r.getMarketGroupForMarketFunc(ctx, marketID)
}

func (r *projectionRepo) MarkMarketGroupResolved(ctx context.Context, groupID int64, resolvedAt time.Time) error {
	if r.markMarketGroupResolvedFunc == nil {
		return errUnexpectedMarketsTestCall
	}
	return r.markMarketGroupResolvedFunc(ctx, groupID, resolvedAt)
}

func (r *projectionRepo) UpdateMarketGroupAnswerAdditionAutoApproval(ctx context.Context, groupID int64, enabled bool, updatedAt time.Time) (*markets.MarketGroup, error) {
	if r.updateMarketGroupAnswerAdditionAutoApprovalFunc == nil {
		return nil, errUnexpectedMarketsTestCall
	}
	return r.updateMarketGroupAnswerAdditionAutoApprovalFunc(ctx, groupID, enabled, updatedAt)
}

func (r *projectionRepo) GetUserPosition(ctx context.Context, marketID int64, username string) (*markets.UserPosition, error) {
	if r.getUserPositionFunc == nil {
		return nil, errUnexpectedMarketsTestCall
	}
	return r.getUserPositionFunc(ctx, marketID, username)
}
func (r *projectionRepo) ListMarketPositions(ctx context.Context, marketID int64) (markets.MarketPositions, error) {
	if r.listMarketPositionsFunc == nil {
		return nil, errUnexpectedMarketsTestCall
	}
	return r.listMarketPositionsFunc(ctx, marketID)
}
func (r *projectionRepo) ListBetsForMarket(ctx context.Context, marketID int64) ([]*markets.Bet, error) {
	if r.listBetsForMarketFunc == nil {
		return nil, errUnexpectedMarketsTestCall
	}
	return r.listBetsForMarketFunc(ctx, marketID)
}
func (r *projectionRepo) GetByID(ctx context.Context, id int64) (*markets.Market, error) {
	if r.getByIDFunc == nil {
		return nil, errUnexpectedMarketsTestCall
	}
	return r.getByIDFunc(ctx, id)
}
func (r *projectionRepo) CalculatePayoutPositions(ctx context.Context, marketID int64) ([]*markets.PayoutPosition, error) {
	if r.calculatePayoutPositionsFunc == nil {
		return nil, errUnexpectedMarketsTestCall
	}
	return r.calculatePayoutPositionsFunc(ctx, marketID)
}

func (r *projectionRepo) GetPublicMarket(ctx context.Context, marketID int64) (*markets.PublicMarket, error) {
	if r.getPublicMarketFunc == nil {
		return nil, errUnexpectedMarketsTestCall
	}
	return r.getPublicMarketFunc(ctx, marketID)
}

type projectionClock struct{ nowFunc func() time.Time }

func newProjectionClock(now time.Time) projectionClock {
	return projectionClock{nowFunc: func() time.Time { return now }}
}

func (c projectionClock) Now() time.Time {
	if c.nowFunc == nil {
		return marketsTestTime()
	}
	return c.nowFunc()
}

func TestProjectProbability_ComputesProjection(t *testing.T) {
	createdAt := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
	bets := []*markets.Bet{
		{Username: "alice", MarketID: 55, Amount: 100, Outcome: "YES", PlacedAt: createdAt.Add(5 * time.Minute), CreatedAt: createdAt.Add(5 * time.Minute)},
		{Username: "bob", MarketID: 55, Amount: 100, Outcome: "NO", PlacedAt: createdAt.Add(10 * time.Minute), CreatedAt: createdAt.Add(10 * time.Minute)},
	}
	repo := newProjectionRepo(
		withProjectionRepoMarket(&markets.Market{
			ID:                 55,
			Status:             "active",
			CreatedAt:          createdAt,
			ResolutionDateTime: createdAt.Add(48 * time.Hour),
		}),
		withProjectionRepoBets(bets),
	)

	svc := markets.NewService(repo, nil, newProjectionClock(createdAt.Add(20*time.Minute)), markets.Config{})

	projection, err := svc.ProjectProbability(context.Background(), markets.ProbabilityProjectionRequest{
		MarketID: 55,
		Amount:   50,
		Outcome:  "YES",
	})
	if err != nil {
		t.Fatalf("ProjectProbability returned error: %v", err)
	}

	if projection.CurrentProbability <= 0 || projection.CurrentProbability >= 1 {
		t.Fatalf("unexpected current probability: %v", projection.CurrentProbability)
	}

	expected := wpam.ProjectNewProbabilityWPAM(createdAt, marketsToBoundaryBets(bets), boundaryBet(createdAt.Add(20*time.Minute), 55, 50, "YES"))
	if absDiff(projection.ProjectedProbability, expected.Probability) > 1e-6 {
		t.Fatalf("expected projected %v got %v", expected.Probability, projection.ProjectedProbability)
	}

	if err := (&projectionRepo{}).Create(context.Background(), nil); !errors.Is(err, errUnexpectedMarketsTestCall) {
		t.Fatalf("expected zero-value repo to fail predictably, got %v", err)
	}
}

func TestProjectProbability_InvalidOutcome(t *testing.T) {
	now := marketsTestTime()
	repo := newProjectionRepo(withProjectionRepoMarket(&markets.Market{ID: 1, Status: "active", CreatedAt: now, ResolutionDateTime: now.Add(time.Hour)}))
	svc := markets.NewService(repo, nil, newProjectionClock(now), markets.Config{})

	_, err := svc.ProjectProbability(context.Background(), markets.ProbabilityProjectionRequest{MarketID: 1, Amount: 10, Outcome: "MAYBE"})
	requireInvalidInput(t, err)

	if got := (projectionClock{}).Now(); !got.Equal(marketsTestTime()) {
		t.Fatalf("expected zero-value clock fallback, got %v", got)
	}
}

// helpers for tests

func marketsToBoundaryBets(bets []*markets.Bet) []boundary.Bet {
	return markets.ToBoundaryBets(bets)
}

func boundaryBet(placed time.Time, marketID int64, amount int64, outcome string) boundary.Bet {
	return boundary.Bet{Username: "preview", MarketID: uint(marketID), Amount: amount, Outcome: outcome, PlacedAt: placed}
}

func absDiff(a, b float64) float64 {
	diff := a - b
	if diff < 0 {
		return -diff
	}
	return diff
}
