package markets

import (
	"context"
	"errors"
	"testing"
	"time"

	dmarkets "socialpredict/internal/domain/markets"
	"socialpredict/models"
	"socialpredict/models/modelstesting"
)

func TestGormRepositoryCreateAndGetByID(t *testing.T) {
	db := modelstesting.NewFakeDB(t)
	repo := NewGormRepository(db)
	ctx := context.Background()

	now := time.Now().UTC().Truncate(time.Second)
	market := &dmarkets.Market{
		QuestionTitle:           "Test market",
		Description:             "Description",
		OutcomeType:             "BINARY",
		ResolutionDateTime:      now.Add(24 * time.Hour),
		FinalResolutionDateTime: now.Add(48 * time.Hour),
		ResolutionResult:        "",
		CreatorUsername:         "creator",
		StewardUsername:         "creator",
		YesLabel:                "YES",
		NoLabel:                 "NO",
		Status:                  dmarkets.MarketStatusActive,
		LifecycleStatus:         dmarkets.MarketLifecyclePublished,
		CreatedAt:               now,
		UpdatedAt:               now,
		InitialProbability:      0.5,
		UTCOffset:               -5,
	}

	if err := repo.Create(ctx, market); err != nil {
		t.Fatalf("Create returned error: %v", err)
	}
	if market.ID == 0 {
		t.Fatalf("expected market ID to be set")
	}

	got, err := repo.GetByID(ctx, market.ID)
	if err != nil {
		t.Fatalf("GetByID returned error: %v", err)
	}
	if got.QuestionTitle != market.QuestionTitle || got.CreatorUsername != market.CreatorUsername || got.CurrentStewardUsername() != "creator" || got.YesLabel != "YES" || got.InitialProbability != 0.5 || got.LifecycleStatus != dmarkets.MarketLifecyclePublished {
		t.Fatalf("unexpected market data: %+v", got)
	}

	if _, err := repo.GetByID(ctx, market.ID+999); !errors.Is(err, dmarkets.ErrMarketNotFound) {
		t.Fatalf("expected ErrMarketNotFound, got %v", err)
	}
}

func TestGormRepositoryUpdateLabels(t *testing.T) {
	db := modelstesting.NewFakeDB(t)
	repo := NewGormRepository(db)
	ctx := context.Background()

	seed := modelstesting.GenerateMarket(100, "creator")
	if err := db.Create(&seed).Error; err != nil {
		t.Fatalf("seed market: %v", err)
	}

	if err := repo.UpdateLabels(ctx, seed.ID, "Moon", "Sun"); err != nil {
		t.Fatalf("UpdateLabels returned error: %v", err)
	}

	var refreshed models.Market
	if err := db.First(&refreshed, seed.ID).Error; err != nil {
		t.Fatalf("reload market: %v", err)
	}
	if refreshed.YesLabel != "Moon" || refreshed.NoLabel != "Sun" {
		t.Fatalf("labels not updated: %+v", refreshed)
	}

	if err := repo.UpdateLabels(ctx, seed.ID+1, "A", "B"); !errors.Is(err, dmarkets.ErrMarketNotFound) {
		t.Fatalf("expected ErrMarketNotFound for missing market, got %v", err)
	}
}

