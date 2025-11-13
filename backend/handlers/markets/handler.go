package marketshandlers

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"

	"socialpredict/handlers/markets/dto"
	dmarkets "socialpredict/internal/domain/markets"
	authsvc "socialpredict/internal/service/auth"

	"github.com/gorilla/mux"
)

// Service defines the interface for the markets domain service
type Service interface {
	CreateMarket(ctx context.Context, req dmarkets.MarketCreateRequest, creatorUsername string) (*dmarkets.Market, error)
	SetCustomLabels(ctx context.Context, marketID int64, yesLabel, noLabel string) error
	GetMarket(ctx context.Context, id int64) (*dmarkets.Market, error)
	ListMarkets(ctx context.Context, filters dmarkets.ListFilters) ([]*dmarkets.Market, error)
	GetMarketDetails(ctx context.Context, marketID int64) (*dmarkets.MarketOverview, error)
	SearchMarkets(ctx context.Context, query string, filters dmarkets.SearchFilters) (*dmarkets.SearchResults, error)
	ResolveMarket(ctx context.Context, marketID int64, resolution string, username string) error
	ListByStatus(ctx context.Context, status string, p dmarkets.Page) ([]*dmarkets.Market, error)
	GetMarketLeaderboard(ctx context.Context, marketID int64, p dmarkets.Page) ([]*dmarkets.LeaderboardRow, error)
	ProjectProbability(ctx context.Context, req dmarkets.ProbabilityProjectionRequest) (*dmarkets.ProbabilityProjection, error)
}

// Handler handles HTTP requests for markets
type Handler struct {
	service Service
	auth    authsvc.Authenticator
}

// NewHandler creates a new markets handler
func NewHandler(service Service, auth authsvc.Authenticator) *Handler {
	return &Handler{
		service: service,
		auth:    auth,
	}
}

// CreateMarket handles POST /markets
func (h *Handler) CreateMarket(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method is not supported.", http.StatusMethodNotAllowed)
		return
	}

	if h.auth == nil {
		http.Error(w, "authentication service unavailable", http.StatusInternalServerError)
		return
	}

	// Validate user authentication via auth service
	user, httperr := h.auth.CurrentUser(r)
	if httperr != nil {
		http.Error(w, httperr.Error(), httperr.StatusCode)
		return
	}

	// Parse request body
	var req dto.CreateMarketRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON in request body", http.StatusBadRequest)
		return
	}

	// Convert DTO to domain model
	createReq := dmarkets.MarketCreateRequest{
		QuestionTitle:      req.QuestionTitle,
		Description:        req.Description,
		OutcomeType:        req.OutcomeType,
		ResolutionDateTime: req.ResolutionDateTime,
		YesLabel:           req.YesLabel,
		NoLabel:            req.NoLabel,
	}

	// Call service
	market, err := h.service.CreateMarket(r.Context(), createReq, user.Username)
	if err != nil {
		h.handleError(w, err)
		return
	}

	// Convert to response DTO
	response := marketToResponse(market)

	// Send response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

// UpdateLabels handles PUT /markets/{id}/labels
func (h *Handler) UpdateLabels(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "Method is not supported.", http.StatusMethodNotAllowed)
		return
	}

	// Parse market ID from URL
	vars := mux.Vars(r)
	idStr := vars["id"]
	if idStr == "" {
		http.Error(w, "Market ID is required", http.StatusBadRequest)
		return
	}

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid market ID", http.StatusBadRequest)
		return
	}

	// Parse request body
	var req dto.UpdateLabelsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON in request body", http.StatusBadRequest)
		return
	}

	// Call service
	if err := h.service.SetCustomLabels(r.Context(), id, req.YesLabel, req.NoLabel); err != nil {
		h.handleError(w, err)
		return
	}

	// Send success response
	w.WriteHeader(http.StatusNoContent)
}

