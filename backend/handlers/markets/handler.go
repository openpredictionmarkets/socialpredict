package marketshandlers

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"

	"socialpredict/handlers/markets/dto"
	dmarkets "socialpredict/internal/domain/markets"
	"socialpredict/middleware"
	"socialpredict/util"

	"github.com/gorilla/mux"
)

// Service defines the interface for the markets domain service
type Service interface {
	CreateMarket(ctx context.Context, req dmarkets.MarketCreateRequest, creatorUsername string) (*dmarkets.Market, error)
	SetCustomLabels(ctx context.Context, marketID int64, yesLabel, noLabel string) error
	GetMarket(ctx context.Context, id int64) (*dmarkets.Market, error)
	ListMarkets(ctx context.Context, filters dmarkets.ListFilters) ([]*dmarkets.Market, error)
	SearchMarkets(ctx context.Context, query string, filters dmarkets.SearchFilters) ([]*dmarkets.Market, error)
	ResolveMarket(ctx context.Context, marketID int64, resolution string, username string) error
	ListByStatus(ctx context.Context, status string, p dmarkets.Page) ([]*dmarkets.Market, error)
	GetMarketLeaderboard(ctx context.Context, marketID int64, p dmarkets.Page) ([]*dmarkets.LeaderboardRow, error)
	ProjectProbability(ctx context.Context, req dmarkets.ProbabilityProjectionRequest) (*dmarkets.ProbabilityProjection, error)
	GetMarketDetails(ctx context.Context, marketID int64) (*dmarkets.MarketOverview, error)
}

// Handler handles HTTP requests for markets
type Handler struct {
	service Service
}

// NewHandler creates a new markets handler
func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

// CreateMarket handles POST /markets
func (h *Handler) CreateMarket(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method is not supported.", http.StatusMethodNotAllowed)
		return
	}

	// Validate user authentication
	db := util.GetDB()
	user, httperr := middleware.ValidateUserAndEnforcePasswordChangeGetUser(r, db)
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
	response := h.marketToResponse(market)

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
	response := h.marketToResponse(market)

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
	params.Status = r.URL.Query().Get("status")
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
	markets, err := h.service.ListMarkets(r.Context(), filters)
	if err != nil {
		h.handleError(w, err)
		return
	}

	// Convert to response DTOs
	responses := make([]*dto.MarketResponse, len(markets))
	for i, market := range markets {
		responses[i] = h.marketToResponse(market)
	}

	response := dto.SimpleListMarketsResponse{
		Markets: responses,
		Total:   len(responses),
	}

	// Send response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// SearchMarkets handles GET /markets/search
func (h *Handler) SearchMarkets(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method is not supported.", http.StatusMethodNotAllowed)
		return
	}

	// Parse query parameters
	var params dto.SearchMarketsQueryParams
	params.Query = r.URL.Query().Get("q")
	params.Status = r.URL.Query().Get("status")

	if params.Query == "" {
		http.Error(w, "Query parameter 'q' is required", http.StatusBadRequest)
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
	markets, err := h.service.SearchMarkets(r.Context(), params.Query, filters)
	if err != nil {
		h.handleError(w, err)
		return
	}

	// Convert to response DTOs
	responses := make([]*dto.MarketResponse, len(markets))
	for i, market := range markets {
		responses[i] = h.marketToResponse(market)
	}

	response := dto.SimpleListMarketsResponse{
		Markets: responses,
		Total:   len(responses),
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

	// Get user for authorization
	db := util.GetDB()
	user, httperr := middleware.ValidateUserAndEnforcePasswordChangeGetUser(r, db)
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

// marketToResponse converts a domain market to a response DTO
func (h *Handler) marketToResponse(market *dmarkets.Market) *dto.MarketResponse {
	return &dto.MarketResponse{
		ID:                 market.ID,
		QuestionTitle:      market.QuestionTitle,
		Description:        market.Description,
		OutcomeType:        market.OutcomeType,
		ResolutionDateTime: market.ResolutionDateTime,
		CreatorUsername:    market.CreatorUsername,
		YesLabel:           market.YesLabel,
		NoLabel:            market.NoLabel,
		Status:             market.Status,
		CreatedAt:          market.CreatedAt,
		UpdatedAt:          market.UpdatedAt,
	}
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