func TestGormRepositoryListBetsForMarket(t *testing.T) {
	db := modelstesting.NewFakeDB(t)
	repo := NewGormRepository(db)
	ctx := context.Background()

	creator := modelstesting.GenerateUser("creator", 1000)
	bettor := modelstesting.GenerateUser("bettor", 1000)
	if err := db.Create(&creator).Error; err != nil {
		t.Fatalf("seed creator: %v", err)
	}
	if err := db.Create(&bettor).Error; err != nil {
		t.Fatalf("seed bettor: %v", err)
	}

	market := modelstesting.GenerateMarket(200, creator.Username)
	if err := db.Create(&market).Error; err != nil {
		t.Fatalf("seed market: %v", err)
	}

	first := models.Bet{
		Username: "bettor",
		MarketID: uint(market.ID),
		Amount:   10,
		Outcome:  "YES",
		PlacedAt: time.Now().Add(-2 * time.Minute),
	}
	second := models.Bet{
		Username: "bettor",
		MarketID: uint(market.ID),
		Amount:   15,
		Outcome:  "NO",
		PlacedAt: time.Now().Add(-1 * time.Minute),
	}
	if err := db.Create(&first).Error; err != nil {
		t.Fatalf("insert first bet: %v", err)
	}
	if err := db.Create(&second).Error; err != nil {
		t.Fatalf("insert second bet: %v", err)
	}

	bets, err := repo.ListBetsForMarket(ctx, market.ID)
	if err != nil {
		t.Fatalf("ListBetsForMarket returned error: %v", err)
	}

	if len(bets) != 2 {
		t.Fatalf("expected 2 bets, got %d", len(bets))
	}
	if bets[0].Username != "bettor" || bets[0].Amount != 10 || bets[0].Outcome != "YES" {
		t.Fatalf("unexpected first bet: %+v", bets[0])
	}
	if !bets[0].PlacedAt.Before(bets[1].PlacedAt) {
		t.Fatalf("bets not ordered ascending by PlacedAt")
	}
}

func TestGormRepositoryListHonorsStatusAndCreatorFilter(t *testing.T) {
	db := modelstesting.NewFakeDB(t)
	repo := NewGormRepository(db)
	ctx := context.Background()

	alice := modelstesting.GenerateUser("alice", 1000)
	bob := modelstesting.GenerateUser("bob", 1000)
	if err := db.Create(&alice).Error; err != nil {
		t.Fatalf("seed alice: %v", err)
	}
	if err := db.Create(&bob).Error; err != nil {
		t.Fatalf("seed bob: %v", err)
	}

	now := time.Now().UTC().Truncate(time.Second)
	active := modelstesting.GenerateMarket(300, alice.Username)
	active.ResolutionDateTime = now.Add(24 * time.Hour)
	active.IsResolved = false

	closed := modelstesting.GenerateMarket(301, alice.Username)
	closed.ResolutionDateTime = now.Add(-24 * time.Hour)
	closed.IsResolved = false

	resolved := modelstesting.GenerateMarket(302, alice.Username)
	resolved.ResolutionDateTime = now.Add(-48 * time.Hour)
	resolved.FinalResolutionDateTime = now.Add(-12 * time.Hour)
	resolved.IsResolved = true
	resolved.ResolutionResult = "YES"

	bobsClosed := modelstesting.GenerateMarket(303, bob.Username)
	bobsClosed.ResolutionDateTime = now.Add(-24 * time.Hour)
	bobsClosed.IsResolved = false

	for _, market := range []any{&active, &closed, &resolved, &bobsClosed} {
		if err := db.Create(market).Error; err != nil {
			t.Fatalf("seed market: %v", err)
		}
	}

	tests := []struct {
		name       string
		filters    dmarkets.ListFilters
		wantID     int64
		wantStatus string
	}{
		{
			name:       "active for alice",
			filters:    dmarkets.ListFilters{Status: "active", CreatedBy: alice.Username, Limit: 10},
			wantID:     active.ID,
			wantStatus: "active",
		},
		{
			name:       "closed for alice",
			filters:    dmarkets.ListFilters{Status: "closed", CreatedBy: alice.Username, Limit: 10},
			wantID:     closed.ID,
			wantStatus: "closed",
		},
		{
			name:       "resolved for alice",
			filters:    dmarkets.ListFilters{Status: "resolved", CreatedBy: alice.Username, Limit: 10},
			wantID:     resolved.ID,
			wantStatus: "resolved",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			markets, err := repo.List(ctx, tt.filters)
			if err != nil {
				t.Fatalf("List returned error: %v", err)
			}
			if len(markets) != 1 {
				t.Fatalf("expected 1 market, got %d", len(markets))
			}
			if markets[0].ID != tt.wantID {
				t.Fatalf("expected market ID %d, got %d", tt.wantID, markets[0].ID)
			}
			if markets[0].Status != tt.wantStatus {
				t.Fatalf("expected status %q, got %q", tt.wantStatus, markets[0].Status)
			}
		})
	}
}

