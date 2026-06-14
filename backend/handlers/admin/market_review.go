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

type marketGroupReviewer interface {
	ApproveProposedMarketGroup(ctx context.Context, groupID int64, actorUsername string, confirmed bool) (*dmarkets.MarketGroup, error)
	RejectProposedMarketGroup(ctx context.Context, groupID int64, actorUsername string, reason string) (*dmarkets.MarketGroup, error)
}

type marketReviewLister interface {
	ListLifecycleMarkets(ctx context.Context, filters dmarkets.ListFilters) ([]*dmarkets.Market, error)
}

type marketStewardReassigner interface {
	ReassignMarketSteward(ctx context.Context, marketID int64, newStewardUsername string, actorUsername string, reason string) (*dmarkets.Market, error)
}

type marketGroupStewardReassigner interface {
	ReassignMarketGroupSteward(ctx context.Context, groupID int64, newStewardUsername string, actorUsername string, reason string) (*dmarkets.MarketGroup, error)
}

type marketTagAdjuster interface {
	UpdateMarketTags(ctx context.Context, marketID int64, tagSlugs []string, actorUsername string) (*dmarkets.Market, error)
}

type marketGroupLookup interface {
	GetMarketGroupForMarket(ctx context.Context, marketID int64) (*dmarkets.MarketGroup, error)
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
	MarketGroup        *marketGroupReviewLink           `json:"marketGroup,omitempty"`
	CreatedAt          time.Time                        `json:"createdAt,omitempty"`
	UpdatedAt          time.Time                        `json:"updatedAt,omitempty"`
	ResolutionDateTime time.Time                        `json:"resolutionDateTime,omitempty"`
}

type marketGroupReviewLink struct {
	ID              int64  `json:"id"`
	QuestionTitle   string `json:"questionTitle"`
	Description     string `json:"description,omitempty"`
	GroupType       string `json:"groupType"`
	LifecycleStatus string `json:"lifecycleStatus"`
	Status          string `json:"status"`
	AnswerLabel     string `json:"answerLabel,omitempty"`
	AnswerCount     int    `json:"answerCount"`
	ProposalCost    int64  `json:"proposalCost,omitempty"`
	CreatorUsername string `json:"creatorUsername,omitempty"`
	StewardUsername string `json:"stewardUsername,omitempty"`
}

type marketGroupReviewResponse struct {
	ID                 int64      `json:"id"`
	QuestionTitle      string     `json:"questionTitle"`
	Description        string     `json:"description,omitempty"`
	GroupType          string     `json:"groupType"`
	ProbabilityPolicy  string     `json:"probabilityPolicy"`
	ResolutionPolicy   string     `json:"resolutionPolicy"`
	LifecycleStatus    string     `json:"lifecycleStatus"`
	Status             string     `json:"status"`
	ProposalCost       int64      `json:"proposalCost,omitempty"`
	CreatorUsername    string     `json:"creatorUsername,omitempty"`
	StewardUsername    string     `json:"stewardUsername,omitempty"`
	ApprovedBy         string     `json:"approvedBy,omitempty"`
	ApprovedAt         *time.Time `json:"approvedAt,omitempty"`
	RejectedBy         string     `json:"rejectedBy,omitempty"`
	RejectedAt         *time.Time `json:"rejectedAt,omitempty"`
	RejectionReason    string     `json:"rejectionReason,omitempty"`
	ResolutionDateTime time.Time  `json:"resolutionDateTime,omitempty"`
	CreatedAt          time.Time  `json:"createdAt,omitempty"`
	UpdatedAt          time.Time  `json:"updatedAt,omitempty"`
	AnswerCount        int        `json:"answerCount"`
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
		_ = handlers.WriteResult(w, http.StatusOK, marketReviewResponseFromMarket(r.Context(), market, nil))
	}
}

func ApproveMarketGroupHandler(svc marketGroupReviewer, auth authsvc.Authenticator) http.HandlerFunc {
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
		groupID, ok := marketGroupIDFromRequest(w, r)
		if !ok {
			return
		}
		var req approveMarketRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			_ = handlers.WriteFailure(w, http.StatusBadRequest, handlers.ReasonInvalidRequest)
			return
		}

		group, err := svc.ApproveProposedMarketGroup(r.Context(), groupID, admin.Username, req.Confirm)
		if err != nil {
			writeMarketReviewError(w, err)
			return
		}
		_ = handlers.WriteResult(w, http.StatusOK, marketGroupReviewResponseFromGroup(group))
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

		var lookup marketGroupLookup
		if groupLookup, ok := svc.(marketGroupLookup); ok {
			lookup = groupLookup
		}

		response := marketReviewListResponse{
			Markets: make([]marketReviewResponse, 0, len(markets)),
			Total:   len(markets),
		}
		for _, market := range markets {
			response.Markets = append(response.Markets, marketReviewResponseFromMarket(r.Context(), market, lookup))
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
		_ = handlers.WriteResult(w, http.StatusOK, marketReviewResponseFromMarket(r.Context(), market, nil))
	}
}

