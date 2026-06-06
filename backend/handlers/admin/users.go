package adminhandlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"

	"socialpredict/handlers"
	"socialpredict/handlers/authhttp"
	dusers "socialpredict/internal/domain/users"
	authsvc "socialpredict/internal/service/auth"
)

const defaultAdminUsersLimit = 100
const maxAdminUsersLimit = 250

type adminUserManager interface {
	ListUsers(ctx context.Context, filters dusers.ListFilters) ([]*dusers.User, error)
	PromoteToModerator(ctx context.Context, username, actorUsername, reason string) (*dusers.User, error)
	SuspendModerator(ctx context.Context, username, actorUsername, reason string, suspendedAt time.Time) (*dusers.User, error)
	UnsuspendModerator(ctx context.Context, username, actorUsername, reason string) (*dusers.User, error)
}

type adminUserResponse struct {
	ID                        int64   `json:"id"`
	Username                  string  `json:"username"`
	DisplayName               string  `json:"displayName"`
	UserType                  string  `json:"usertype"`
	ModeratorStatus           string  `json:"moderatorStatus"`
	AccountBalance            int64   `json:"accountBalance"`
	MustChangePassword        bool    `json:"mustChangePassword"`
	ModeratorSuspensionReason string  `json:"moderatorSuspensionReason,omitempty"`
	ModeratorSuspendedBy      string  `json:"moderatorSuspendedBy,omitempty"`
	ModeratorSuspendedAt      *string `json:"moderatorSuspendedAt,omitempty"`
	CreatedAt                 string  `json:"createdAt,omitempty"`
	UpdatedAt                 string  `json:"updatedAt,omitempty"`
}

type adminUsersResponse struct {
	Users  []adminUserResponse `json:"users"`
	Total  int                 `json:"total"`
	Limit  int                 `json:"limit"`
	Offset int                 `json:"offset"`
}

type adminRoleUpdateRequest struct {
	UserType string `json:"usertype"`
	Reason   string `json:"reason"`
}

type adminModeratorSuspensionRequest struct {
	Suspended bool   `json:"suspended"`
	Reason    string `json:"reason"`
}

func ListAdminUsersHandler(svc adminUserManager, auth authsvc.Authenticator) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			_ = handlers.WriteFailure(w, http.StatusMethodNotAllowed, handlers.ReasonMethodNotAllowed)
			return
		}
		if _, ok := requireAdminForUserManagement(w, r, auth); !ok {
			return
		}
		if svc == nil {
			_ = handlers.WriteFailure(w, http.StatusInternalServerError, handlers.ReasonInternalError)
			return
		}

		filters, ok := adminUserListFiltersFromRequest(w, r)
		if !ok {
			return
		}
		users, err := svc.ListUsers(r.Context(), filters)
		if err != nil {
			writeAdminUserError(w, err)
			return
		}

		responseUsers := make([]adminUserResponse, 0, len(users))
		for _, user := range users {
			responseUsers = append(responseUsers, adminUserResponseFromUser(user))
		}

		_ = handlers.WriteResult(w, http.StatusOK, adminUsersResponse{
			Users:  responseUsers,
			Total:  len(responseUsers),
			Limit:  filters.Limit,
			Offset: filters.Offset,
		})
	}
}

func UpdateAdminUserRoleHandler(svc adminUserManager, auth authsvc.Authenticator) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch {
			_ = handlers.WriteFailure(w, http.StatusMethodNotAllowed, handlers.ReasonMethodNotAllowed)
			return
		}
		admin, ok := requireAdminForUserManagement(w, r, auth)
		if !ok {
			return
		}
		if svc == nil {
			_ = handlers.WriteFailure(w, http.StatusInternalServerError, handlers.ReasonInternalError)
			return
		}
		username, ok := adminUsernameFromRequest(w, r)
		if !ok {
			return
		}

		var req adminRoleUpdateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			_ = handlers.WriteFailure(w, http.StatusBadRequest, handlers.ReasonInvalidRequest)
			return
		}
		if dusers.NormalizeUserType(req.UserType) != dusers.UserTypeModerator {
			_ = handlers.WriteFailure(w, http.StatusBadRequest, handlers.ReasonValidationFailed)
			return
		}

		user, err := svc.PromoteToModerator(r.Context(), username, admin.Username, req.Reason)
		if err != nil {
			writeAdminUserError(w, err)
			return
		}
		_ = handlers.WriteResult(w, http.StatusOK, adminUserResponseFromUser(user))
	}
}