// GetMarket handles GET /markets/{id}
func (h *Handler) GetMarket(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method is not supported.", http.StatusMethodNotAllowed)
		return
	}

	// Parse market ID from URL
	vars := mux.Vars(r)
	idStr := vars["id"]
	if idStr == "" {
		http.Error(w, "Market ID is required", http.StatusBadRequest)
		return
	}

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid market ID", http.StatusBadRequest)
		return
	}

	// Call service
	market, err := h.service.GetMarket(r.Context(), id)
	if err != nil {
		h.handleError(w, err)
		return
	}

	// Convert to response DTO
	response := marketToResponse(market)

	// Send response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// ListMarkets handles GET /markets
func (h *Handler) ListMarkets(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method is not supported.", http.StatusMethodNotAllowed)
		return
	}

	// Parse query parameters
	var params dto.ListMarketsQueryParams
	status, err := normalizeStatusParam(r.URL.Query().Get("status"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	params.Status = status
	params.CreatedBy = r.URL.Query().Get("created_by")

	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil {
			params.Limit = limit
		}
	}

	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		if offset, err := strconv.Atoi(offsetStr); err == nil {
			params.Offset = offset
		}
	}

	// Convert to domain filters
	filters := dmarkets.ListFilters{
		Status:    params.Status,
		CreatedBy: params.CreatedBy,
		Limit:     params.Limit,
		Offset:    params.Offset,
	}

	// Call service
	var markets []*dmarkets.Market
	if params.Status != "" {
		page := dmarkets.Page{Limit: params.Limit, Offset: params.Offset}
		markets, err = h.service.ListByStatus(r.Context(), params.Status, page)
	} else {
		markets, err = h.service.ListMarkets(r.Context(), filters)
	}
	if err != nil {
		h.handleError(w, err)
		return
	}

	overviews, err := buildMarketOverviewResponses(r.Context(), h.service, markets)
	if err != nil {
		h.handleError(w, err)
		return
	}

	// Send response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(dto.ListMarketsResponse{
		Markets: overviews,
		Total:   len(overviews),
	})
}

// SearchMarkets handles GET /markets/search
func (h *Handler) SearchMarkets(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method is not supported.", http.StatusMethodNotAllowed)
		return
	}

	// Parse query parameters
	var params dto.SearchMarketsQueryParams
	params.Query = r.URL.Query().Get("query")
	if params.Query == "" {
		params.Query = r.URL.Query().Get("q")
	}
	status, err := normalizeStatusParam(r.URL.Query().Get("status"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	params.Status = status

	if params.Query == "" {
		http.Error(w, "Query parameter 'query' is required", http.StatusBadRequest)
		return
	}

	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil {
			params.Limit = limit
		}
	}

	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		if offset, err := strconv.Atoi(offsetStr); err == nil {
			params.Offset = offset
		}
	}

	// Convert to domain filters
	filters := dmarkets.SearchFilters{
		Status: params.Status,
		Limit:  params.Limit,
		Offset: params.Offset,
	}

	// Call service
	searchResults, err := h.service.SearchMarkets(r.Context(), params.Query, filters)
	if err != nil {
		h.handleError(w, err)
		return
	}

	primaryOverviews, err := buildMarketOverviewResponses(r.Context(), h.service, searchResults.PrimaryResults)
	if err != nil {
		h.handleError(w, err)
		return
	}

	fallbackOverviews, err := buildMarketOverviewResponses(r.Context(), h.service, searchResults.FallbackResults)
	if err != nil {
		h.handleError(w, err)
		return
	}

	response := dto.SearchResponse{
		PrimaryResults:  primaryOverviews,
		FallbackResults: fallbackOverviews,
		Query:           searchResults.Query,
		PrimaryStatus:   searchResults.PrimaryStatus,
		PrimaryCount:    searchResults.PrimaryCount,
		FallbackCount:   searchResults.FallbackCount,
		TotalCount:      searchResults.TotalCount,
		FallbackUsed:    searchResults.FallbackUsed,
	}

	// Send response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// ResolveMarket handles POST /markets/{id}/resolve
