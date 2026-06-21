package adminhandlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"socialpredict/handlers"
	dmarkets "socialpredict/internal/domain/markets"
	dusers "socialpredict/internal/domain/users"
)

type amendmentReviewServiceMock struct {
	listFn        func(context.Context, dmarkets.AdminDescriptionAmendmentReviewFilters) (*dmarkets.AdminDescriptionAmendmentReviewPage, error)
	groupReviewFn func(context.Context, []int64, string, string, string) ([]dmarkets.MarketDescriptionAmendment, error)
}

func (m amendmentReviewServiceMock) GetMarketGovernanceSettings(context.Context) (*dmarkets.MarketGovernanceSettings, error) {
	return &dmarkets.MarketGovernanceSettings{}, nil
}

func (m amendmentReviewServiceMock) UpdateMarketGovernanceSettings(context.Context, dmarkets.MarketGovernanceSettingsUpdate) (*dmarkets.MarketGovernanceSettings, error) {
	return &dmarkets.MarketGovernanceSettings{}, nil
}

func (m amendmentReviewServiceMock) ListAdminMarketDescriptionAmendmentRows(ctx context.Context, filters dmarkets.AdminDescriptionAmendmentReviewFilters) (*dmarkets.AdminDescriptionAmendmentReviewPage, error) {
	return m.listFn(ctx, filters)
}

func (m amendmentReviewServiceMock) ReviewMarketDescriptionAmendment(context.Context, int64, string, string, string) (*dmarkets.MarketDescriptionAmendment, error) {
	return &dmarkets.MarketDescriptionAmendment{}, nil
}

func (m amendmentReviewServiceMock) ReviewGroupedMarketDescriptionAmendments(ctx context.Context, ids []int64, status string, actorUsername string, reason string) ([]dmarkets.MarketDescriptionAmendment, error) {
	if m.groupReviewFn != nil {
		return m.groupReviewFn(ctx, ids, status, actorUsername, reason)
	}
	return []dmarkets.MarketDescriptionAmendment{}, nil
}

type answerAdditionReviewServiceMock struct {
	listFn func(context.Context, dmarkets.AdminAnswerAdditionReviewFilters) (*dmarkets.AdminAnswerAdditionReviewPage, error)
}

func (m answerAdditionReviewServiceMock) ListAdminMarketGroupAnswerAdditionRows(ctx context.Context, filters dmarkets.AdminAnswerAdditionReviewFilters) (*dmarkets.AdminAnswerAdditionReviewPage, error) {
	return m.listFn(ctx, filters)
}

func (m answerAdditionReviewServiceMock) ApproveMarketGroupAnswerAddition(context.Context, int64, string, bool) (*dmarkets.MarketGroupAnswerAddition, error) {
	return &dmarkets.MarketGroupAnswerAddition{}, nil
}

func (m answerAdditionReviewServiceMock) RejectMarketGroupAnswerAddition(context.Context, int64, string, string) (*dmarkets.MarketGroupAnswerAddition, error) {
	return &dmarkets.MarketGroupAnswerAddition{}, nil
}

