package marketshandlers

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"socialpredict/middleware"
	"socialpredict/security"
	"socialpredict/util"

	"socialpredict/handlers/markets/dto"
	dmarkets "socialpredict/internal/domain/markets"
)

type CreateMarketHandler struct {
	svc dmarkets.Service
}

func NewCreateMarketHandler(svc dmarkets.Service) *CreateMarketHandler {
	return &CreateMarketHandler{svc: svc}
}

func (h *CreateMarketHandler) Handle(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method is not supported.", http.StatusMethodNotAllowed)
		return
	}

	// Validate user and get username
	db := util.GetDB()
	user, httperr := middleware.ValidateUserAndEnforcePasswordChangeGetUser(r, db)
	if httperr != nil {
		http.Error(w, httperr.Error(), httperr.StatusCode)
		return
	}

	// Parse request body
	var req dto.CreateMarketRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("Error reading request body: %v", err)
		http.Error(w, "Error reading request body", http.StatusBadRequest)
		return
	}

	// Optional security validation (keep existing behavior)
	securityService := security.NewSecurityService()
	marketInput := security.MarketInput{
		Title:       req.QuestionTitle,
		Description: req.Description,
		EndTime:     req.ResolutionDateTime.String(),
	}

	sanitizedInput, err := securityService.ValidateAndSanitizeMarketInput(marketInput)
	if err != nil {
		http.Error(w, "Invalid market data: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Update with sanitized data
	req.QuestionTitle = sanitizedInput.Title
	req.Description = sanitizedInput.Description

	// Convert DTO to domain request
	domainReq := dmarkets.MarketCreateRequest{
		QuestionTitle:      req.QuestionTitle,
		Description:        req.Description,
		OutcomeType:        req.OutcomeType,
		ResolutionDateTime: req.ResolutionDateTime,
		YesLabel:           req.YesLabel,
		NoLabel:            req.NoLabel,
	}

	// Call domain service
	market, err := h.svc.CreateMarket(context.Background(), domainReq, user.Username)
	if err != nil {
		// Map domain errors to HTTP status codes
		switch err {
		case dmarkets.ErrUserNotFound:
			http.Error(w, "User not found", http.StatusNotFound)
		case dmarkets.ErrInsufficientBalance:
			http.Error(w, "Insufficient balance", http.StatusBadRequest)
		case dmarkets.ErrInvalidQuestionLength,
			dmarkets.ErrInvalidDescriptionLength,
			dmarkets.ErrInvalidLabel,
			dmarkets.ErrInvalidResolutionTime:
			http.Error(w, err.Error(), http.StatusBadRequest)
		default:
			log.Printf("Error creating market: %v", err)
			http.Error(w, "Error creating market", http.StatusInternalServerError)
		}
		return
	}

	// Convert domain model to response DTO
	response := dto.CreateMarketResponse{
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
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}