func TestGormRepositoryPublicSearchStatusResolveAndDelete(t *testing.T) {
	db := modelstesting.NewFakeDB(t)
	repo := NewGormRepository(db)
	ctx := context.Background()

	creator := modelstesting.GenerateUser("creator", 1000)
	if err := db.Create(&creator).Error; err != nil {
		t.Fatalf("seed creator: %v", err)
	}

	now := time.Now().UTC().Truncate(time.Second)
	active := modelstesting.GenerateMarket(400, creator.Username)
	active.QuestionTitle = "Will oranges rally?"
	active.Description = "Citrus market"
	active.ResolutionDateTime = now.Add(24 * time.Hour)
	active.IsResolved = false

	closed := modelstesting.GenerateMarket(401, creator.Username)
	closed.QuestionTitle = "Will apples close?"
	closed.Description = "Orchard market"
	closed.ResolutionDateTime = now.Add(-24 * time.Hour)
	closed.IsResolved = false

	resolved := modelstesting.GenerateMarket(402, creator.Username)
	resolved.QuestionTitle = "Will pears resolve?"
	resolved.Description = "Resolved orchard market"
	resolved.ResolutionDateTime = now.Add(-48 * time.Hour)
	resolved.IsResolved = true
	resolved.ResolutionResult = "NO"

	for _, market := range []*models.Market{&active, &closed, &resolved} {
		if err := db.Create(market).Error; err != nil {
			t.Fatalf("seed market %d: %v", market.ID, err)
		}
	}

	publicMarket, err := repo.GetPublicMarket(ctx, active.ID)
	if err != nil {
		t.Fatalf("GetPublicMarket returned error: %v", err)
	}
	if publicMarket.ID != active.ID || publicMarket.QuestionTitle != active.QuestionTitle || publicMarket.CreatorUsername != creator.Username {
		t.Fatalf("unexpected public market: %+v", publicMarket)
	}
	if _, err := repo.GetPublicMarket(ctx, 9999); !errors.Is(err, dmarkets.ErrMarketNotFound) {
		t.Fatalf("expected ErrMarketNotFound for missing public market, got %v", err)
	}

	searchResults, err := repo.Search(ctx, "orchard", dmarkets.SearchFilters{Status: "closed", Limit: 10})
	if err != nil {
		t.Fatalf("Search returned error: %v", err)
	}
	if len(searchResults) != 1 || searchResults[0].ID != closed.ID {
		t.Fatalf("unexpected closed search results: %+v", searchResults)
	}

	resolvedResults, err := repo.ListByStatus(ctx, "resolved", dmarkets.Page{Limit: 10})
	if err != nil {
		t.Fatalf("ListByStatus returned error: %v", err)
	}
	if len(resolvedResults) != 1 || resolvedResults[0].ID != resolved.ID || resolvedResults[0].Status != "resolved" {
		t.Fatalf("unexpected resolved results: %+v", resolvedResults)
	}

	if err := repo.ResolveMarket(ctx, active.ID, "YES"); err != nil {
		t.Fatalf("ResolveMarket returned error: %v", err)
	}
	refreshed, err := repo.GetByID(ctx, active.ID)
	if err != nil {
		t.Fatalf("GetByID after resolve returned error: %v", err)
	}
	if refreshed.Status != "resolved" || refreshed.ResolutionResult != "YES" {
		t.Fatalf("unexpected resolved market: %+v", refreshed)
	}
	if err := repo.ResolveMarket(ctx, 9999, "YES"); !errors.Is(err, dmarkets.ErrMarketNotFound) {
		t.Fatalf("expected ErrMarketNotFound for missing resolve, got %v", err)
	}

	if err := repo.Delete(ctx, closed.ID); err != nil {
		t.Fatalf("Delete returned error: %v", err)
	}
	if _, err := repo.GetByID(ctx, closed.ID); !errors.Is(err, dmarkets.ErrMarketNotFound) {
		t.Fatalf("expected deleted market to be missing, got %v", err)
	}
	if err := repo.Delete(ctx, closed.ID); !errors.Is(err, dmarkets.ErrMarketNotFound) {
		t.Fatalf("expected ErrMarketNotFound for repeated delete, got %v", err)
	}
}

