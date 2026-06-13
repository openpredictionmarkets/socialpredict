package marketshandlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"socialpredict/handlers"
	"socialpredict/handlers/markets/dto"
	dmarkets "socialpredict/internal/domain/markets"
	dusers "socialpredict/internal/domain/users"
	authsvc "socialpredict/internal/service/auth"
)

type lifecycleAuthMock struct {
	user *dusers.User
	err  *authsvc.AuthError
}

func (m lifecycleAuthMock) CurrentUser(*http.Request) (*dusers.User, *authsvc.AuthError) {
	return m.user, m.err
}

func (m lifecycleAuthMock) RequireUser(*http.Request) (*dusers.User, *authsvc.AuthError) {
	return m.user, m.err
}

func (m lifecycleAuthMock) RequireAdmin(*http.Request) (*dusers.User, *authsvc.AuthError) {
	return m.user, m.err
}

func TestListMyLifecycleMarketsAttachesMarketGroupMetadata(t *testing.T) {
	now := time.Now().UTC()
	group := &dmarkets.MarketGroup{
		ID:                 77,
		QuestionTitle:      "Match winner",
		Description:        "Grouped market description",
		GroupType:          dmarkets.MarketGroupTypeMultipleChoiceBinary,
		LifecycleStatus:    dmarkets.MarketLifecyclePublished,
		ProposalCost:       10,
		CreatorUsername:    "moderator",
		StewardUsername:    "steward",
		ApprovedBy:         "admin",
		ApprovedAt:         &now,
		ResolutionDateTime: now.Add(24 * time.Hour),
		CreatedAt:          now,
		UpdatedAt:          now,
		Members: []dmarkets.MarketGroupMember{
			{MarketID: 101, AnswerLabel: "Home", DisplayOrder: 0},
			{MarketID: 102, AnswerLabel: "Away", DisplayOrder: 1},
		},
	}
	svc := &MockService{
		ListLifecycleFn: func(_ context.Context, filters dmarkets.ListFilters) ([]*dmarkets.Market, error) {
			if filters.CreatedBy != "moderator" {
				t.Fatalf("CreatedBy = %q, want moderator", filters.CreatedBy)
			}
			if filters.Status != dmarkets.MarketLifecyclePublished {
				t.Fatalf("Status = %q, want published", filters.Status)
			}
			return []*dmarkets.Market{
				{
					ID:                 101,
					QuestionTitle:      "Match winner - Home",
					Description:        "Child description",
					OutcomeType:        "BINARY",
					ResolutionDateTime: group.ResolutionDateTime,
					CreatorUsername:    "moderator",
					StewardUsername:    "steward",
					Status:             dmarkets.MarketStatusActive,
					LifecycleStatus:    dmarkets.MarketLifecyclePublished,
					CreatedAt:          now,
					UpdatedAt:          now,
				},
			}, nil
		},
		MarketGroupLookupFn: func(_ context.Context, marketID int64) (*dmarkets.MarketGroup, error) {
			if marketID != 101 {
				t.Fatalf("marketID = %d, want 101", marketID)
			}
			return group, nil
		},
	}
	handler := ListMyLifecycleMarketsHandler(svc, lifecycleAuthMock{
		user: &dusers.User{Username: "moderator", UserType: string(dusers.UserTypeModerator)},
	})
	req := httptest.NewRequest(http.MethodGet, "/v0/profile/markets?status=published", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body=%s", rec.Code, rec.Body.String())
	}
	var envelope handlers.SuccessEnvelope[lifecycleMarketsResponse]
	if err := json.Unmarshal(rec.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if !envelope.OK || len(envelope.Result.Markets) != 1 {
		t.Fatalf("unexpected response: %+v", envelope)
	}
	market := envelope.Result.Markets[0]
	if market.MarketGroup == nil {
		t.Fatalf("expected marketGroup metadata")
	}
	want := dto.MarketGroupLink{
		ID:              group.ID,
		QuestionTitle:   group.QuestionTitle,
		Description:     group.Description,
		LifecycleStatus: group.LifecycleStatus,
		Status:          dmarkets.MarketStatusActive,
		AnswerLabel:     "Home",
		DisplayOrder:    0,
		AnswerCount:     2,
		ProposalCost:    10,
		CreatorUsername: "moderator",
		StewardUsername: "steward",
		ApprovedBy:      "admin",
	}
	got := *market.MarketGroup
	if got.ID != want.ID ||
		got.QuestionTitle != want.QuestionTitle ||
		got.Description != want.Description ||
		got.LifecycleStatus != want.LifecycleStatus ||
		got.Status != want.Status ||
		got.AnswerLabel != want.AnswerLabel ||
		got.DisplayOrder != want.DisplayOrder ||
		got.AnswerCount != want.AnswerCount ||
		got.ProposalCost != want.ProposalCost ||
		got.CreatorUsername != want.CreatorUsername ||
		got.StewardUsername != want.StewardUsername ||
		got.ApprovedBy != want.ApprovedBy {
		t.Fatalf("marketGroup = %+v, want %+v", got, want)
	}
}