func TestListMarketDescriptionAmendmentsHandlerGroupsSearchesAndPaginates(t *testing.T) {
	now := time.Date(2026, 6, 20, 14, 0, 0, 0, time.UTC)
	group := &dmarkets.MarketGroup{
		ID:              7,
		QuestionTitle:   "Grouped orchard",
		Description:     "Grouped contract",
		LifecycleStatus: dmarkets.MarketLifecyclePublished,
		CreatorUsername: "moderator",
		Members: []dmarkets.MarketGroupMember{
			{MarketID: 101, AnswerLabel: "Apples", DisplayOrder: 0},
			{MarketID: 102, AnswerLabel: "Pears", DisplayOrder: 1},
		},
	}
	svc := amendmentReviewServiceMock{
		listFn: func(_ context.Context, filters dmarkets.AdminDescriptionAmendmentReviewFilters) (*dmarkets.AdminDescriptionAmendmentReviewPage, error) {
			if filters.Status != dmarkets.DescriptionAmendmentStatusPending || filters.Query != "orchard" || filters.Limit != 1 || filters.Offset != 0 {
				t.Fatalf("unexpected filters: %+v", filters)
			}
			return &dmarkets.AdminDescriptionAmendmentReviewPage{
				Rows: []dmarkets.AdminDescriptionAmendmentReviewRow{{
					RowKey:                 "group-amendment:7|pending|Clarify grouped orchard contract|moderator|",
					IsMarketGroupAmendment: true,
					Amendment: dmarkets.MarketDescriptionAmendment{
						ID:                1,
						MarketID:          101,
						MarketTitle:       "Grouped orchard",
						MarketDescription: "Grouped contract",
						MarketGroup:       group,
						Version:           2,
						Body:              "Clarify grouped orchard contract",
						Status:            dmarkets.DescriptionAmendmentStatusPending,
						CreatedBy:         "moderator",
						CreatedAt:         now,
						UpdatedAt:         now,
					},
					ChildAmendments: []dmarkets.MarketDescriptionAmendment{{
						ID:          1,
						MarketID:    101,
						MarketTitle: "Grouped orchard - Apples",
						MarketGroup: group,
						Version:     2,
						Body:        "Clarify grouped orchard contract",
						Status:      dmarkets.DescriptionAmendmentStatusPending,
						CreatedBy:   "moderator",
						CreatedAt:   now,
						UpdatedAt:   now,
					}, {
						ID:          2,
						MarketID:    102,
						MarketTitle: "Grouped orchard - Pears",
						MarketGroup: group,
						Version:     2,
						Body:        "Clarify grouped orchard contract",
						Status:      dmarkets.DescriptionAmendmentStatusPending,
						CreatedBy:   "moderator",
						CreatedAt:   now,
						UpdatedAt:   now,
					}},
				}},
				Total:  1,
				Limit:  filters.Limit,
				Offset: filters.Offset,
			}, nil
		},
	}
	handler := ListMarketDescriptionAmendmentsHandler(svc, marketReviewAuthMock{admin: &dusers.User{Username: "admin", UserType: string(dusers.UserTypeAdmin)}})
	req := httptest.NewRequest(http.MethodGet, "/v0/admin/market-description-amendments?status=pending&query=orchard&limit=1&offset=0", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body=%s", rec.Code, rec.Body.String())
	}
	var envelope handlers.SuccessEnvelope[marketDescriptionAmendmentListResponse]
	if err := json.Unmarshal(rec.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if !envelope.OK || envelope.Result.Total != 1 || len(envelope.Result.Amendments) != 1 {
		t.Fatalf("unexpected response: %+v", envelope)
	}
	row := envelope.Result.Amendments[0]
	if !row.IsMarketGroupAmendment || row.MarketTitle != "Grouped orchard" || len(row.ChildAmendments) != 2 {
		t.Fatalf("expected grouped amendment row with both children, got %+v", row)
	}
}

func TestListMarketDescriptionAmendmentsHandlerRejectsInvalidPagination(t *testing.T) {
	svc := amendmentReviewServiceMock{
		listFn: func(context.Context, dmarkets.AdminDescriptionAmendmentReviewFilters) (*dmarkets.AdminDescriptionAmendmentReviewPage, error) {
			t.Fatal("service should not be called for invalid pagination")
			return nil, nil
		},
	}
	handler := ListMarketDescriptionAmendmentsHandler(svc, marketReviewAuthMock{admin: &dusers.User{Username: "admin", UserType: string(dusers.UserTypeAdmin)}})
	for _, path := range []string{
		"/v0/admin/market-description-amendments?status=pending&limit=abc",
		"/v0/admin/market-description-amendments?status=pending&limit=0",
		"/v0/admin/market-description-amendments?status=pending&limit=201",
		"/v0/admin/market-description-amendments?status=pending&offset=-1",
	} {
		t.Run(path, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, path, nil)
			rec := httptest.NewRecorder()

			handler.ServeHTTP(rec, req)

			if rec.Code != http.StatusBadRequest {
				t.Fatalf("status = %d, body=%s", rec.Code, rec.Body.String())
			}
		})
	}
}