func TestGormRepositoryHidesNonPublicLifecycleMarketsFromPublicQueries(t *testing.T) {
	db := modelstesting.NewFakeDB(t)
	repo := NewGormRepository(db)
	ctx := context.Background()

	creator := modelstesting.GenerateUser("lifecycle_creator", 1000)
	if err := db.Create(&creator).Error; err != nil {
		t.Fatalf("seed creator: %v", err)
	}

	now := time.Now().UTC().Truncate(time.Second)
	published := modelstesting.GenerateMarket(501, creator.Username)
	published.QuestionTitle = "Published lifecycle market"
	published.ResolutionDateTime = now.Add(24 * time.Hour)
	published.LifecycleStatus = dmarkets.MarketLifecyclePublished

	proposed := modelstesting.GenerateMarket(502, creator.Username)
	proposed.QuestionTitle = "Proposed lifecycle market"
	proposed.ResolutionDateTime = now.Add(24 * time.Hour)
	proposed.LifecycleStatus = dmarkets.MarketLifecycleProposed

	rejected := modelstesting.GenerateMarket(503, creator.Username)
	rejected.QuestionTitle = "Rejected lifecycle market"
	rejected.ResolutionDateTime = now.Add(24 * time.Hour)
	rejected.LifecycleStatus = dmarkets.MarketLifecycleRejected

	cancelled := modelstesting.GenerateMarket(504, creator.Username)
	cancelled.QuestionTitle = "Cancelled lifecycle market"
	cancelled.ResolutionDateTime = now.Add(24 * time.Hour)
	cancelled.LifecycleStatus = dmarkets.MarketLifecycleCancelled

	for _, market := range []*models.Market{&published, &proposed, &rejected, &cancelled} {
		if err := db.Create(market).Error; err != nil {
			t.Fatalf("seed market %d: %v", market.ID, err)
		}
	}

	listed, err := repo.List(ctx, dmarkets.ListFilters{Limit: 10})
	if err != nil {
		t.Fatalf("List returned error: %v", err)
	}
	if len(listed) != 1 || listed[0].ID != published.ID {
		t.Fatalf("expected only published market in public list, got %+v", listed)
	}

	searched, err := repo.Search(ctx, "lifecycle", dmarkets.SearchFilters{Limit: 10})
	if err != nil {
		t.Fatalf("Search returned error: %v", err)
	}
	if len(searched) != 1 || searched[0].ID != published.ID {
		t.Fatalf("expected only published market in public search, got %+v", searched)
	}

	if _, err := repo.GetPublicMarket(ctx, proposed.ID); !errors.Is(err, dmarkets.ErrMarketNotFound) {
		t.Fatalf("expected proposed market to be hidden from public get, got %v", err)
	}
}

