package adminhandlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"

	"socialpredict/handlers"
	dmarkets "socialpredict/internal/domain/markets"
	dusers "socialpredict/internal/domain/users"
)

type marketTagServiceMock struct {
	listFn   func(context.Context, bool) ([]dmarkets.MarketTag, error)
	createFn func(context.Context, dmarkets.MarketTagRequest, string) (*dmarkets.MarketTag, error)
	updateFn func(context.Context, string, dmarkets.MarketTagRequest) (*dmarkets.MarketTag, error)
}

func (m marketTagServiceMock) ListMarketTags(ctx context.Context, includeInactive bool) ([]dmarkets.MarketTag, error) {
	return m.listFn(ctx, includeInactive)
}

func (m marketTagServiceMock) CreateMarketTag(ctx context.Context, req dmarkets.MarketTagRequest, actorUsername string) (*dmarkets.MarketTag, error) {
	return m.createFn(ctx, req, actorUsername)
}

func (m marketTagServiceMock) UpdateMarketTag(ctx context.Context, slug string, req dmarkets.MarketTagRequest) (*dmarkets.MarketTag, error) {
	return m.updateFn(ctx, slug, req)
}

func TestListAdminMarketTagsHandlerReturnsTags(t *testing.T) {
	svc := marketTagServiceMock{
		listFn: func(_ context.Context, includeInactive bool) ([]dmarkets.MarketTag, error) {
			if !includeInactive {
				t.Fatalf("expected includeInactive")
			}
			return []dmarkets.MarketTag{{ID: 1, Slug: "sports", DisplayName: "Sports", IsActive: true}}, nil
		},
	}
	handler := ListAdminMarketTagsHandler(svc, marketReviewAuthMock{admin: &dusers.User{Username: "admin", UserType: string(dusers.UserTypeAdmin)}})
	req := httptest.NewRequest(http.MethodGet, "/v0/admin/market-tags?includeInactive=true", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body=%s", rec.Code, rec.Body.String())
	}
	var envelope handlers.SuccessEnvelope[marketTagsResponse]
	if err := json.Unmarshal(rec.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if !envelope.OK || envelope.Result.Total != 1 || envelope.Result.Tags[0].Slug != "sports" {
		t.Fatalf("unexpected response: %+v", envelope)
	}
}

func TestCreateAdminMarketTagHandlerPassesActor(t *testing.T) {
	svc := marketTagServiceMock{
		createFn: func(_ context.Context, req dmarkets.MarketTagRequest, actorUsername string) (*dmarkets.MarketTag, error) {
			if req.DisplayName != "Sports" || actorUsername != "admin" {
				t.Fatalf("unexpected create args req=%+v actor=%q", req, actorUsername)
			}
			return &dmarkets.MarketTag{ID: 2, Slug: "sports", DisplayName: req.DisplayName, IsActive: true, CreatedBy: actorUsername}, nil
		},
	}
	handler := CreateAdminMarketTagHandler(svc, marketReviewAuthMock{admin: &dusers.User{Username: "admin", UserType: string(dusers.UserTypeAdmin)}})
	req := httptest.NewRequest(http.MethodPost, "/v0/admin/market-tags", bytes.NewBufferString(`{"displayName":"Sports"}`))
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, body=%s", rec.Code, rec.Body.String())
	}
}

func TestUpdateAdminMarketTagHandlerRequiresDeactivateConfirmation(t *testing.T) {
	svc := marketTagServiceMock{
		updateFn: func(context.Context, string, dmarkets.MarketTagRequest) (*dmarkets.MarketTag, error) {
			t.Fatalf("update should not be called without deactivate confirmation")
			return nil, nil
		},
	}
	handler := UpdateAdminMarketTagHandler(svc, marketReviewAuthMock{admin: &dusers.User{Username: "admin", UserType: string(dusers.UserTypeAdmin)}})
	req := mux.SetURLVars(httptest.NewRequest(http.MethodPatch, "/v0/admin/market-tags/sports", bytes.NewBufferString(`{"isActive":false}`)), map[string]string{"slug": "sports"})
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	assertAdminFailure(t, rec, http.StatusBadRequest, handlers.ReasonValidationFailed)

	svc.updateFn = func(_ context.Context, slug string, req dmarkets.MarketTagRequest) (*dmarkets.MarketTag, error) {
		if slug != "sports" || req.IsActive == nil || *req.IsActive {
			t.Fatalf("unexpected update args slug=%q req=%+v", slug, req)
		}
		return &dmarkets.MarketTag{ID: 2, Slug: slug, DisplayName: "Sports", IsActive: false}, nil
	}
	handler = UpdateAdminMarketTagHandler(svc, marketReviewAuthMock{admin: &dusers.User{Username: "admin", UserType: string(dusers.UserTypeAdmin)}})
	req = mux.SetURLVars(httptest.NewRequest(http.MethodPatch, "/v0/admin/market-tags/sports", bytes.NewBufferString(`{"isActive":false,"confirmDeactivate":true}`)), map[string]string{"slug": "sports"})
	rec = httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body=%s", rec.Code, rec.Body.String())
	}
}
