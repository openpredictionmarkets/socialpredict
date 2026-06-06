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
	dmarkets "socialpredict/internal/domain/markets"
	dusers "socialpredict/internal/domain/users"
	authsvc "socialpredict/internal/service/auth"
)

type marketReviewer interface {
	ApproveProposedMarket(ctx context.Context, marketID int64, actorUsername string, confirmed bool) (*dmarkets.Market, error)
	RejectProposedMarket(ctx context.Context, marketID int64, actorUsername string, reason string) (*dmarkets.Market, error)
}

type marketReviewLister interface {
	ListLifecycleMarkets(ctx context.Context, filters dmarkets.ListFilters) ([]*dmarkets.Market, error)
}

type marketStewardReassigner interface {
	ReassignMarketSteward(ctx context.Context, marketID int64, newStewardUsername string, actorUsername string, reason string) (*dmarkets.Market, error)
}

type marketTagAdjuster interface {
	UpdateMarketTags(ctx context.Context, marketID int64, tagSlugs []string, actorUsername string) (*dmarkets.Market, error)
}

type approveMarketRequest struct {
	Confirm bool `json:"confirm"`
}

type rejectMarketRequest struct {
	Reason string `json:"reason"`
}

type reassignMarketStewardRequest struct {
	StewardUsername string `json:"stewardUsername"`
	Reason          string `json:"reason"`
}

type updateMarketTagsRequest struct {
	TagSlugs []string `json:"tagSlugs"`
}

type marketReviewResponse struct {
	ID                 int64                            `json:"id"`
	QuestionTitle      string                           `json:"questionTitle,omitempty"`
	Description        string                           `json:"description,omitempty"`
	CreatorUsername    string                           `json:"creatorUsername,omitempty"`
	StewardUsername    string                           `json:"stewardUsername,omitempty"`
	YesLabel           string                           `json:"yesLabel,omitempty"`
	NoLabel            string                           `json:"noLabel,omitempty"`
	Status             string                           `json:"status"`
	LifecycleStatus    string                           `json:"lifecycleStatus"`
	ApprovedBy         string                           `json:"approvedBy,omitempty"`
	ApprovedAt         *time.Time                       `json:"approvedAt,omitempty"`
	RejectedBy         string                           `json:"rejectedBy,omitempty"`
	RejectedAt         *time.Time                       `json:"rejectedAt,omitempty"`
	RejectionReason    string                           `json:"rejectionReason,omitempty"`
	ProposalCost       int64                            `json:"proposalCost,omitempty"`
	StewardshipAudits  []marketStewardshipAuditResponse `json:"stewardshipAudits,omitempty"`
	Tags               []marketTagResponse              `json:"tags,omitempty"`
	CreatedAt          time.Time                        `json:"createdAt,omitempty"`
	UpdatedAt          time.Time                        `json:"updatedAt,omitempty"`
	ResolutionDateTime time.Time                        `json:"resolutionDateTime,omitempty"`
}

type marketStewardshipAuditResponse struct {
	ID                  int64     `json:"id,omitempty"`
	MarketID            int64     `json:"marketId"`
	FromStewardUsername string    `json:"fromStewardUsername,omitempty"`
	ToStewardUsername   string    `json:"toStewardUsername"`
	ActorUsername       string    `json:"actorUsername"`
	Reason              string    `json:"reason,omitempty"`
	CreatedAt           time.Time `json:"createdAt,omitempty"`
}

type marketReviewListResponse struct {
	Markets []marketReviewResponse `json:"markets"`
	Total   int                    `json:"total"`
}

func ApproveMarketHandler(svc marketReviewer, auth authsvc.Authenticator) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch {
			_ = handlers.WriteFailure(w, http.StatusMethodNotAllowed, handlers.ReasonMethodNotAllowed)
			return
		}
		admin, ok := requireAdminForMarketReview(w, r, auth)
		if !ok {
			return
		}
		if svc == nil {
			_ = handlers.WriteFailure(w, http.StatusInternalServerError, handlers.ReasonInternalError)
			return
		}
		marketID, ok := marketIDFromRequest(w, r)
		if !ok {
			return
		}
		var req approveMarketRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			_ = handlers.WriteFailure(w, http.StatusBadRequest, handlers.ReasonInvalidRequest)
			return
		}

		market, err := svc.ApproveProposedMarket(r.Context(), marketID, admin.Username, req.Confirm)
		if err != nil {
			writeMarketReviewError(w, err)
			return
		}
		_ = handlers.WriteResult(w, http.StatusOK, marketReviewResponseFromMarket(market))
	}
}