func TestGormRepositoryApproveAndRejectMarketPersistReviewMetadata(t *testing.T) {
	db := modelstesting.NewFakeDB(t)
	repo := NewGormRepository(db)
	ctx := context.Background()

	creator := modelstesting.GenerateUser("review_creator", 1000)
	if err := db.Create(&creator).Error; err != nil {
		t.Fatalf("seed creator: %v", err)
	}

	approveTarget := modelstesting.GenerateMarket(601, creator.Username)
	approveTarget.LifecycleStatus = dmarkets.MarketLifecycleProposed
	rejectTarget := modelstesting.GenerateMarket(602, creator.Username)
	rejectTarget.LifecycleStatus = dmarkets.MarketLifecycleProposed
	rejectTarget.ProposalCost = 7
	published := modelstesting.GenerateMarket(603, creator.Username)
	published.LifecycleStatus = dmarkets.MarketLifecyclePublished
	for _, market := range []*models.Market{&approveTarget, &rejectTarget, &published} {
		if err := db.Create(market).Error; err != nil {
			t.Fatalf("seed market %d: %v", market.ID, err)
		}
	}

	approvedAt := time.Date(2026, 5, 24, 10, 0, 0, 0, time.UTC)
	if err := repo.ApproveMarket(ctx, approveTarget.ID, "admin", approvedAt); err != nil {
		t.Fatalf("ApproveMarket returned error: %v", err)
	}
	approved, err := repo.GetByID(ctx, approveTarget.ID)
	if err != nil {
		t.Fatalf("load approved market: %v", err)
	}
	if approved.LifecycleStatus != dmarkets.MarketLifecyclePublished || approved.ApprovedBy != "admin" || approved.ApprovedAt == nil || !approved.ApprovedAt.Equal(approvedAt) {
		t.Fatalf("unexpected approved market: %+v", approved)
	}

	rejectedAt := approvedAt.Add(time.Hour)
	if err := repo.RejectMarket(ctx, rejectTarget.ID, "admin", rejectedAt, "duplicate"); err != nil {
		t.Fatalf("RejectMarket returned error: %v", err)
	}
	rejected, err := repo.GetByID(ctx, rejectTarget.ID)
	if err != nil {
		t.Fatalf("load rejected market: %v", err)
	}
	if rejected.LifecycleStatus != dmarkets.MarketLifecycleRejected || rejected.RejectedBy != "admin" || rejected.RejectionReason != "duplicate" || rejected.ProposalCost != 7 || rejected.RejectedAt == nil || !rejected.RejectedAt.Equal(rejectedAt) {
		t.Fatalf("unexpected rejected market: %+v", rejected)
	}

	if err := repo.ApproveMarket(ctx, published.ID, "admin", approvedAt); !errors.Is(err, dmarkets.ErrInvalidState) {
		t.Fatalf("ApproveMarket wrong state error = %v, want ErrInvalidState", err)
	}
}

func TestGormRepositoryReassignMarketStewardPersistsAudit(t *testing.T) {
	db := modelstesting.NewFakeDB(t)
	repo := NewGormRepository(db)
	ctx := context.Background()

	creator := modelstesting.GenerateUser("steward_creator", 1000)
	backup := modelstesting.GenerateUser("steward_backup", 1000)
	if err := db.Create(&creator).Error; err != nil {
		t.Fatalf("seed creator: %v", err)
	}
	if err := db.Create(&backup).Error; err != nil {
		t.Fatalf("seed backup: %v", err)
	}

	market := modelstesting.GenerateMarket(701, creator.Username)
	market.StewardUsername = creator.Username
	if err := db.Create(&market).Error; err != nil {
		t.Fatalf("seed market: %v", err)
	}

	changedAt := time.Date(2026, 6, 3, 10, 0, 0, 0, time.UTC)
	if err := repo.ReassignMarketSteward(ctx, market.ID, creator.Username, backup.Username, "admin", "creator inactive", changedAt); err != nil {
		t.Fatalf("ReassignMarketSteward returned error: %v", err)
	}

	updated, err := repo.GetByID(ctx, market.ID)
	if err != nil {
		t.Fatalf("load market: %v", err)
	}
	if updated.CurrentStewardUsername() != backup.Username {
		t.Fatalf("steward = %q, want %q", updated.CurrentStewardUsername(), backup.Username)
	}

	var audits []models.MarketStewardshipAudit
	if err := db.Where("market_id = ?", market.ID).Find(&audits).Error; err != nil {
		t.Fatalf("load audits: %v", err)
	}
	if len(audits) != 1 {
		t.Fatalf("expected one audit, got %d", len(audits))
	}
	audit := audits[0]
	if audit.FromStewardUsername != creator.Username || audit.ToStewardUsername != backup.Username || audit.ActorUsername != "admin" || audit.Reason != "creator inactive" {
		t.Fatalf("unexpected audit: %+v", audit)
	}

	if err := repo.ReassignMarketSteward(ctx, market.ID+999, creator.Username, backup.Username, "admin", "missing", changedAt); !errors.Is(err, dmarkets.ErrMarketNotFound) {
		t.Fatalf("missing market error = %v, want ErrMarketNotFound", err)
	}
}