func TestListMarketGroupAnswerAdditionsHandlerSearchesAndPaginates(t *testing.T) {
	now := time.Date(2026, 6, 20, 14, 15, 0, 0, time.UTC)
	svc := answerAdditionReviewServiceMock{
		listFn: func(_ context.Context, filters dmarkets.AdminAnswerAdditionReviewFilters) (*dmarkets.AdminAnswerAdditionReviewPage, error) {
			if filters.Status != dmarkets.MarketGroupAnswerAdditionStatusPending || filters.Query != "cedar" || filters.Limit != 1 || filters.Offset != 0 {
				t.Fatalf("unexpected filters: %+v", filters)
			}
			return &dmarkets.AdminAnswerAdditionReviewPage{
				Rows: []dmarkets.MarketGroupAnswerAddition{{
					ID:          1,
					GroupID:     9,
					GroupTitle:  "Favorite tree",
					AnswerLabel: "Cedar",
					Status:      dmarkets.MarketGroupAnswerAdditionStatusPending,
					ProposedBy:  "moderator",
					CreatedAt:   now,
					UpdatedAt:   now,
				}},
				Total:  1,
				Limit:  filters.Limit,
				Offset: filters.Offset,
			}, nil
		},
	}
	handler := ListMarketGroupAnswerAdditionsHandler(svc, marketReviewAuthMock{admin: &dusers.User{Username: "admin", UserType: string(dusers.UserTypeAdmin)}})
	req := httptest.NewRequest(http.MethodGet, "/v0/admin/market-group-answer-additions?status=pending&query=cedar&limit=1&offset=0", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body=%s", rec.Code, rec.Body.String())
	}
	var envelope handlers.SuccessEnvelope[marketGroupAnswerAdditionListResponse]
	if err := json.Unmarshal(rec.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if !envelope.OK || envelope.Result.Total != 1 || len(envelope.Result.Additions) != 1 {
		t.Fatalf("unexpected response: %+v", envelope)
	}
	if envelope.Result.Additions[0].AnswerLabel != "Cedar" {
		t.Fatalf("expected Cedar result, got %+v", envelope.Result.Additions[0])
	}
}

func TestListMarketGroupAnswerAdditionsHandlerRejectsInvalidPagination(t *testing.T) {
	svc := answerAdditionReviewServiceMock{
		listFn: func(context.Context, dmarkets.AdminAnswerAdditionReviewFilters) (*dmarkets.AdminAnswerAdditionReviewPage, error) {
			t.Fatal("service should not be called for invalid pagination")
			return nil, nil
		},
	}
	handler := ListMarketGroupAnswerAdditionsHandler(svc, marketReviewAuthMock{admin: &dusers.User{Username: "admin", UserType: string(dusers.UserTypeAdmin)}})
	for _, path := range []string{
		"/v0/admin/market-group-answer-additions?status=pending&limit=abc",
		"/v0/admin/market-group-answer-additions?status=pending&limit=0",
		"/v0/admin/market-group-answer-additions?status=pending&limit=201",
		"/v0/admin/market-group-answer-additions?status=pending&offset=-1",
	} {
		t.Run(path, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, path, nil)
			rec := httptest.NewRecorder()

			handler.ServeHTTP(rec, req)

			if rec.Code != http.StatusBadRequest {
				t.Fatalf("status = %d, body=%s", rec.Code, rec.Body.String())
			}
		})
	}
}

func TestReviewGroupedMarketDescriptionAmendmentsHandlerReviewsAllChildren(t *testing.T) {
	now := time.Date(2026, 6, 20, 15, 30, 0, 0, time.UTC)
	svc := amendmentReviewServiceMock{
		groupReviewFn: func(_ context.Context, ids []int64, status string, actorUsername string, reason string) ([]dmarkets.MarketDescriptionAmendment, error) {
			if actorUsername != "admin" || status != dmarkets.DescriptionAmendmentStatusApproved || reason != "approve group" {
				t.Fatalf("unexpected review args: ids=%v status=%q actor=%q reason=%q", ids, status, actorUsername, reason)
			}
			if len(ids) != 2 || ids[0] != 10 || ids[1] != 11 {
				t.Fatalf("ids = %v, want [10 11]", ids)
			}
			return []dmarkets.MarketDescriptionAmendment{
				{ID: 10, MarketID: 101, Version: 2, Body: "Grouped clarification", Status: dmarkets.DescriptionAmendmentStatusApproved, CreatedBy: "moderator", ApprovedBy: actorUsername, ApprovedAt: &now},
				{ID: 11, MarketID: 102, Version: 2, Body: "Grouped clarification", Status: dmarkets.DescriptionAmendmentStatusApproved, CreatedBy: "moderator", ApprovedBy: actorUsername, ApprovedAt: &now},
			}, nil
		},
	}
	handler := ReviewGroupedMarketDescriptionAmendmentsHandler(svc, marketReviewAuthMock{admin: &dusers.User{Username: "admin", UserType: string(dusers.UserTypeAdmin)}})
	req := httptest.NewRequest(http.MethodPatch, "/v0/admin/market-description-amendments/grouped-review", strings.NewReader(`{"amendmentIds":[10,11],"status":"approved","reason":"approve group"}`))
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body=%s", rec.Code, rec.Body.String())
	}
	var envelope handlers.SuccessEnvelope[marketDescriptionAmendmentListResponse]
	if err := json.Unmarshal(rec.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if !envelope.OK || envelope.Result.Total != 2 || len(envelope.Result.Amendments) != 2 {
		t.Fatalf("unexpected response: %+v", envelope)
	}
}
