package marketshandlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	authsvc "socialpredict/internal/service/auth"
	"socialpredict/security"
	"socialpredict/setup"
	"time"

	"socialpredict/handlers/markets/dto"
	dmarkets "socialpredict/internal/domain/markets"
)

// Constants for backward compatibility with tests
const (
	maxQuestionTitleLength = 160
)

// Helper functions for backward compatibility with tests
func checkQuestionTitleLength(title string) error {
	if len(title) > maxQuestionTitleLength || len(title) < 1 {
		return fmt.Errorf("question title exceeds %d characters or is blank", maxQuestionTitleLength)
	}
	return nil
}

func checkQuestionDescriptionLength(description string) error {
	if len(description) > 2000 {
		return errors.New("question description exceeds 2000 characters")
	}
	return nil
}

// ValidateMarketResolutionTime - test helper function for backward compatibility
func ValidateMarketResolutionTime(resolutionTime time.Time, config *setup.EconomicConfig) error {
	now := time.Now()
	minimumDuration := time.Duration(config.Economics.MarketCreation.MinimumFutureHours * float64(time.Hour))
	minimumFutureTime := now.Add(minimumDuration)

	if resolutionTime.Before(minimumFutureTime) || resolutionTime.Equal(minimumFutureTime) {
		return fmt.Errorf("market resolution time must be at least %.1f hours in the future",
			config.Economics.MarketCreation.MinimumFutureHours)
	}
	return nil
}

type CreateMarketService struct {
	svc  dmarkets.Service
	auth authsvc.Authenticator
}

func NewCreateMarketService(svc dmarkets.Service, auth authsvc.Authenticator) *CreateMarketService {
	return &CreateMarketService{
		svc:  svc,
		auth: auth,
	}
}

func (h *CreateMarketService) Handle(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method is not supported.", http.StatusMethodNotAllowed)
		return
	}

	// Validate user and get username
	if h.auth == nil {
		http.Error(w, "authentication service unavailable", http.StatusInternalServerError)
		return
	}

	user, httperr := h.auth.CurrentUser(r)
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

// CreateMarketHandlerWithService creates a handler with service injection
func CreateMarketHandlerWithService(svc dmarkets.ServiceInterface, auth authsvc.Authenticator, econConfig *setup.EconomicConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method is not supported.", http.StatusMethodNotAllowed)
			return
		}

		// Validate user and get username
		if auth == nil {
			http.Error(w, "authentication service unavailable", http.StatusInternalServerError)
			return
		}
		user, httperr := auth.CurrentUser(r)
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
		market, err := svc.CreateMarket(r.Context(), domainReq, user.Username)
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
}

// Legacy bridge function for backward compatibility with server routing
func CreateMarketHandler(loadEconConfig setup.EconConfigLoader) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		// TODO: This is a temporary bridge - should be replaced with proper DI container
		// For now, just return an error indicating this needs proper wiring
		http.Error(w, "Market creation temporarily disabled - handler needs proper dependency injection wiring", http.StatusServiceUnavailable)
	}
}