func (h *Handler) ResolveMarket(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method is not supported.", http.StatusMethodNotAllowed)
		return
	}

	// Parse market ID from URL
	vars := mux.Vars(r)
	idStr := vars["id"]
	if idStr == "" {
		http.Error(w, "Market ID is required", http.StatusBadRequest)
		return
	}

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid market ID", http.StatusBadRequest)
		return
	}

	if h.auth == nil {
		http.Error(w, "authentication service unavailable", http.StatusInternalServerError)
		return
	}

	// Get user for authorization
	user, httperr := h.auth.CurrentUser(r)
	if httperr != nil {
		http.Error(w, httperr.Error(), httperr.StatusCode)
		return
	}

	// Parse request body
	var req dto.ResolveMarketRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON in request body", http.StatusBadRequest)
		return
	}

	// Call service
	if err := h.service.ResolveMarket(r.Context(), id, req.Resolution, user.Username); err != nil {
		h.handleError(w, err)
		return
	}

	// Send success response
	w.WriteHeader(http.StatusNoContent)
}

// ListByStatus handles GET /markets/status/{status}
func (h *Handler) ListByStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method is not supported.", http.StatusMethodNotAllowed)
		return
	}

	// Parse status from URL
	vars := mux.Vars(r)
	statusValue, err := normalizeStatusParam(vars["status"])
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if vars["status"] == "" {
		http.Error(w, "Status is required", http.StatusBadRequest)
		return
	}

	// Parse pagination parameters
	limit := 100
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}

	offset := 0
	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		if parsedOffset, err := strconv.Atoi(offsetStr); err == nil && parsedOffset >= 0 {
			offset = parsedOffset
		}
	}

	page := dmarkets.Page{
		Limit:  limit,
		Offset: offset,
	}

	var markets []*dmarkets.Market
	if statusValue == "" {
		filters := dmarkets.ListFilters{
			Status: "",
			Limit:  limit,
			Offset: offset,
		}
		markets, err = h.service.ListMarkets(r.Context(), filters)
	} else {
		markets, err = h.service.ListByStatus(r.Context(), statusValue, page)
	}
	if err != nil {
		h.handleError(w, err)
		return
	}

	// Convert to response DTOs
	overviews, err := buildMarketOverviewResponses(r.Context(), h.service, markets)
	if err != nil {
		h.handleError(w, err)
		return
	}

	// Send response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(dto.ListMarketsResponse{
		Markets: overviews,
		Total:   len(overviews),
	})
}

