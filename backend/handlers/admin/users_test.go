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
	dusers "socialpredict/internal/domain/users"
)

type adminUserManagerMock struct {
	listFn      func(context.Context, dusers.ListFilters) ([]*dusers.User, error)
	promoteFn   func(context.Context, string, string, string) (*dusers.User, error)
	suspendFn   func(context.Context, string, string, string, time.Time) (*dusers.User, error)
	unsuspendFn func(context.Context, string, string, string) (*dusers.User, error)
}

func (m adminUserManagerMock) ListUsers(ctx context.Context, filters dusers.ListFilters) ([]*dusers.User, error) {
	return m.listFn(ctx, filters)
}

func (m adminUserManagerMock) PromoteToModerator(ctx context.Context, username, actorUsername, reason string) (*dusers.User, error) {
	return m.promoteFn(ctx, username, actorUsername, reason)
}

func (m adminUserManagerMock) SuspendModerator(ctx context.Context, username, actorUsername, reason string, suspendedAt time.Time) (*dusers.User, error) {
	return m.suspendFn(ctx, username, actorUsername, reason, suspendedAt)
}

func (m adminUserManagerMock) UnsuspendModerator(ctx context.Context, username, actorUsername, reason string) (*dusers.User, error) {
	return m.unsuspendFn(ctx, username, actorUsername, reason)
}

func TestListAdminUsersHandlerReturnsUserQueue(t *testing.T) {
	svc := adminUserManagerMock{
		listFn: func(_ context.Context, filters dusers.ListFilters) ([]*dusers.User, error) {
			if filters.Limit != 50 || filters.Offset != 10 {
				t.Fatalf("unexpected filters: %+v", filters)
			}
			return []*dusers.User{
				{ID: 1, Username: "regular", DisplayName: "Regular User", UserType: string(dusers.UserTypeRegular), ModeratorStatus: dusers.ModeratorStatusNone},
				{ID: 2, Username: "moderator", DisplayName: "Moderator User", UserType: string(dusers.UserTypeModerator), ModeratorStatus: dusers.ModeratorStatusActive},
			}, nil
		},
	}
	handler := ListAdminUsersHandler(svc, marketReviewAuthMock{admin: &dusers.User{Username: "admin", UserType: string(dusers.UserTypeAdmin)}})
	req := httptest.NewRequest(http.MethodGet, "/v0/admin/users?limit=50&offset=10", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body=%s", rec.Code, rec.Body.String())
	}
	var envelope handlers.SuccessEnvelope[adminUsersResponse]
	if err := json.Unmarshal(rec.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if !envelope.OK || envelope.Result.Total != 2 {
		t.Fatalf("unexpected envelope: %+v", envelope)
	}
	if envelope.Result.Users[1].UserType != string(dusers.UserTypeModerator) || envelope.Result.Users[1].ModeratorStatus != string(dusers.ModeratorStatusActive) {
		t.Fatalf("moderator state missing from response: %+v", envelope.Result.Users[1])
	}
}

func TestUpdateAdminUserRoleHandlerPromotesModerator(t *testing.T) {
	svc := adminUserManagerMock{
		promoteFn: func(_ context.Context, username, actorUsername, reason string) (*dusers.User, error) {
			if username != "candidate" || actorUsername != "admin" || reason != "trusted" {
				t.Fatalf("unexpected promote args: username=%q actor=%q reason=%q", username, actorUsername, reason)
			}
			return &dusers.User{ID: 3, Username: username, UserType: string(dusers.UserTypeModerator), ModeratorStatus: dusers.ModeratorStatusActive}, nil
		},
	}
	handler := UpdateAdminUserRoleHandler(svc, marketReviewAuthMock{admin: &dusers.User{Username: "admin", UserType: string(dusers.UserTypeAdmin)}})
	req := mux.SetURLVars(
		httptest.NewRequest(http.MethodPatch, "/v0/admin/users/candidate/role", bytes.NewBufferString(`{"usertype":"MODERATOR","reason":"trusted"}`)),
		map[string]string{"username": "candidate"},
	)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body=%s", rec.Code, rec.Body.String())
	}
	var envelope handlers.SuccessEnvelope[adminUserResponse]
	if err := json.Unmarshal(rec.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if envelope.Result.UserType != string(dusers.UserTypeModerator) || envelope.Result.ModeratorStatus != string(dusers.ModeratorStatusActive) {
		t.Fatalf("unexpected response: %+v", envelope.Result)
	}
}

func TestUpdateAdminModeratorSuspensionHandlerSuspendsAndUnsuspends(t *testing.T) {
	fixedNow := time.Date(2026, 5, 24, 20, 0, 0, 0, time.UTC)

	t.Run("suspend", func(t *testing.T) {
		svc := adminUserManagerMock{
			suspendFn: func(_ context.Context, username, actorUsername, reason string, suspendedAt time.Time) (*dusers.User, error) {
				if username != "moderator" || actorUsername != "admin" || reason != "policy" || !suspendedAt.Equal(fixedNow) {
					t.Fatalf("unexpected suspend args: username=%q actor=%q reason=%q at=%s", username, actorUsername, reason, suspendedAt)
				}
				return &dusers.User{Username: username, UserType: string(dusers.UserTypeModerator), ModeratorStatus: dusers.ModeratorStatusSuspended, ModeratorSuspensionReason: reason, ModeratorSuspendedBy: actorUsername, ModeratorSuspendedAt: &suspendedAt}, nil
			},
		}
		handler := UpdateAdminModeratorSuspensionHandler(svc, marketReviewAuthMock{admin: &dusers.User{Username: "admin", UserType: string(dusers.UserTypeAdmin)}}, func() time.Time { return fixedNow })
		req := mux.SetURLVars(
			httptest.NewRequest(http.MethodPatch, "/v0/admin/moderators/moderator/suspension", bytes.NewBufferString(`{"suspended":true,"reason":"policy"}`)),
			map[string]string{"username": "moderator"},
		)
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, body=%s", rec.Code, rec.Body.String())
		}
	})

	t.Run("unsuspend", func(t *testing.T) {
		svc := adminUserManagerMock{
			unsuspendFn: func(_ context.Context, username, actorUsername, reason string) (*dusers.User, error) {
				if username != "moderator" || actorUsername != "admin" || reason != "appeal" {
					t.Fatalf("unexpected unsuspend args: username=%q actor=%q reason=%q", username, actorUsername, reason)
				}
				return &dusers.User{Username: username, UserType: string(dusers.UserTypeModerator), ModeratorStatus: dusers.ModeratorStatusActive}, nil
			},
		}
		handler := UpdateAdminModeratorSuspensionHandler(svc, marketReviewAuthMock{admin: &dusers.User{Username: "admin", UserType: string(dusers.UserTypeAdmin)}}, func() time.Time { return fixedNow })
		req := mux.SetURLVars(
			httptest.NewRequest(http.MethodPatch, "/v0/admin/moderators/moderator/suspension", bytes.NewBufferString(`{"suspended":false,"reason":"appeal"}`)),
			map[string]string{"username": "moderator"},
		)
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, body=%s", rec.Code, rec.Body.String())
		}
	})
}
