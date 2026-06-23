package adminhandlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gorilla/mux"

	"socialpredict/handlers"
	dmarkets "socialpredict/internal/domain/markets"
	dusers "socialpredict/internal/domain/users"
	authsvc "socialpredict/internal/service/auth"
)

type marketReviewServiceMock struct {
	approveFn       func(context.Context, int64, string, bool) (*dmarkets.Market, error)
	groupApproveFn  func(context.Context, int64, string, bool) (*dmarkets.MarketGroup, error)
	rejectFn        func(context.Context, int64, string, string) (*dmarkets.Market, error)
	groupRejectFn   func(context.Context, int64, string, string) (*dmarkets.MarketGroup, error)
	listFn          func(context.Context, dmarkets.AdminMarketReviewFilters) (*dmarkets.AdminMarketReviewPage, error)
	groupLookupFn   func(context.Context, int64) (*dmarkets.MarketGroup, error)
	reassignFn      func(context.Context, int64, string, string, string) (*dmarkets.Market, error)
	groupReassignFn func(context.Context, int64, string, string, string) (*dmarkets.MarketGroup, error)
	tagsFn          func(context.Context, int64, []string, string) (*dmarkets.Market, error)
	groupTagsFn     func(context.Context, int64, []string, string) (*dmarkets.AdminMarketReviewRow, error)
}

func (m marketReviewServiceMock) ApproveProposedMarket(ctx context.Context, marketID int64, actorUsername string, confirmed bool) (*dmarkets.Market, error) {
	return m.approveFn(ctx, marketID, actorUsername, confirmed)
}

func (m marketReviewServiceMock) ApproveProposedMarketGroup(ctx context.Context, groupID int64, actorUsername string, confirmed bool) (*dmarkets.MarketGroup, error) {
	return m.groupApproveFn(ctx, groupID, actorUsername, confirmed)
}

func (m marketReviewServiceMock) RejectProposedMarket(ctx context.Context, marketID int64, actorUsername string, reason string) (*dmarkets.Market, error) {
	return m.rejectFn(ctx, marketID, actorUsername, reason)
}

func (m marketReviewServiceMock) RejectProposedMarketGroup(ctx context.Context, groupID int64, actorUsername string, reason string) (*dmarkets.MarketGroup, error) {
	return m.groupRejectFn(ctx, groupID, actorUsername, reason)
}

func (m marketReviewServiceMock) ListAdminMarketReviewRows(ctx context.Context, filters dmarkets.AdminMarketReviewFilters) (*dmarkets.AdminMarketReviewPage, error) {
	return m.listFn(ctx, filters)
}

func (m marketReviewServiceMock) GetMarketGroupForMarket(ctx context.Context, marketID int64) (*dmarkets.MarketGroup, error) {
	if m.groupLookupFn == nil {
		return nil, dmarkets.ErrMarketGroupNotFound
	}
	return m.groupLookupFn(ctx, marketID)
}

func (m marketReviewServiceMock) ReassignMarketSteward(ctx context.Context, marketID int64, newStewardUsername string, actorUsername string, reason string) (*dmarkets.Market, error) {
	return m.reassignFn(ctx, marketID, newStewardUsername, actorUsername, reason)
}

func (m marketReviewServiceMock) ReassignMarketGroupSteward(ctx context.Context, groupID int64, newStewardUsername string, actorUsername string, reason string) (*dmarkets.MarketGroup, error) {
	return m.groupReassignFn(ctx, groupID, newStewardUsername, actorUsername, reason)
}

func (m marketReviewServiceMock) UpdateMarketTags(ctx context.Context, marketID int64, tagSlugs []string, actorUsername string) (*dmarkets.Market, error) {
	return m.tagsFn(ctx, marketID, tagSlugs, actorUsername)
}

func (m marketReviewServiceMock) UpdateMarketGroupTags(ctx context.Context, groupID int64, tagSlugs []string, actorUsername string) (*dmarkets.AdminMarketReviewRow, error) {
	return m.groupTagsFn(ctx, groupID, tagSlugs, actorUsername)
}

type marketReviewAuthMock struct {
	admin *dusers.User
	err   *authsvc.AuthError
}