// GetDetails handles GET /markets/{id} with full market details
func (h *Handler) GetDetails(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method is not supported.", http.StatusMethodNotAllowed)
		return
	}

	// Parse market ID from URL
	vars := mux.Vars(r)
	idStr := vars["id"]
	if idStr == "" {
		http.Error(w, "Market ID is required", http.StatusBadRequest)
		return
	}

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid market ID", http.StatusBadRequest)
		return
	}

	// Call service
	details, err := h.service.GetMarketDetails(r.Context(), id)
	if err != nil {
		h.handleError(w, err)
		return
	}

	response := dto.MarketDetailsResponse{
		Market:             publicMarketResponseFromDomain(details.Market),
		Creator:            creatorResponseFromSummary(details.Creator),
		ProbabilityChanges: probabilityChangesToResponse(details.ProbabilityChanges),
		NumUsers:           details.NumUsers,
		TotalVolume:        details.TotalVolume,
		MarketDust:         details.MarketDust,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// MarketLeaderboard handles GET /markets/{id}/leaderboard
func (h *Handler) MarketLeaderboard(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method is not supported.", http.StatusMethodNotAllowed)
		return
	}

	// Parse market ID from URL
	vars := mux.Vars(r)
	idStr := vars["id"]
	if idStr == "" {
		http.Error(w, "Market ID is required", http.StatusBadRequest)
		return
	}

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid market ID", http.StatusBadRequest)
		return
	}

	// Parse pagination parameters
	limit := 100
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}

	offset := 0
	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		if parsedOffset, err := strconv.Atoi(offsetStr); err == nil && parsedOffset >= 0 {
			offset = parsedOffset
		}
	}

	page := dmarkets.Page{
		Limit:  limit,
		Offset: offset,
	}

	// Call service
	leaderboard, err := h.service.GetMarketLeaderboard(r.Context(), id, page)
	if err != nil {
		h.handleError(w, err)
		return
	}

	// Convert to response DTOs
	var leaderRows []dto.LeaderboardRow
	for _, row := range leaderboard {
		leaderRows = append(leaderRows, dto.LeaderboardRow{
			Username:       row.Username,
			Profit:         row.Profit,
			CurrentValue:   row.CurrentValue,
			TotalSpent:     row.TotalSpent,
			Position:       row.Position,
			YesSharesOwned: row.YesSharesOwned,
			NoSharesOwned:  row.NoSharesOwned,
			Rank:           row.Rank,
		})
	}

	// Ensure empty array instead of null
	if leaderRows == nil {
		leaderRows = make([]dto.LeaderboardRow, 0)
	}

	response := dto.LeaderboardResponse{
		MarketID:    id,
		Leaderboard: leaderRows,
		Total:       len(leaderRows),
	}

	// Send response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// ProjectProbability handles GET /markets/{id}/projection
func (h *Handler) ProjectProbability(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method is not supported.", http.StatusMethodNotAllowed)
		return
	}

	// Parse market ID from URL
	vars := mux.Vars(r)
	marketIdStr := vars["marketId"]
	amountStr := vars["amount"]
	outcome := vars["outcome"]

	// Parse marketId
	marketId, err := strconv.ParseInt(marketIdStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid market ID", http.StatusBadRequest)
		return
	}

	// Parse amount
	amount, err := strconv.ParseInt(amountStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid amount value", http.StatusBadRequest)
		return
	}

	// Build domain request
	projectionReq := dmarkets.ProbabilityProjectionRequest{
		MarketID: marketId,
		Amount:   amount,
		Outcome:  outcome,
	}

	// Call service
	projection, err := h.service.ProjectProbability(r.Context(), projectionReq)
	if err != nil {
		h.handleError(w, err)
		return
	}

	// Return response DTO
	response := dto.ProbabilityProjectionResponse{
		MarketID:             marketId,
		CurrentProbability:   projection.CurrentProbability,
		ProjectedProbability: projection.ProjectedProbability,
		Amount:               amount,
		Outcome:              outcome,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleError maps domain errors to HTTP responses
func (h *Handler) handleError(w http.ResponseWriter, err error) {
	var statusCode int
	var message string

	switch err {
	case dmarkets.ErrMarketNotFound:
		statusCode = http.StatusNotFound
		message = "Market not found"
	case dmarkets.ErrInvalidQuestionLength, dmarkets.ErrInvalidDescriptionLength, dmarkets.ErrInvalidLabel, dmarkets.ErrInvalidResolutionTime:
		statusCode = http.StatusBadRequest
		message = err.Error()
	case dmarkets.ErrUserNotFound:
		statusCode = http.StatusNotFound
		message = "User not found"
	case dmarkets.ErrInsufficientBalance:
		statusCode = http.StatusBadRequest
		message = "Insufficient balance"
	case dmarkets.ErrUnauthorized:
		statusCode = http.StatusUnauthorized
		message = "Unauthorized"
	default:
		statusCode = http.StatusInternalServerError
		message = "Internal server error"
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	response := dto.ErrorResponse{
		Error: message,
	}
	json.NewEncoder(w).Encode(response)
}