func UpdateAdminModeratorSuspensionHandler(svc adminUserManager, auth authsvc.Authenticator, now func() time.Time) http.HandlerFunc {
	if now == nil {
		now = time.Now
	}

	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch {
			_ = handlers.WriteFailure(w, http.StatusMethodNotAllowed, handlers.ReasonMethodNotAllowed)
			return
		}
		admin, ok := requireAdminForUserManagement(w, r, auth)
		if !ok {
			return
		}
		if svc == nil {
			_ = handlers.WriteFailure(w, http.StatusInternalServerError, handlers.ReasonInternalError)
			return
		}
		username, ok := adminUsernameFromRequest(w, r)
		if !ok {
			return
		}

		var req adminModeratorSuspensionRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			_ = handlers.WriteFailure(w, http.StatusBadRequest, handlers.ReasonInvalidRequest)
			return
		}

		var (
			user *dusers.User
			err  error
		)
		if req.Suspended {
			user, err = svc.SuspendModerator(r.Context(), username, admin.Username, req.Reason, now().UTC())
		} else {
			user, err = svc.UnsuspendModerator(r.Context(), username, admin.Username, req.Reason)
		}
		if err != nil {
			writeAdminUserError(w, err)
			return
		}
		_ = handlers.WriteResult(w, http.StatusOK, adminUserResponseFromUser(user))
	}
}

func requireAdminForUserManagement(w http.ResponseWriter, r *http.Request, auth authsvc.Authenticator) (*dusers.User, bool) {
	if auth == nil {
		_ = handlers.WriteFailure(w, http.StatusInternalServerError, handlers.ReasonInternalError)
		return nil, false
	}
	admin, authErr := auth.RequireAdmin(r)
	if authErr != nil {
		_ = authhttp.WriteFailure(w, authErr)
		return nil, false
	}
	return admin, true
}

func adminUserListFiltersFromRequest(w http.ResponseWriter, r *http.Request) (dusers.ListFilters, bool) {
	query := r.URL.Query()
	limit, ok := parseAdminUserListInt(w, query.Get("limit"), defaultAdminUsersLimit, 1, maxAdminUsersLimit)
	if !ok {
		return dusers.ListFilters{}, false
	}
	offset, ok := parseAdminUserListInt(w, query.Get("offset"), 0, 0, 100000)
	if !ok {
		return dusers.ListFilters{}, false
	}

	return dusers.ListFilters{
		UserType: strings.TrimSpace(query.Get("usertype")),
		Query:    strings.TrimSpace(query.Get("query")),
		Limit:    limit,
		Offset:   offset,
	}, true
}

func parseAdminUserListInt(w http.ResponseWriter, raw string, fallback, min, max int) (int, bool) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return fallback, true
	}
	value, err := strconv.Atoi(raw)
	if err != nil || value < min || value > max {
		_ = handlers.WriteFailure(w, http.StatusBadRequest, handlers.ReasonInvalidRequest)
		return 0, false
	}
	return value, true
}

func adminUsernameFromRequest(w http.ResponseWriter, r *http.Request) (string, bool) {
	username := strings.TrimSpace(mux.Vars(r)["username"])
	if username == "" {
		_ = handlers.WriteFailure(w, http.StatusBadRequest, handlers.ReasonInvalidRequest)
		return "", false
	}
	return username, true
}

func writeAdminUserError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, dusers.ErrUserNotFound):
		_ = handlers.WriteFailure(w, http.StatusNotFound, handlers.ReasonUserNotFound)
	case errors.Is(err, dusers.ErrUnauthorized):
		_ = handlers.WriteFailure(w, http.StatusForbidden, handlers.ReasonAuthorizationDenied)
	case errors.Is(err, dusers.ErrInvalidModeratorState):
		_ = handlers.WriteFailure(w, http.StatusConflict, handlers.ReasonInvalidState)
	case errors.Is(err, dusers.ErrInvalidUserData):
		_ = handlers.WriteFailure(w, http.StatusBadRequest, handlers.ReasonValidationFailed)
	default:
		_ = handlers.WriteFailure(w, http.StatusInternalServerError, handlers.ReasonInternalError)
	}
}

func adminUserResponseFromUser(user *dusers.User) adminUserResponse {
	if user == nil {
		return adminUserResponse{}
	}

	response := adminUserResponse{
		ID:                        user.ID,
		Username:                  user.Username,
		DisplayName:               user.DisplayName,
		UserType:                  user.UserType,
		ModeratorStatus:           string(dusers.NormalizeModeratorStatus(user.UserType, string(user.ModeratorStatus))),
		AccountBalance:            user.AccountBalance,
		MustChangePassword:        user.MustChangePassword,
		ModeratorSuspensionReason: user.ModeratorSuspensionReason,
		ModeratorSuspendedBy:      user.ModeratorSuspendedBy,
		CreatedAt:                 formatAdminTime(user.CreatedAt),
		UpdatedAt:                 formatAdminTime(user.UpdatedAt),
	}
	if user.ModeratorSuspendedAt != nil {
		value := formatAdminTime(*user.ModeratorSuspendedAt)
		response.ModeratorSuspendedAt = &value
	}
	return response
}

func formatAdminTime(value time.Time) string {
	if value.IsZero() {
		return ""
	}
	return value.UTC().Format(time.RFC3339)
}
