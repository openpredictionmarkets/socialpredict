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
	approveFn  func(context.Context, int64, string, bool) (*dmarkets.Market, error)
	rejectFn   func(context.Context, int64, string, string) (*dmarkets.Market, error)
	listFn     func(context.Context, dmarkets.ListFilters) ([]*dmarkets.Market, error)
	reassignFn func(context.Context, int64, string, string, string) (*dmarkets.Market, error)
	tagsFn     func(context.Context, int64, []string, string) (*dmarkets.Market, error)
}

func (m marketReviewServiceMock) ApproveProposedMarket(ctx context.Context, marketID int64, actorUsername string, confirmed bool) (*dmarkets.Market, error) {
	return m.approveFn(ctx, marketID, actorUsername, confirmed)
}

func (m marketReviewServiceMock) RejectProposedMarket(ctx context.Context, marketID int64, actorUsername string, reason string) (*dmarkets.Market, error) {
	return m.rejectFn(ctx, marketID, actorUsername, reason)
}

func (m marketReviewServiceMock) ListLifecycleMarkets(ctx context.Context, filters dmarkets.ListFilters) ([]*dmarkets.Market, error) {
	return m.listFn(ctx, filters)
}

func (m marketReviewServiceMock) ReassignMarketSteward(ctx context.Context, marketID int64, newStewardUsername string, actorUsername string, reason string) (*dmarkets.Market, error) {
	return m.reassignFn(ctx, marketID, newStewardUsername, actorUsername, reason)
}

func (m marketReviewServiceMock) UpdateMarketTags(ctx context.Context, marketID int64, tagSlugs []string, actorUsername string) (*dmarkets.Market, error) {
	return m.tagsFn(ctx, marketID, tagSlugs, actorUsername)
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

func TestListReviewMarketsHandlerReturnsQueue(t *testing.T) {
	changedAt := time.Date(2026, 6, 4, 12, 0, 0, 0, time.UTC)
	svc := marketReviewServiceMock{
		listFn: func(_ context.Context, filters dmarkets.ListFilters) ([]*dmarkets.Market, error) {
			if filters.Status != dmarkets.MarketLifecycleProposed || filters.Limit != 25 || filters.Offset != 5 {
				t.Fatalf("unexpected filters: %+v", filters)
			}
			return []*dmarkets.Market{{
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
			}}, nil
		},
	}
	handler := ListReviewMarketsHandler(svc, marketReviewAuthMock{admin: &dusers.User{Username: "admin", UserType: string(dusers.UserTypeAdmin)}})
	req := httptest.NewRequest(http.MethodGet, "/v0/admin/markets?status=proposed&limit=25&offset=5", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body=%s", rec.Code, rec.Body.String())
	}
	var envelope handlers.SuccessEnvelope[marketReviewListResponse]
	if err := json.Unmarshal(rec.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if !envelope.OK || envelope.Result.Total != 1 || envelope.Result.Markets[0].CreatorUsername != "moderator" {
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

func TestListReviewMarketsHandlerSupportsAllStatusSearch(t *testing.T) {
	svc := marketReviewServiceMock{
		listFn: func(_ context.Context, filters dmarkets.ListFilters) ([]*dmarkets.Market, error) {
			if filters.Status != dmarkets.MarketStatusAll || filters.Query != "orchard" || filters.Limit != 100 || filters.Offset != 0 {
				t.Fatalf("unexpected filters: %+v", filters)
			}
			return []*dmarkets.Market{{
				ID:              45,
				QuestionTitle:   "Orchard market",
				CreatorUsername: "moderator",
				Status:          dmarkets.MarketStatusResolved,
				LifecycleStatus: dmarkets.MarketLifecycleResolved,
			}}, nil
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