func RejectMarketGroupHandler(svc marketGroupReviewer, auth authsvc.Authenticator) http.HandlerFunc {
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
		groupID, ok := marketGroupIDFromRequest(w, r)
		if !ok {
			return
		}
		var req rejectMarketRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			_ = handlers.WriteFailure(w, http.StatusBadRequest, handlers.ReasonInvalidRequest)
			return
		}

		group, err := svc.RejectProposedMarketGroup(r.Context(), groupID, admin.Username, req.Reason)
		if err != nil {
			writeMarketReviewError(w, err)
			return
		}
		_ = handlers.WriteResult(w, http.StatusOK, marketGroupReviewResponseFromGroup(group))
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
		_ = handlers.WriteResult(w, http.StatusOK, marketReviewResponseFromMarket(r.Context(), market, nil))
	}
}

func ReassignMarketGroupStewardHandler(svc marketGroupStewardReassigner, auth authsvc.Authenticator) http.HandlerFunc {
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
		groupID, ok := marketGroupIDFromRequest(w, r)
		if !ok {
			return
		}
		var req reassignMarketStewardRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			_ = handlers.WriteFailure(w, http.StatusBadRequest, handlers.ReasonInvalidRequest)
			return
		}

		group, err := svc.ReassignMarketGroupSteward(r.Context(), groupID, req.StewardUsername, admin.Username, req.Reason)
		if err != nil {
			writeMarketReviewError(w, err)
			return
		}
		_ = handlers.WriteResult(w, http.StatusOK, marketGroupReviewResponseFromGroup(group))
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
		_ = handlers.WriteResult(w, http.StatusOK, marketReviewResponseFromMarket(r.Context(), market, nil))
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

func marketGroupIDFromRequest(w http.ResponseWriter, r *http.Request) (int64, bool) {
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
	case errors.Is(err, dmarkets.ErrInsufficientBalance):
		_ = handlers.WriteFailure(w, http.StatusUnprocessableEntity, handlers.ReasonInsufficientBalance)
	case errors.Is(err, dmarkets.ErrInvalidState):
		_ = handlers.WriteFailure(w, http.StatusConflict, handlers.ReasonInvalidState)
	case errors.Is(err, dmarkets.ErrInvalidInput):
		_ = handlers.WriteFailure(w, http.StatusBadRequest, handlers.ReasonValidationFailed)
	default:
		_ = handlers.WriteFailure(w, http.StatusInternalServerError, handlers.ReasonInternalError)
	}
}

func marketReviewResponseFromMarket(ctx context.Context, market *dmarkets.Market, lookup marketGroupLookup) marketReviewResponse {
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
		MarketGroup:        marketGroupReviewLinkForMarket(ctx, market.ID, lookup),
		CreatedAt:          market.CreatedAt,
		UpdatedAt:          market.UpdatedAt,
		ResolutionDateTime: market.ResolutionDateTime,
	}
}

func marketGroupReviewLinkForMarket(ctx context.Context, marketID int64, lookup marketGroupLookup) *marketGroupReviewLink {
	if lookup == nil || marketID <= 0 {
		return nil
	}
	group, err := lookup.GetMarketGroupForMarket(ctx, marketID)
	if err != nil || group == nil {
		return nil
	}
	return marketGroupReviewLinkFromGroup(group, marketID)
}

func marketGroupReviewLinkFromGroup(group *dmarkets.MarketGroup, marketID int64) *marketGroupReviewLink {
	if group == nil || group.ID <= 0 {
		return nil
	}
	status := group.LifecycleStatus
	if status == dmarkets.MarketLifecyclePublished {
		status = dmarkets.MarketStatusActive
	}
	link := &marketGroupReviewLink{
		ID:              group.ID,
		QuestionTitle:   group.QuestionTitle,
		Description:     group.Description,
		GroupType:       group.GroupType,
		LifecycleStatus: group.LifecycleStatus,
		Status:          status,
		AnswerCount:     len(group.Members),
		ProposalCost:    group.ProposalCost,
		CreatorUsername: group.CreatorUsername,
		StewardUsername: group.StewardUsername,
	}
	for _, member := range group.Members {
		if member.MarketID == marketID {
			link.AnswerLabel = member.AnswerLabel
			break
		}
	}
	return link
}

func marketGroupReviewResponseFromGroup(group *dmarkets.MarketGroup) marketGroupReviewResponse {
	if group == nil {
		return marketGroupReviewResponse{}
	}
	status := group.LifecycleStatus
	if status == dmarkets.MarketLifecyclePublished {
		status = dmarkets.MarketStatusActive
	}
	return marketGroupReviewResponse{
		ID:                 group.ID,
		QuestionTitle:      group.QuestionTitle,
		Description:        group.Description,
		GroupType:          group.GroupType,
		ProbabilityPolicy:  group.ProbabilityPolicy,
		ResolutionPolicy:   group.ResolutionPolicy,
		LifecycleStatus:    group.LifecycleStatus,
		Status:             status,
		ProposalCost:       group.ProposalCost,
		CreatorUsername:    group.CreatorUsername,
		StewardUsername:    group.StewardUsername,
		ApprovedBy:         group.ApprovedBy,
		ApprovedAt:         group.ApprovedAt,
		RejectedBy:         group.RejectedBy,
		RejectedAt:         group.RejectedAt,
		RejectionReason:    group.RejectionReason,
		ResolutionDateTime: group.ResolutionDateTime,
		CreatedAt:          group.CreatedAt,
		UpdatedAt:          group.UpdatedAt,
		AnswerCount:        len(group.Members),
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
