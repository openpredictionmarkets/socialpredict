package marketshandlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
	"time"

	"socialpredict/handlers"
	"socialpredict/handlers/markets/dto"
	dmarkets "socialpredict/internal/domain/markets"
	dusers "socialpredict/internal/domain/users"

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

func TestResolveMarketGroupHandlerAcceptsManualChildResolutions(t *testing.T) {
	var captured dmarkets.MarketGroupResolveRequest
	service := &MockService{
		ResolveMarketGroupFn: func(_ context.Context, groupID int64, req dmarkets.MarketGroupResolveRequest, username string) (*dmarkets.MarketGroup, error) {
			if groupID != 9 || username != "moderator" {
				t.Fatalf("unexpected args groupID=%d username=%q", groupID, username)
			}
			captured = req
			return &dmarkets.MarketGroup{
				ID:              groupID,
				QuestionTitle:   "Grouped market",
				LifecycleStatus: dmarkets.MarketLifecycleResolved,
				Members: []dmarkets.MarketGroupMember{
					{MarketID: 101, AnswerLabel: "Home", DisplayOrder: 0},
					{MarketID: 102, AnswerLabel: "Away", DisplayOrder: 1},
				},
			}, nil
		},
	}
	req := mux.SetURLVars(
		httptest.NewRequest(
			http.MethodPost,
			"/v0/market-groups/9/resolve",
			strings.NewReader(`{"mode":"manual","resolutions":[{"marketId":101,"resolution":"YES"},{"marketId":102,"resolution":"NO"}]}`),
		),
		map[string]string{"id": "9"},
	)
	rec := httptest.NewRecorder()

	NewHandler(service, lifecycleAuthMock{
		user: &dusers.User{Username: "moderator", UserType: string(dusers.UserTypeModerator)},
	}, nil).ResolveMarketGroup(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d body=%s", http.StatusOK, rec.Code, rec.Body.String())
	}
	expected := dmarkets.MarketGroupResolveRequest{
		Mode: "manual",
		Resolutions: []dmarkets.MarketGroupChildResolution{
			{MarketID: 101, Resolution: "YES"},
			{MarketID: 102, Resolution: "NO"},
		},
	}
	if !reflect.DeepEqual(captured, expected) {
		t.Fatalf("captured = %+v, want %+v", captured, expected)
	}
}

func TestResolveMarketGroupHandlerReportsUnpublishedChild(t *testing.T) {
	service := &MockService{
		ResolveMarketGroupFn: func(_ context.Context, groupID int64, req dmarkets.MarketGroupResolveRequest, username string) (*dmarkets.MarketGroup, error) {
			if groupID != 9 || username != "moderator" {
				t.Fatalf("unexpected args groupID=%d username=%q", groupID, username)
			}
			return nil, &dmarkets.MarketGroupChildNotPublishedError{
				MarketID:        102,
				AnswerLabel:     "Away",
				LifecycleStatus: dmarkets.MarketLifecycleProposed,
			}
		},
	}
	req := mux.SetURLVars(
		httptest.NewRequest(
			http.MethodPost,
			"/v0/market-groups/9/resolve",
			strings.NewReader(`{"mode":"exclusive_yes","winningMarketId":101}`),
		),
		map[string]string{"id": "9"},
	)
	rec := httptest.NewRecorder()

	NewHandler(service, lifecycleAuthMock{
		user: &dusers.User{Username: "moderator", UserType: string(dusers.UserTypeModerator)},
	}, nil).ResolveMarketGroup(rec, req)

	if rec.Code != http.StatusConflict {
		t.Fatalf("expected status %d, got %d body=%s", http.StatusConflict, rec.Code, rec.Body.String())
	}
	var response handlers.FailureEnvelope
	if err := json.NewDecoder(rec.Body).Decode(&response); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if response.Reason != string(handlers.ReasonMarketGroupChildUnpublished) {
		t.Fatalf("reason = %q, want %q", response.Reason, handlers.ReasonMarketGroupChildUnpublished)
	}
	if response.Message == "" || !strings.Contains(response.Message, "Away") || !strings.Contains(response.Message, "not published") {
		t.Fatalf("unexpected message: %q", response.Message)
	}
	if response.Details["marketId"] != float64(102) ||
		response.Details["answerLabel"] != "Away" ||
		response.Details["lifecycleStatus"] != dmarkets.MarketLifecycleProposed {
		t.Fatalf("unexpected details: %+v", response.Details)
	}
}