func ListReviewMarketsHandler(svc marketReviewLister, auth authsvc.Authenticator) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			_ = handlers.WriteFailure(w, http.StatusMethodNotAllowed, handlers.ReasonMethodNotAllowed)
			return
		}
		if _, ok := requireAdminForMarketReview(w, r, auth); !ok {
			return
		}
		if svc == nil {
			_ = handlers.WriteFailure(w, http.StatusInternalServerError, handlers.ReasonInternalError)
			return
		}

		filters, ok := parseAdminReviewMarketFilters(w, r)
		if !ok {
			return
		}
		markets, err := svc.ListLifecycleMarkets(r.Context(), filters)
		if err != nil {
			writeMarketReviewError(w, err)
			return
		}

		response := marketReviewListResponse{
			Markets: make([]marketReviewResponse, 0, len(markets)),
			Total:   len(markets),
		}
		for _, market := range markets {
			response.Markets = append(response.Markets, marketReviewResponseFromMarket(market))
		}
		_ = handlers.WriteResult(w, http.StatusOK, response)
	}
}

func RejectMarketHandler(svc marketReviewer, auth authsvc.Authenticator) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch {
			_ = handlers.WriteFailure(w, http.StatusMethodNotAllowed, handlers.ReasonMethodNotAllowed)
			return
		}
		admin, ok := requireAdminForMarketReview(w, r, auth)
		if !ok {
			return
		}
		if svc == nil {
			_ = handlers.WriteFailure(w, http.StatusInternalServerError, handlers.ReasonInternalError)
			return
		}
		marketID, ok := marketIDFromRequest(w, r)
		if !ok {
			return
		}
		var req rejectMarketRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			_ = handlers.WriteFailure(w, http.StatusBadRequest, handlers.ReasonInvalidRequest)
			return
		}

		market, err := svc.RejectProposedMarket(r.Context(), marketID, admin.Username, req.Reason)
		if err != nil {
			writeMarketReviewError(w, err)
			return
		}
		_ = handlers.WriteResult(w, http.StatusOK, marketReviewResponseFromMarket(market))
	}
}

func ReassignMarketStewardHandler(svc marketStewardReassigner, auth authsvc.Authenticator) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch {
			_ = handlers.WriteFailure(w, http.StatusMethodNotAllowed, handlers.ReasonMethodNotAllowed)
			return
		}
		admin, ok := requireAdminForMarketReview(w, r, auth)
		if !ok {
			return
		}
		if svc == nil {
			_ = handlers.WriteFailure(w, http.StatusInternalServerError, handlers.ReasonInternalError)
			return
		}
		marketID, ok := marketIDFromRequest(w, r)
		if !ok {
			return
		}
		var req reassignMarketStewardRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			_ = handlers.WriteFailure(w, http.StatusBadRequest, handlers.ReasonInvalidRequest)
			return
		}

		market, err := svc.ReassignMarketSteward(r.Context(), marketID, req.StewardUsername, admin.Username, req.Reason)
		if err != nil {
			writeMarketReviewError(w, err)
			return
		}
		_ = handlers.WriteResult(w, http.StatusOK, marketReviewResponseFromMarket(market))
	}
}

func UpdateMarketTagsHandler(svc marketTagAdjuster, auth authsvc.Authenticator) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch {
			_ = handlers.WriteFailure(w, http.StatusMethodNotAllowed, handlers.ReasonMethodNotAllowed)
			return
		}
		admin, ok := requireAdminForMarketReview(w, r, auth)
		if !ok {
			return
		}
		if svc == nil {
			_ = handlers.WriteFailure(w, http.StatusInternalServerError, handlers.ReasonInternalError)
			return
		}
		marketID, ok := marketIDFromRequest(w, r)
		if !ok {
			return
		}
		var req updateMarketTagsRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			_ = handlers.WriteFailure(w, http.StatusBadRequest, handlers.ReasonInvalidRequest)
			return
		}

		market, err := svc.UpdateMarketTags(r.Context(), marketID, req.TagSlugs, admin.Username)
		if err != nil {
			writeMarketReviewError(w, err)
			return
		}
		_ = handlers.WriteResult(w, http.StatusOK, marketReviewResponseFromMarket(market))
	}
}