func (m marketReviewAuthMock) CurrentUser(r *http.Request) (*dusers.User, *authsvc.AuthError) {
	return m.admin, m.err
}

func (m marketReviewAuthMock) RequireUser(r *http.Request) (*dusers.User, *authsvc.AuthError) {
	return m.admin, m.err
}

func (m marketReviewAuthMock) RequireAdmin(r *http.Request) (*dusers.User, *authsvc.AuthError) {
	return m.admin, m.err
}

func TestApproveMarketHandlerApprovesProposal(t *testing.T) {
	svc := marketReviewServiceMock{
		approveFn: func(_ context.Context, marketID int64, actorUsername string, confirmed bool) (*dmarkets.Market, error) {
			if marketID != 42 || actorUsername != "admin" || !confirmed {
				t.Fatalf("unexpected approve args: id=%d actor=%q confirmed=%v", marketID, actorUsername, confirmed)
			}
			return &dmarkets.Market{ID: marketID, Status: dmarkets.MarketStatusActive, LifecycleStatus: dmarkets.MarketLifecyclePublished, ApprovedBy: actorUsername}, nil
		},
	}
	handler := ApproveMarketHandler(svc, marketReviewAuthMock{admin: &dusers.User{Username: "admin", UserType: string(dusers.UserTypeAdmin)}})
	req := mux.SetURLVars(httptest.NewRequest(http.MethodPatch, "/v0/admin/markets/42/approve", bytes.NewBufferString(`{"confirm":true}`)), map[string]string{"id": "42"})
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body=%s", rec.Code, rec.Body.String())
	}
	var envelope handlers.SuccessEnvelope[marketReviewResponse]
	if err := json.Unmarshal(rec.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if !envelope.OK || envelope.Result.LifecycleStatus != dmarkets.MarketLifecyclePublished || envelope.Result.ApprovedBy != "admin" {
		t.Fatalf("unexpected response: %+v", envelope)
	}
}

func TestApproveMarketGroupHandlerApprovesProposalGroup(t *testing.T) {
	svc := marketReviewServiceMock{
		groupApproveFn: func(_ context.Context, groupID int64, actorUsername string, confirmed bool) (*dmarkets.MarketGroup, error) {
			if groupID != 50 || actorUsername != "admin" || !confirmed {
				t.Fatalf("unexpected group approve args: id=%d actor=%q confirmed=%v", groupID, actorUsername, confirmed)
			}
			return &dmarkets.MarketGroup{
				ID:              groupID,
				QuestionTitle:   "Grouped match winner",
				LifecycleStatus: dmarkets.MarketLifecyclePublished,
				ApprovedBy:      actorUsername,
				Members: []dmarkets.MarketGroupMember{
					{MarketID: 101, AnswerLabel: "Home", DisplayOrder: 0},
					{MarketID: 102, AnswerLabel: "Away", DisplayOrder: 1},
				},
			}, nil
		},
	}
	handler := ApproveMarketGroupHandler(svc, marketReviewAuthMock{admin: &dusers.User{Username: "admin", UserType: string(dusers.UserTypeAdmin)}})
	req := mux.SetURLVars(httptest.NewRequest(http.MethodPatch, "/v0/admin/market-groups/50/approve", bytes.NewBufferString(`{"confirm":true}`)), map[string]string{"id": "50"})
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body=%s", rec.Code, rec.Body.String())
	}
	var envelope handlers.SuccessEnvelope[marketGroupReviewResponse]
	if err := json.Unmarshal(rec.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if !envelope.OK || envelope.Result.ID != 50 || envelope.Result.Status != dmarkets.MarketStatusActive || envelope.Result.AnswerCount != 2 {
		t.Fatalf("unexpected response: %+v", envelope)
	}
}