func TestMarketGroupBetsHandlerReturnsGroupedBetRows(t *testing.T) {
	now := time.Date(2026, 6, 13, 12, 0, 0, 0, time.UTC)
	service := &MockService{
		MarketGroupBetsFn: func(_ context.Context, groupID int64, p dmarkets.Page) (*dmarkets.MarketGroupBetsPage, error) {
			if groupID != 9 {
				t.Fatalf("expected group id 9, got %d", groupID)
			}
			if p.Limit != 21 || p.Offset != 20 {
				t.Fatalf("expected page limit=21 offset=20, got %+v", p)
			}
			return &dmarkets.MarketGroupBetsPage{
				GroupID: groupID,
				Total:   22,
				Bets: []*dmarkets.MarketGroupBetDisplayInfo{{
					AnswerMarketID: 101,
					AnswerLabel:    "Spain",
					DisplayOrder:   0,
					Username:       "alice",
					Outcome:        "YES",
					Amount:         10,
					Probability:    0.62,
					PlacedAt:       now,
				}},
			}, nil
		},
	}
	req := mux.SetURLVars(httptest.NewRequest(http.MethodGet, "/v0/market-groups/9/bets?limit=21&offset=20", nil), map[string]string{"id": "9"})
	rec := httptest.NewRecorder()

	NewHandler(service, nil, nil).MarketGroupBets(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d body=%s", http.StatusOK, rec.Code, rec.Body.String())
	}
	var response handlers.SuccessEnvelope[dto.MarketGroupBetsResponse]
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if !response.OK || response.Result.Total != 22 || len(response.Result.Bets) != 1 {
		t.Fatalf("unexpected response: %+v", response)
	}
	if response.Result.Freshness == nil ||
		response.Result.Freshness.Source != "live" ||
		response.Result.Freshness.TransactionSafeRead ||
		response.Result.Freshness.TargetFreshnessSeconds != 0 {
		t.Fatalf("unexpected freshness metadata: %+v", response.Result.Freshness)
	}
	if got := response.Result.Bets[0].AnswerLabel; got != "Spain" {
		t.Fatalf("expected answer label Spain, got %q", got)
	}
}

func TestMarketGroupPositionsHandlerReturnsGroupedPositionRows(t *testing.T) {
	service := &MockService{
		MarketGroupPositionsFn: func(_ context.Context, groupID int64, p dmarkets.Page) (*dmarkets.MarketGroupPositionsPage, error) {
			if groupID != 9 {
				t.Fatalf("expected group id 9, got %d", groupID)
			}
			return &dmarkets.MarketGroupPositionsPage{
				GroupID: groupID,
				Total:   1,
				Positions: []*dmarkets.MarketGroupPositionRow{{
					Username:       "alice",
					YesSharesOwned: 7,
					Value:          12,
					Answers: []*dmarkets.MarketGroupPositionAnswer{{
						AnswerMarketID: 101,
						AnswerLabel:    "Spain",
						DisplayOrder:   0,
						YesSharesOwned: 7,
						Value:          12,
					}},
				}},
			}, nil
		},
	}
	req := mux.SetURLVars(httptest.NewRequest(http.MethodGet, "/v0/market-groups/9/positions", nil), map[string]string{"id": "9"})
	rec := httptest.NewRecorder()

	NewHandler(service, nil, nil).MarketGroupPositions(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d body=%s", http.StatusOK, rec.Code, rec.Body.String())
	}
	var response handlers.SuccessEnvelope[dto.MarketGroupPositionsResponse]
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if !response.OK || len(response.Result.Positions) != 1 || len(response.Result.Positions[0].Answers) != 1 {
		t.Fatalf("unexpected response: %+v", response)
	}
	if response.Result.Freshness == nil ||
		response.Result.Freshness.Source != "live" ||
		response.Result.Freshness.TransactionSafeRead ||
		response.Result.Freshness.TargetFreshnessSeconds != 0 {
		t.Fatalf("unexpected freshness metadata: %+v", response.Result.Freshness)
	}
}

func TestMarketGroupLeaderboardHandlerReturnsGroupedLeaderboardRows(t *testing.T) {
	service := &MockService{
		MarketGroupLeaderboardFn: func(_ context.Context, groupID int64, p dmarkets.Page) (*dmarkets.MarketGroupLeaderboardPage, error) {
			if groupID != 9 {
				t.Fatalf("expected group id 9, got %d", groupID)
			}
			return &dmarkets.MarketGroupLeaderboardPage{
				GroupID: groupID,
				Total:   1,
				Leaderboard: []*dmarkets.MarketGroupLeaderboardRow{{
					Username:       "alice",
					Profit:         5,
					CurrentValue:   20,
					TotalSpent:     15,
					Position:       "YES",
					YesSharesOwned: 4,
					Rank:           1,
					Answers: []*dmarkets.MarketGroupLeaderboardAnswer{{
						AnswerMarketID: 101,
						AnswerLabel:    "Spain",
						Profit:         5,
					}},
				}},
			}, nil
		},
	}
	req := mux.SetURLVars(httptest.NewRequest(http.MethodGet, "/v0/market-groups/9/leaderboard", nil), map[string]string{"id": "9"})
	rec := httptest.NewRecorder()

	NewHandler(service, nil, nil).MarketGroupLeaderboard(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d body=%s", http.StatusOK, rec.Code, rec.Body.String())
	}
	var response handlers.SuccessEnvelope[dto.MarketGroupLeaderboardResponse]
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if !response.OK || len(response.Result.Leaderboard) != 1 || len(response.Result.Leaderboard[0].Answers) != 1 {
		t.Fatalf("unexpected response: %+v", response)
	}
	if response.Result.Freshness == nil ||
		response.Result.Freshness.Source != "live" ||
		response.Result.Freshness.TransactionSafeRead ||
		response.Result.Freshness.TargetFreshnessSeconds != 0 {
		t.Fatalf("unexpected freshness metadata: %+v", response.Result.Freshness)
	}
}