func requireAdminForMarketReview(w http.ResponseWriter, r *http.Request, auth authsvc.Authenticator) (*dusers.User, bool) {
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

func marketIDFromRequest(w http.ResponseWriter, r *http.Request) (int64, bool) {
	id, err := strconv.ParseInt(mux.Vars(r)["id"], 10, 64)
	if err != nil || id <= 0 {
		_ = handlers.WriteFailure(w, http.StatusBadRequest, handlers.ReasonInvalidRequest)
		return 0, false
	}
	return id, true
}

func writeMarketReviewError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, dmarkets.ErrMarketNotFound):
		_ = handlers.WriteFailure(w, http.StatusNotFound, handlers.ReasonMarketNotFound)
	case errors.Is(err, dmarkets.ErrUserNotFound):
		_ = handlers.WriteFailure(w, http.StatusNotFound, handlers.ReasonUserNotFound)
	case errors.Is(err, dmarkets.ErrUnauthorized):
		_ = handlers.WriteFailure(w, http.StatusForbidden, handlers.ReasonAuthorizationDenied)
	case errors.Is(err, dmarkets.ErrInvalidState):
		_ = handlers.WriteFailure(w, http.StatusConflict, handlers.ReasonInvalidState)
	case errors.Is(err, dmarkets.ErrInvalidInput):
		_ = handlers.WriteFailure(w, http.StatusBadRequest, handlers.ReasonValidationFailed)
	default:
		_ = handlers.WriteFailure(w, http.StatusInternalServerError, handlers.ReasonInternalError)
	}
}

func marketReviewResponseFromMarket(market *dmarkets.Market) marketReviewResponse {
	if market == nil {
		return marketReviewResponse{}
	}
	return marketReviewResponse{
		ID:                 market.ID,
		QuestionTitle:      market.QuestionTitle,
		Description:        market.Description,
		CreatorUsername:    market.CreatorUsername,
		StewardUsername:    market.CurrentStewardUsername(),
		YesLabel:           market.YesLabel,
		NoLabel:            market.NoLabel,
		Status:             market.Status,
		LifecycleStatus:    market.LifecycleStatus,
		ApprovedBy:         market.ApprovedBy,
		ApprovedAt:         market.ApprovedAt,
		RejectedBy:         market.RejectedBy,
		RejectedAt:         market.RejectedAt,
		RejectionReason:    market.RejectionReason,
		ProposalCost:       market.ProposalCost,
		StewardshipAudits:  marketStewardshipAuditResponsesFromRecords(market.StewardshipAudits),
		Tags:               marketTagResponses(market.Tags),
		CreatedAt:          market.CreatedAt,
		UpdatedAt:          market.UpdatedAt,
		ResolutionDateTime: market.ResolutionDateTime,
	}
}

func marketStewardshipAuditResponsesFromRecords(records []dmarkets.MarketStewardshipAuditRecord) []marketStewardshipAuditResponse {
	if len(records) == 0 {
		return nil
	}
	responses := make([]marketStewardshipAuditResponse, 0, len(records))
	for _, record := range records {
		responses = append(responses, marketStewardshipAuditResponse{
			ID:                  record.ID,
			MarketID:            record.MarketID,
			FromStewardUsername: record.FromStewardUsername,
			ToStewardUsername:   record.ToStewardUsername,
			ActorUsername:       record.ActorUsername,
			Reason:              record.Reason,
			CreatedAt:           record.CreatedAt,
		})
	}
	return responses
}

func parseAdminReviewMarketFilters(w http.ResponseWriter, r *http.Request) (dmarkets.ListFilters, bool) {
	query := r.URL.Query()
	status := dmarkets.NormalizeLifecycleStatus(query.Get("status"))
	if status == "" {
		status = dmarkets.MarketLifecycleProposed
	}
	switch status {
	case dmarkets.MarketStatusAll, dmarkets.MarketLifecycleProposed, dmarkets.MarketLifecyclePublished, dmarkets.MarketLifecycleRejected, dmarkets.MarketLifecycleClosed, dmarkets.MarketLifecycleResolved:
	default:
		_ = handlers.WriteFailure(w, http.StatusBadRequest, handlers.ReasonInvalidRequest)
		return dmarkets.ListFilters{}, false
	}

	searchQuery := strings.TrimSpace(query.Get("query"))
	if len(searchQuery) > 100 {
		_ = handlers.WriteFailure(w, http.StatusBadRequest, handlers.ReasonInvalidRequest)
		return dmarkets.ListFilters{}, false
	}

	return dmarkets.ListFilters{
		Status: status,
		Query:  searchQuery,
		Limit:  boundedAdminReviewInt(query.Get("limit"), 50, 1, 100),
		Offset: boundedAdminReviewInt(query.Get("offset"), 0, 0, 100000),
	}, true
}

func boundedAdminReviewInt(value string, fallback int, min int, max int) int {
	if value == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}
	if parsed < min {
		return min
	}
	if parsed > max {
		return max
	}
	return parsed
}