func TestRejectMarketHandlerRejectsProposal(t *testing.T) {
	svc := marketReviewServiceMock{
		rejectFn: func(_ context.Context, marketID int64, actorUsername string, reason string) (*dmarkets.Market, error) {
			if marketID != 43 || actorUsername != "admin" || reason != "duplicate" {
				t.Fatalf("unexpected reject args: id=%d actor=%q reason=%q", marketID, actorUsername, reason)
			}
			return &dmarkets.Market{ID: marketID, Status: dmarkets.MarketLifecycleRejected, LifecycleStatus: dmarkets.MarketLifecycleRejected, RejectedBy: actorUsername, RejectionReason: reason}, nil
		},
	}
	handler := RejectMarketHandler(svc, marketReviewAuthMock{admin: &dusers.User{Username: "admin", UserType: string(dusers.UserTypeAdmin)}})
	req := mux.SetURLVars(httptest.NewRequest(http.MethodPatch, "/v0/admin/markets/43/reject", bytes.NewBufferString(`{"reason":"duplicate"}`)), map[string]string{"id": "43"})
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body=%s", rec.Code, rec.Body.String())
	}
	var envelope handlers.SuccessEnvelope[marketReviewResponse]
	if err := json.Unmarshal(rec.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if !envelope.OK || envelope.Result.LifecycleStatus != dmarkets.MarketLifecycleRejected || envelope.Result.RejectionReason != "duplicate" {
		t.Fatalf("unexpected response: %+v", envelope)
	}
}

func TestRejectMarketGroupHandlerRejectsProposalGroup(t *testing.T) {
	svc := marketReviewServiceMock{
		groupRejectFn: func(_ context.Context, groupID int64, actorUsername string, reason string) (*dmarkets.MarketGroup, error) {
			if groupID != 51 || actorUsername != "admin" || reason != "duplicate answers" {
				t.Fatalf("unexpected group reject args: id=%d actor=%q reason=%q", groupID, actorUsername, reason)
			}
			return &dmarkets.MarketGroup{
				ID:              groupID,
				QuestionTitle:   "Grouped match winner",
				LifecycleStatus: dmarkets.MarketLifecycleRejected,
				RejectedBy:      actorUsername,
				RejectionReason: reason,
				Members: []dmarkets.MarketGroupMember{
					{MarketID: 101, AnswerLabel: "Home", DisplayOrder: 0},
					{MarketID: 102, AnswerLabel: "Away", DisplayOrder: 1},
				},
			}, nil
		},
	}
	handler := RejectMarketGroupHandler(svc, marketReviewAuthMock{admin: &dusers.User{Username: "admin", UserType: string(dusers.UserTypeAdmin)}})
	req := mux.SetURLVars(httptest.NewRequest(http.MethodPatch, "/v0/admin/market-groups/51/reject", bytes.NewBufferString(`{"reason":"duplicate answers"}`)), map[string]string{"id": "51"})
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body=%s", rec.Code, rec.Body.String())
	}
	var envelope handlers.SuccessEnvelope[marketGroupReviewResponse]
	if err := json.Unmarshal(rec.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if !envelope.OK || envelope.Result.LifecycleStatus != dmarkets.MarketLifecycleRejected || envelope.Result.RejectionReason != "duplicate answers" {
		t.Fatalf("unexpected response: %+v", envelope)
	}
}

