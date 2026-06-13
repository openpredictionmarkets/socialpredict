package marketshandlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"socialpredict/handlers/markets/dto"
	dmarkets "socialpredict/internal/domain/markets"

	"github.com/gorilla/mux"
)

func TestGetMarketGroupIncludesChildProbabilityHistoryAndAmendments(t *testing.T) {
	now := time.Date(2026, 6, 12, 12, 0, 0, 0, time.UTC)
	service := &MockService{
		MarketGroupOverviewFn: func(_ context.Context, groupID int64) (*dmarkets.MarketGroupOverview, error) {
			if groupID != 9 {
				t.Fatalf("expected group id 9, got %d", groupID)
			}
			return &dmarkets.MarketGroupOverview{
				Group: &dmarkets.MarketGroup{
					ID:                 9,
					QuestionTitle:      "Spain vs Canada",
					Description:        "Grouped market",
					GroupType:          dmarkets.MarketGroupTypeMultipleChoiceBinary,
					ProbabilityPolicy:  dmarkets.MarketGroupProbabilityPolicyIndependentBinary,
					ResolutionPolicy:   dmarkets.MarketGroupResolutionPolicyIndependentChildren,
					LifecycleStatus:    dmarkets.MarketLifecyclePublished,
					CreatorUsername:    "moderator",
					StewardUsername:    "moderator",
					ResolutionDateTime: now.Add(24 * time.Hour),
					CreatedAt:          now,
					UpdatedAt:          now,
					Members: []dmarkets.MarketGroupMember{
						{ID: 1, GroupID: 9, MarketID: 101, AnswerLabel: "Spain", DisplayOrder: 0},
					},
				},
				Creator: &dmarkets.CreatorSummary{Username: "moderator"},
				Answers: []dmarkets.MarketGroupAnswerOverview{
					{
						Member: dmarkets.MarketGroupMember{ID: 1, GroupID: 9, MarketID: 101, AnswerLabel: "Spain", DisplayOrder: 0},
						Overview: &dmarkets.MarketOverview{
							Market: &dmarkets.Market{
								ID:                 101,
								QuestionTitle:      "Spain vs Canada - Spain",
								Description:        "Child market",
								OutcomeType:        "BINARY",
								ResolutionDateTime: now.Add(24 * time.Hour),
								CreatorUsername:    "moderator",
								StewardUsername:    "moderator",
								YesLabel:           "YES",
								NoLabel:            "NO",
								Status:             dmarkets.MarketStatusActive,
								LifecycleStatus:    dmarkets.MarketLifecyclePublished,
								CreatedAt:          now,
								UpdatedAt:          now,
								InitialProbability: 0.5,
							},
							Creator:         &dmarkets.CreatorSummary{Username: "moderator"},
							LastProbability: 0.64,
							ProbabilityChanges: []dmarkets.ProbabilityPoint{
								{Probability: 0.50, Timestamp: now},
								{Probability: 0.64, Timestamp: now.Add(time.Minute)},
							},
							DescriptionAmendments: []dmarkets.MarketDescriptionAmendment{
								{
									ID:         77,
									MarketID:   101,
									Version:    2,
									Body:       "Clarifies resolution source.",
									BodyFormat: dmarkets.DescriptionAmendmentFormatMarkdownLite,
									Status:     dmarkets.DescriptionAmendmentStatusApproved,
									CreatedBy:  "moderator",
									CreatedAt:  now,
									UpdatedAt:  now,
								},
							},
						},
					},
				},
			}, nil
		},
	}

	req := mux.SetURLVars(httptest.NewRequest(http.MethodGet, "/v0/market-groups/9", nil), map[string]string{"id": "9"})
	rec := httptest.NewRecorder()

	NewHandler(service, nil, nil).GetMarketGroup(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d body=%s", http.StatusOK, rec.Code, rec.Body.String())
	}

	var response dto.MarketGroupDetailsResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(response.Answers) != 1 {
		t.Fatalf("expected 1 answer, got %d", len(response.Answers))
	}
	answer := response.Answers[0]
	if len(answer.ProbabilityChanges) != 2 {
		t.Fatalf("expected 2 probability points, got %d", len(answer.ProbabilityChanges))
	}
	if got := answer.ProbabilityChanges[1].Probability; got != 0.64 {
		t.Fatalf("expected latest probability 0.64, got %v", got)
	}
	if len(answer.DescriptionAmendments) != 1 {
		t.Fatalf("expected 1 description amendment, got %d", len(answer.DescriptionAmendments))
	}
	if got := answer.DescriptionAmendments[0].Body; got != "Clarifies resolution source." {
		t.Fatalf("expected amendment body to round trip, got %q", got)
	}
}
