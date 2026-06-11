package marketshandlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"socialpredict/handlers"
	dmarkets "socialpredict/internal/domain/markets"
	readmodelrepo "socialpredict/internal/repository/readmodels"

	"github.com/gorilla/mux"
)

func TestMarketSummaryReadModelUsesSnapshotBackedService(t *testing.T) {
	generatedAt := time.Date(2026, 6, 11, 12, 0, 0, 0, time.UTC)
	svc := &MockService{
		MarketSummaryFn: func(_ctx context.Context, marketID int64) (*dmarkets.MarketSummaryReadModel, error) {
			return &dmarkets.MarketSummaryReadModel{
				Market: &dmarkets.Market{
					ID:                 marketID,
					QuestionTitle:      "Cached Summary Market",
					Description:        "summary",
					OutcomeType:        "BINARY",
					ResolutionDateTime: generatedAt.Add(24 * time.Hour),
					CreatorUsername:    "creator",
					YesLabel:           "YES",
					NoLabel:            "NO",
					Status:             dmarkets.MarketStatusActive,
					CreatedAt:          generatedAt,
					UpdatedAt:          generatedAt,
				},
				Creator: &dmarkets.CreatorSummary{Username: "creator", DisplayName: "Creator"},
				Accounting: dmarkets.MarketAccountingSnapshot{
					MarketID:           marketID,
					GeneratedAt:        generatedAt,
					ProbabilityChanges: []dmarkets.ProbabilityPoint{{Probability: 0.7, Timestamp: generatedAt}},
					LastProbability:    0.7,
					UserCount:          4,
					VolumeWithDust:     120,
					MarketDust:         1,
					Source:             "read_model",
					IsStale:            true,
					StaleReason:        "bet_accepted",
				},
			}, nil
		},
	}
	handler := NewHandler(svc, nil, nil)
	req := httptest.NewRequest(http.MethodGet, "/v0/read/markets/7/summary", nil)
	req = mux.SetURLVars(req, map[string]string{"id": "7"})
	rec := httptest.NewRecorder()

	handler.MarketSummaryReadModel(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", rec.Code, rec.Body.String())
	}
	if svc.DetailsCalls != 0 {
		t.Fatalf("GetMarketDetails was called %d times; want 0", svc.DetailsCalls)
	}

	var envelope handlers.SuccessEnvelope[marketSummaryReadModelResponse]
	if err := json.Unmarshal(rec.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if !envelope.OK || envelope.Result.TotalVolume != 120 || envelope.Result.NumUsers != 4 {
		t.Fatalf("unexpected summary response: %+v", envelope)
	}
	if !envelope.Result.Freshness.IsStale || envelope.Result.Freshness.StaleReason != "bet_accepted" {
		t.Fatalf("expected stale freshness from snapshot, got %+v", envelope.Result.Freshness)
	}
	if len(envelope.Result.Probability) != 1 || envelope.Result.Probability[0].Probability != 0.7 {
		t.Fatalf("unexpected probability response: %+v", envelope.Result.Probability)
	}
}

func TestMarketOverviewResponsesPreferSummaryReadModel(t *testing.T) {
	svc := &MockService{}
	markets := []*dmarkets.Market{{ID: 1}, {ID: 2}}

	responses, err := buildMarketOverviewResponses(context.Background(), svc, markets)
	if err != nil {
		t.Fatalf("buildMarketOverviewResponses returned error: %v", err)
	}
	if len(responses) != 2 {
		t.Fatalf("expected 2 responses, got %d", len(responses))
	}
	if svc.DetailsCalls != 0 {
		t.Fatalf("GetMarketDetails was called %d times; want 0", svc.DetailsCalls)
	}
	if responses[0].TotalVolume != 1000 || responses[0].MarketDust != 50 {
		t.Fatalf("expected summary read-model values, got %+v", responses[0])
	}
}

func TestMarketSummaryToDetailsResponseDoesNotIncludeAmendments(t *testing.T) {
	generatedAt := time.Date(2026, 6, 11, 12, 0, 0, 0, time.UTC)
	response := marketSummaryToDetailsResponse(&dmarkets.MarketSummaryReadModel{
		Market: &dmarkets.Market{ID: 3, QuestionTitle: "Pinned", CreatorUsername: "creator", CreatedAt: generatedAt},
		Creator: &dmarkets.CreatorSummary{
			Username: "creator",
		},
		Accounting: dmarkets.MarketAccountingSnapshot{
			MarketID:           3,
			GeneratedAt:        generatedAt,
			ProbabilityChanges: []dmarkets.ProbabilityPoint{{Probability: 0.6, Timestamp: generatedAt}},
			UserCount:          2,
			VolumeWithDust:     30,
			MarketDust:         1,
		},
	})

	if response.Market.ID != 3 || response.TotalVolume != 30 || response.MarketDust != 1 {
		t.Fatalf("unexpected details response: %+v", response)
	}
	if len(response.DescriptionAmendments) != 0 {
		t.Fatalf("summary read model should not hydrate amendments, got %+v", response.DescriptionAmendments)
	}
}

func TestMarketDiscoverySnapshotUsableAllowsYoungStaleSnapshot(t *testing.T) {
	now := time.Date(2026, 6, 11, 12, 0, 0, 0, time.UTC)
	handler := &MarketDiscoveryReadModelHandler{clock: func() time.Time { return now }}
	snapshot := &readmodelrepo.Snapshot{
		PayloadJSON: `{"markets":[]}`,
		GeneratedAt: now.Add(-1 * time.Minute),
		IsStale:     true,
		StaleReason: "bet_accepted",
	}

	if !handler.snapshotUsable(snapshot) {
		t.Fatalf("expected young stale discovery snapshot to be usable")
	}

	snapshot.GeneratedAt = now.Add(-marketDiscoverySnapshotTargetFreshness - time.Second)
	if handler.snapshotUsable(snapshot) {
		t.Fatalf("expected expired discovery snapshot to be unusable")
	}
}