func TestGormRepositoryListByLifecycleHydratesStewardshipAudits(t *testing.T) {
	db := modelstesting.NewFakeDB(t)
	repo := NewGormRepository(db)
	ctx := context.Background()

	creator := modelstesting.GenerateUser("lifecycle_steward_creator", 1000)
	backup := modelstesting.GenerateUser("lifecycle_steward_backup", 1000)
	if err := db.Create(&creator).Error; err != nil {
		t.Fatalf("seed creator: %v", err)
	}
	if err := db.Create(&backup).Error; err != nil {
		t.Fatalf("seed backup: %v", err)
	}

	market := modelstesting.GenerateMarket(702, creator.Username)
	market.LifecycleStatus = dmarkets.MarketLifecyclePublished
	market.StewardUsername = backup.Username
	if err := db.Create(&market).Error; err != nil {
		t.Fatalf("seed market: %v", err)
	}

	changedAt := time.Date(2026, 6, 4, 15, 0, 0, 0, time.UTC)
	audit := models.MarketStewardshipAudit{
		MarketID:            market.ID,
		FromStewardUsername: creator.Username,
		ToStewardUsername:   backup.Username,
		ActorUsername:       "admin",
		Reason:              "creator suspended",
	}
	audit.CreatedAt = changedAt
	audit.UpdatedAt = changedAt
	if err := db.Create(&audit).Error; err != nil {
		t.Fatalf("seed audit: %v", err)
	}

	markets, err := repo.ListByLifecycle(ctx, dmarkets.ListFilters{Status: dmarkets.MarketLifecyclePublished, Limit: 10})
	if err != nil {
		t.Fatalf("ListByLifecycle returned error: %v", err)
	}
	if len(markets) != 1 {
		t.Fatalf("expected one market, got %d", len(markets))
	}
	audits := markets[0].StewardshipAudits
	if len(audits) != 1 || audits[0].Reason != "creator suspended" || audits[0].FromStewardUsername != creator.Username || audits[0].ToStewardUsername != backup.Username || !audits[0].CreatedAt.Equal(changedAt) {
		t.Fatalf("unexpected hydrated audits: %+v", audits)
	}
}

func TestGormRepositoryListByLifecycleAllExcludesRejectedAndSearchesOperationalMarkets(t *testing.T) {
	db := modelstesting.NewFakeDB(t)
	repo := NewGormRepository(db)
	ctx := context.Background()

	creator := modelstesting.GenerateUser("lifecycle_search_creator", 1000)
	if err := db.Create(&creator).Error; err != nil {
		t.Fatalf("seed creator: %v", err)
	}

	published := modelstesting.GenerateMarket(703, creator.Username)
	published.QuestionTitle = "Orchard published market"
	published.LifecycleStatus = dmarkets.MarketLifecyclePublished
	rejected := modelstesting.GenerateMarket(704, creator.Username)
	rejected.QuestionTitle = "Orchard rejected market"
	rejected.LifecycleStatus = dmarkets.MarketLifecycleRejected
	resolved := modelstesting.GenerateMarket(705, creator.Username)
	resolved.QuestionTitle = "Orchard resolved market"
	resolved.LifecycleStatus = dmarkets.MarketLifecycleResolved
	resolved.IsResolved = true

	for _, market := range []models.Market{published, rejected, resolved} {
		if err := db.Create(&market).Error; err != nil {
			t.Fatalf("seed market %q: %v", market.QuestionTitle, err)
		}
	}

	markets, err := repo.ListByLifecycle(ctx, dmarkets.ListFilters{
		Status: dmarkets.MarketStatusAll,
		Query:  "orchard",
		Limit:  10,
	})
	if err != nil {
		t.Fatalf("ListByLifecycle returned error: %v", err)
	}
	if len(markets) != 2 {
		t.Fatalf("expected published and resolved markets only, got %d: %+v", len(markets), markets)
	}
	for _, market := range markets {
		if market.LifecycleStatus == dmarkets.MarketLifecycleRejected {
			t.Fatalf("rejected market should not be included in all stewardship list: %+v", market)
		}
	}
}