func TestListReviewMarketsHandlerReturnsQueue(t *testing.T) {
	changedAt := time.Date(2026, 6, 4, 12, 0, 0, 0, time.UTC)
	svc := marketReviewServiceMock{
		listFn: func(_ context.Context, filters dmarkets.AdminMarketReviewFilters) (*dmarkets.AdminMarketReviewPage, error) {
			if filters.Status != dmarkets.MarketLifecycleProposed || filters.Limit != 25 || filters.Offset != 0 {
				t.Fatalf("unexpected filters: %+v", filters)
			}
			market := &dmarkets.Market{
				ID:              44,
				QuestionTitle:   "Queue item",
				CreatorUsername: "moderator",
				StewardUsername: "backup",
				YesLabel:        "BIG",
				NoLabel:         "SMALL",
				Status:          dmarkets.MarketLifecycleProposed,
				LifecycleStatus: dmarkets.MarketLifecycleProposed,
				StewardshipAudits: []dmarkets.MarketStewardshipAuditRecord{{
					ID:                  9,
					MarketID:            44,
					FromStewardUsername: "moderator",
					ToStewardUsername:   "backup",
					ActorUsername:       "admin",
					Reason:              "moderator inactive",
					CreatedAt:           changedAt,
				}},
			}
			return &dmarkets.AdminMarketReviewPage{
				Rows:   []dmarkets.AdminMarketReviewRow{{RowKey: "market:44", Market: market}},
				Total:  1,
				Limit:  filters.Limit,
				Offset: filters.Offset,
			}, nil
		},
	}
	handler := ListReviewMarketsHandler(svc, marketReviewAuthMock{admin: &dusers.User{Username: "admin", UserType: string(dusers.UserTypeAdmin)}})
	req := httptest.NewRequest(http.MethodGet, "/v0/admin/markets?status=proposed&limit=25&offset=0", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body=%s", rec.Code, rec.Body.String())
	}
	var envelope handlers.SuccessEnvelope[marketReviewListResponse]
	if err := json.Unmarshal(rec.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if !envelope.OK || envelope.Result.Total != 1 || envelope.Result.Limit != 25 || envelope.Result.Markets[0].CreatorUsername != "moderator" {
		t.Fatalf("unexpected response: %+v", envelope)
	}
	if envelope.Result.Markets[0].StewardUsername != "backup" {
		t.Fatalf("expected steward in admin queue response, got %+v", envelope.Result.Markets[0])
	}
	if envelope.Result.Markets[0].YesLabel != "BIG" || envelope.Result.Markets[0].NoLabel != "SMALL" {
		t.Fatalf("expected custom labels in admin queue response, got %+v", envelope.Result.Markets[0])
	}
	audits := envelope.Result.Markets[0].StewardshipAudits
	if len(audits) != 1 || audits[0].Reason != "moderator inactive" || audits[0].FromStewardUsername != "moderator" || audits[0].ToStewardUsername != "backup" || !audits[0].CreatedAt.Equal(changedAt) {
		t.Fatalf("expected stewardship audit in admin queue response, got %+v", audits)
	}
}

func TestListReviewMarketsHandlerRejectsInvalidPagination(t *testing.T) {
	svc := marketReviewServiceMock{
		listFn: func(context.Context, dmarkets.AdminMarketReviewFilters) (*dmarkets.AdminMarketReviewPage, error) {
			t.Fatal("service should not be called for invalid pagination")
			return nil, nil
		},
	}
	handler := ListReviewMarketsHandler(svc, marketReviewAuthMock{admin: &dusers.User{Username: "admin", UserType: string(dusers.UserTypeAdmin)}})
	for _, path := range []string{
		"/v0/admin/markets?status=published&limit=abc",
		"/v0/admin/markets?status=published&limit=0",
		"/v0/admin/markets?status=published&limit=101",
		"/v0/admin/markets?status=published&offset=-1",
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

func TestListReviewMarketsHandlerAttachesMarketGroupMetadata(t *testing.T) {
	svc := marketReviewServiceMock{
		listFn: func(_ context.Context, filters dmarkets.AdminMarketReviewFilters) (*dmarkets.AdminMarketReviewPage, error) {
			if filters.Status != dmarkets.MarketLifecyclePublished {
				t.Fatalf("unexpected filters: %+v", filters)
			}
			child := &dmarkets.Market{
				ID:              101,
				QuestionTitle:   "Grouped match winner - Home",
				CreatorUsername: "moderator",
				Status:          dmarkets.MarketStatusActive,
				LifecycleStatus: dmarkets.MarketLifecyclePublished,
			}
			group := &dmarkets.MarketGroup{
				ID:              50,
				QuestionTitle:   "Grouped match winner",
				Description:     "One parent review item for answer markets.",
				LifecycleStatus: dmarkets.MarketLifecyclePublished,
				CreatorUsername: "moderator",
				StewardUsername: "backup",
				Members: []dmarkets.MarketGroupMember{
					{MarketID: 101, AnswerLabel: "Home", DisplayOrder: 0},
					{MarketID: 102, AnswerLabel: "Away", DisplayOrder: 1},
				},
			}
			return &dmarkets.AdminMarketReviewPage{
				Rows: []dmarkets.AdminMarketReviewRow{{
					RowKey:        "group:50",
					IsMarketGroup: true,
					Market:        child,
					Group:         group,
					Children: []*dmarkets.Market{
						child,
						{ID: 102, QuestionTitle: "Grouped match winner - Away", CreatorUsername: "moderator", Status: dmarkets.MarketStatusActive, LifecycleStatus: dmarkets.MarketLifecyclePublished},
					},
				}},
				Total:  1,
				Limit:  filters.Limit,
				Offset: filters.Offset,
			}, nil
		},
	}
	handler := ListReviewMarketsHandler(svc, marketReviewAuthMock{admin: &dusers.User{Username: "admin", UserType: string(dusers.UserTypeAdmin)}})
	req := httptest.NewRequest(http.MethodGet, "/v0/admin/markets?status=published", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body=%s", rec.Code, rec.Body.String())
	}
	var envelope handlers.SuccessEnvelope[marketReviewListResponse]
	if err := json.Unmarshal(rec.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	group := envelope.Result.Markets[0].MarketGroup
	if !envelope.OK || !envelope.Result.Markets[0].IsMarketGroup || envelope.Result.Markets[0].ID != 50 || group == nil || group.ID != 50 || group.AnswerLabel != "Home" || group.AnswerCount != 2 || group.StewardUsername != "backup" {
		t.Fatalf("expected group metadata in admin queue response, got %+v", envelope.Result.Markets[0])
	}
	if len(envelope.Result.Markets[0].ChildMarkets) != 2 {
		t.Fatalf("expected full child market list in grouped admin row, got %+v", envelope.Result.Markets[0].ChildMarkets)
	}
}

func TestListReviewMarketsHandlerSupportsAllStatusSearch(t *testing.T) {
	svc := marketReviewServiceMock{
		listFn: func(_ context.Context, filters dmarkets.AdminMarketReviewFilters) (*dmarkets.AdminMarketReviewPage, error) {
			if filters.Status != dmarkets.MarketStatusAll || filters.Query != "orchard" || filters.Limit != 100 || filters.Offset != 0 {
				t.Fatalf("unexpected filters: %+v", filters)
			}
			market := &dmarkets.Market{
				ID:              45,
				QuestionTitle:   "Orchard market",
				CreatorUsername: "moderator",
				Status:          dmarkets.MarketStatusResolved,
				LifecycleStatus: dmarkets.MarketLifecycleResolved,
			}
			return &dmarkets.AdminMarketReviewPage{
				Rows:   []dmarkets.AdminMarketReviewRow{{RowKey: "market:45", Market: market}},
				Total:  1,
				Limit:  filters.Limit,
				Offset: filters.Offset,
			}, nil
		},
	}
	handler := ListReviewMarketsHandler(svc, marketReviewAuthMock{admin: &dusers.User{Username: "admin", UserType: string(dusers.UserTypeAdmin)}})
	req := httptest.NewRequest(http.MethodGet, "/v0/admin/markets?status=all&query=orchard&limit=100", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body=%s", rec.Code, rec.Body.String())
	}
	var envelope handlers.SuccessEnvelope[marketReviewListResponse]
	if err := json.Unmarshal(rec.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if !envelope.OK || envelope.Result.Total != 1 || envelope.Result.Markets[0].LifecycleStatus != dmarkets.MarketLifecycleResolved {
		t.Fatalf("unexpected response: %+v", envelope)
	}
}

func TestListReviewMarketsHandlerPaginatesAfterGrouping(t *testing.T) {
	svc := marketReviewServiceMock{
		listFn: func(_ context.Context, filters dmarkets.AdminMarketReviewFilters) (*dmarkets.AdminMarketReviewPage, error) {
			if filters.Status != dmarkets.MarketLifecyclePublished || filters.Limit != 1 || filters.Offset != 1 {
				t.Fatalf("unexpected filters: %+v", filters)
			}
			market := &dmarkets.Market{
				ID:              201,
				QuestionTitle:   "Solo market",
				CreatorUsername: "moderator",
				Status:          dmarkets.MarketStatusActive,
				LifecycleStatus: dmarkets.MarketLifecyclePublished,
			}
			return &dmarkets.AdminMarketReviewPage{
				Rows:   []dmarkets.AdminMarketReviewRow{{RowKey: "market:201", Market: market}},
				Total:  2,
				Limit:  filters.Limit,
				Offset: filters.Offset,
			}, nil
		},
	}
	handler := ListReviewMarketsHandler(svc, marketReviewAuthMock{admin: &dusers.User{Username: "admin", UserType: string(dusers.UserTypeAdmin)}})
	req := httptest.NewRequest(http.MethodGet, "/v0/admin/markets?status=published&limit=1&offset=1", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body=%s", rec.Code, rec.Body.String())
	}
	var envelope handlers.SuccessEnvelope[marketReviewListResponse]
	if err := json.Unmarshal(rec.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if !envelope.OK || envelope.Result.Total != 2 || len(envelope.Result.Markets) != 1 {
		t.Fatalf("unexpected response: %+v", envelope)
	}
	if envelope.Result.Markets[0].QuestionTitle != "Solo market" {
		t.Fatalf("expected second grouped row to be solo market, got %+v", envelope.Result.Markets[0])
	}
}

func TestReassignMarketGroupStewardHandlerReassignsGroupSteward(t *testing.T) {
	changedAt := time.Date(2026, 6, 4, 13, 0, 0, 0, time.UTC)
	svc := marketReviewServiceMock{
		groupReassignFn: func(_ context.Context, groupID int64, newStewardUsername string, actorUsername string, reason string) (*dmarkets.MarketGroup, error) {
			if groupID != 52 || newStewardUsername != "backup" || actorUsername != "admin" || reason != "moderator inactive" {
				t.Fatalf("unexpected group reassign args: id=%d steward=%q actor=%q reason=%q", groupID, newStewardUsername, actorUsername, reason)
			}
			return &dmarkets.MarketGroup{
				ID:              groupID,
				QuestionTitle:   "Grouped match winner",
				CreatorUsername: "moderator",
				StewardUsername: newStewardUsername,
				LifecycleStatus: dmarkets.MarketLifecyclePublished,
				UpdatedAt:       changedAt,
				Members: []dmarkets.MarketGroupMember{
					{MarketID: 101, AnswerLabel: "Home", DisplayOrder: 0},
					{MarketID: 102, AnswerLabel: "Away", DisplayOrder: 1},
				},
			}, nil
		},
	}
	handler := ReassignMarketGroupStewardHandler(svc, marketReviewAuthMock{admin: &dusers.User{Username: "admin", UserType: string(dusers.UserTypeAdmin)}})
	req := mux.SetURLVars(httptest.NewRequest(http.MethodPatch, "/v0/admin/market-groups/52/steward", bytes.NewBufferString(`{"stewardUsername":"backup","reason":"moderator inactive"}`)), map[string]string{"id": "52"})
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body=%s", rec.Code, rec.Body.String())
	}
	var envelope handlers.SuccessEnvelope[marketGroupReviewResponse]
	if err := json.Unmarshal(rec.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if !envelope.OK || envelope.Result.StewardUsername != "backup" || envelope.Result.CreatorUsername != "moderator" || envelope.Result.AnswerCount != 2 {
		t.Fatalf("unexpected response: %+v", envelope)
	}
}

func TestReassignMarketStewardHandlerReassignsSteward(t *testing.T) {
	changedAt := time.Date(2026, 6, 4, 13, 0, 0, 0, time.UTC)
	svc := marketReviewServiceMock{
		reassignFn: func(_ context.Context, marketID int64, newStewardUsername string, actorUsername string, reason string) (*dmarkets.Market, error) {
			if marketID != 45 || newStewardUsername != "backup" || actorUsername != "admin" || reason != "moderator inactive" {
				t.Fatalf("unexpected reassign args: id=%d steward=%q actor=%q reason=%q", marketID, newStewardUsername, actorUsername, reason)
			}
			return &dmarkets.Market{
				ID:              marketID,
				CreatorUsername: "moderator",
				StewardUsername: newStewardUsername,
				Status:          dmarkets.MarketStatusActive,
				LifecycleStatus: dmarkets.MarketLifecyclePublished,
				StewardshipAudits: []dmarkets.MarketStewardshipAuditRecord{{
					MarketID:            marketID,
					FromStewardUsername: "moderator",
					ToStewardUsername:   newStewardUsername,
					ActorUsername:       actorUsername,
					Reason:              reason,
					CreatedAt:           changedAt,
				}},
			}, nil
		},
	}
	handler := ReassignMarketStewardHandler(svc, marketReviewAuthMock{admin: &dusers.User{Username: "admin", UserType: string(dusers.UserTypeAdmin)}})
	req := mux.SetURLVars(httptest.NewRequest(http.MethodPatch, "/v0/admin/markets/45/steward", bytes.NewBufferString(`{"stewardUsername":"backup","reason":"moderator inactive"}`)), map[string]string{"id": "45"})
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body=%s", rec.Code, rec.Body.String())
	}
	var envelope handlers.SuccessEnvelope[marketReviewResponse]
	if err := json.Unmarshal(rec.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if !envelope.OK || envelope.Result.StewardUsername != "backup" || envelope.Result.CreatorUsername != "moderator" {
		t.Fatalf("unexpected response: %+v", envelope)
	}
	if len(envelope.Result.StewardshipAudits) != 1 || envelope.Result.StewardshipAudits[0].Reason != "moderator inactive" {
		t.Fatalf("expected stewardship audit in response, got %+v", envelope.Result.StewardshipAudits)
	}
}

func TestUpdateMarketTagsHandlerRewritesMarketTags(t *testing.T) {
	svc := marketReviewServiceMock{
		tagsFn: func(_ context.Context, marketID int64, tagSlugs []string, actorUsername string) (*dmarkets.Market, error) {
			if marketID != 46 || actorUsername != "admin" || len(tagSlugs) != 2 || tagSlugs[0] != "sports" || tagSlugs[1] != "policy" {
				t.Fatalf("unexpected tag update args: id=%d actor=%q slugs=%+v", marketID, actorUsername, tagSlugs)
			}
			return &dmarkets.Market{
				ID:              marketID,
				QuestionTitle:   "Tagged market",
				Status:          dmarkets.MarketStatusActive,
				LifecycleStatus: dmarkets.MarketLifecyclePublished,
				Tags: []dmarkets.MarketTag{
					{ID: 1, Slug: "policy", DisplayName: "Policy", IsActive: true},
					{ID: 2, Slug: "sports", DisplayName: "Sports", IsActive: true},
				},
			}, nil
		},
	}
	handler := UpdateMarketTagsHandler(svc, marketReviewAuthMock{admin: &dusers.User{Username: "admin", UserType: string(dusers.UserTypeAdmin)}})
	req := mux.SetURLVars(httptest.NewRequest(http.MethodPatch, "/v0/admin/markets/46/tags", bytes.NewBufferString(`{"tagSlugs":["sports","policy"]}`)), map[string]string{"id": "46"})
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body=%s", rec.Code, rec.Body.String())
	}
	var envelope handlers.SuccessEnvelope[marketReviewResponse]
	if err := json.Unmarshal(rec.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if !envelope.OK || len(envelope.Result.Tags) != 2 || envelope.Result.Tags[0].Slug != "policy" {
		t.Fatalf("unexpected response: %+v", envelope)
	}
}

func TestUpdateMarketGroupTagsHandlerRewritesGroupTags(t *testing.T) {
	group := &dmarkets.MarketGroup{
		ID:              47,
		QuestionTitle:   "Grouped tagged market",
		LifecycleStatus: dmarkets.MarketLifecyclePublished,
		CreatorUsername: "moderator",
		Members: []dmarkets.MarketGroupMember{
			{MarketID: 4701, AnswerLabel: "A"},
			{MarketID: 4702, AnswerLabel: "B"},
		},
	}
	svc := marketReviewServiceMock{
		groupTagsFn: func(_ context.Context, groupID int64, tagSlugs []string, actorUsername string) (*dmarkets.AdminMarketReviewRow, error) {
			if groupID != 47 || actorUsername != "admin" || len(tagSlugs) != 2 || tagSlugs[0] != "sports" || tagSlugs[1] != "policy" {
				t.Fatalf("unexpected group tag update args: id=%d actor=%q slugs=%+v", groupID, actorUsername, tagSlugs)
			}
			tags := []dmarkets.MarketTag{
				{ID: 1, Slug: "policy", DisplayName: "Policy", IsActive: true},
				{ID: 2, Slug: "sports", DisplayName: "Sports", IsActive: true},
			}
			return &dmarkets.AdminMarketReviewRow{
				RowKey:        "group:47",
				IsMarketGroup: true,
				Group:         group,
				Market:        &dmarkets.Market{ID: 4701, QuestionTitle: "Grouped tagged market - A", Tags: tags},
				Children: []*dmarkets.Market{
					{ID: 4701, QuestionTitle: "Grouped tagged market - A", Tags: tags},
					{ID: 4702, QuestionTitle: "Grouped tagged market - B", Tags: tags},
				},
				Tags: tags,
			}, nil
		},
	}
	handler := UpdateMarketGroupTagsHandler(svc, marketReviewAuthMock{admin: &dusers.User{Username: "admin", UserType: string(dusers.UserTypeAdmin)}})
	req := mux.SetURLVars(httptest.NewRequest(http.MethodPatch, "/v0/admin/market-groups/47/tags", bytes.NewBufferString(`{"tagSlugs":["sports","policy"]}`)), map[string]string{"id": "47"})
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body=%s", rec.Code, rec.Body.String())
	}
	var envelope handlers.SuccessEnvelope[marketReviewResponse]
	if err := json.Unmarshal(rec.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if !envelope.OK || !envelope.Result.IsMarketGroup || envelope.Result.ID != 47 || len(envelope.Result.Tags) != 2 || len(envelope.Result.ChildMarkets) != 2 {
		t.Fatalf("unexpected response: %+v", envelope)
	}
}

func TestMarketReviewHandlersReturnFailureReasons(t *testing.T) {
	tests := []struct {
		name       string
		handler    http.HandlerFunc
		body       string
		wantStatus int
		wantReason handlers.FailureReason
	}{
		{
			name:       "approve unauthorized",
			handler:    ApproveMarketHandler(marketReviewServiceMock{}, marketReviewAuthMock{err: &authsvc.AuthError{Kind: authsvc.ErrorKindAdminRequired, Message: "admin required"}}),
			body:       `{"confirm":true}`,
			wantStatus: http.StatusForbidden,
			wantReason: handlers.ReasonAuthorizationDenied,
		},
		{
			name: "approve wrong state",
			handler: ApproveMarketHandler(marketReviewServiceMock{approveFn: func(context.Context, int64, string, bool) (*dmarkets.Market, error) {
				return nil, dmarkets.ErrInvalidState
			}}, marketReviewAuthMock{admin: &dusers.User{Username: "admin"}}),
			body:       `{"confirm":true}`,
			wantStatus: http.StatusConflict,
			wantReason: handlers.ReasonInvalidState,
		},
		{
			name: "reject invalid input",
			handler: RejectMarketHandler(marketReviewServiceMock{rejectFn: func(context.Context, int64, string, string) (*dmarkets.Market, error) {
				return nil, dmarkets.ErrInvalidInput
			}}, marketReviewAuthMock{admin: &dusers.User{Username: "admin"}}),
			body:       `{"reason":""}`,
			wantStatus: http.StatusBadRequest,
			wantReason: handlers.ReasonValidationFailed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := mux.SetURLVars(httptest.NewRequest(http.MethodPatch, "/v0/admin/markets/42/review", bytes.NewBufferString(tt.body)), map[string]string{"id": "42"})
			rec := httptest.NewRecorder()

			tt.handler.ServeHTTP(rec, req)

			assertAdminFailure(t, rec, tt.wantStatus, tt.wantReason)
		})
	}
}

func assertAdminFailure(t *testing.T, rec *httptest.ResponseRecorder, wantStatus int, wantReason handlers.FailureReason) {
	t.Helper()
	if rec.Code != wantStatus {
		t.Fatalf("status = %d, want %d, body=%s", rec.Code, wantStatus, rec.Body.String())
	}
	var envelope handlers.FailureEnvelope
	if err := json.Unmarshal(rec.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("decode failure: %v", err)
	}
	if envelope.OK || envelope.Reason != string(wantReason) {
		t.Fatalf("unexpected failure envelope: %+v", envelope)
	}
}
