package marketshandlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	dusers "socialpredict/internal/domain/users"
	authsvc "socialpredict/internal/service/auth"
	"socialpredict/security"
	"socialpredict/setup"
	"time"

	"socialpredict/handlers"
	"socialpredict/handlers/markets/dto"
	dmarkets "socialpredict/internal/domain/markets"
	"socialpredict/logger"
)

// Constants for backward compatibility with tests
const (
	maxQuestionTitleLength = 160
)

var errSecurityServiceUnavailable = errors.New("security service unavailable")

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
	svc             dmarkets.Service
	auth            authsvc.Authenticator
	securityService *security.SecurityService
}

func NewCreateMarketService(svc dmarkets.Service, auth authsvc.Authenticator, securityService *security.SecurityService) *CreateMarketService {
	return &CreateMarketService{
		svc:             svc,
		auth:            auth,
		securityService: securityService,
	}
}

func (h *CreateMarketService) currentUser(r *http.Request) (*dusers.User, *authsvc.AuthError) {
	if h.auth == nil {
		return nil, &authsvc.AuthError{Kind: authsvc.ErrorKindServiceUnavailable, Message: "authentication service unavailable"}
	}
	return h.auth.CurrentUser(r)
}

func (h *CreateMarketService) Handle(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeMethodNotAllowed(w)
		return
	}

	user, httpErr := h.currentUser(r)
	if httpErr != nil {
		logger.LogWarn("CreateMarket", "CurrentUser", httpErr.Message)
		writeAuthError(w, httpErr)
		return
	}

	req, decodeErr := decodeCreateMarketRequest(r)
	if decodeErr != nil {
		logger.LogWarn("CreateMarket", "DecodeRequest", decodeErr.Error())
		writeInvalidRequest(w)
		return
	}

	sanitized, sanitizeErr := sanitizeMarketRequest(h.securityService, req)
	if sanitizeErr != nil {
		logger.LogWarn("CreateMarket", "SanitizeMarketRequest", sanitizeErr.Error())
		writeCreateMarketSanitizationError(w, sanitizeErr)
		return
	}

	domainReq := toDomainCreateRequest(sanitized)

	market, err := h.svc.CreateMarket(context.Background(), domainReq, user.Username)
	if err != nil {
		writeCreateMarketError(w, err)
		logCreateMarketFailure(user.Username, err)
		return
	}

	response := toCreateMarketResponse(market)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
	logger.LogInfo("CreateMarket", "CreateMarket", fmt.Sprintf("Created market %d by user %s", market.ID, user.Username))
}

// CreateMarketHandlerWithService creates a handler with service injection
func CreateMarketHandlerWithService(svc dmarkets.ServiceInterface, auth authsvc.Authenticator, econConfig *setup.EconomicConfig, securityService *security.SecurityService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			writeMethodNotAllowed(w)
			return
		}

		user, httpErr := currentUserOrError(w, r, auth)
		if httpErr != nil {
			return
		}

		req, decodeErr := decodeCreateMarketRequest(r)
		if decodeErr != nil {
			logger.LogWarn("CreateMarket", "DecodeRequest", decodeErr.Error())
			writeInvalidRequest(w)
			return
		}

		sanitized, sanitizeErr := sanitizeMarketRequest(securityService, req)
		if sanitizeErr != nil {
			logger.LogWarn("CreateMarket", "SanitizeMarketRequest", sanitizeErr.Error())
			writeCreateMarketSanitizationError(w, sanitizeErr)
			return
		}

		domainReq := toDomainCreateRequest(sanitized)

		market, err := svc.CreateMarket(r.Context(), domainReq, user.Username)
		if err != nil {
			writeCreateMarketError(w, err)
			logCreateMarketFailure(user.Username, err)
			return
		}

		response := toCreateMarketResponse(market)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(response)
		logger.LogInfo("CreateMarket", "CreateMarket", fmt.Sprintf("Created market %d by user %s", market.ID, user.Username))
	}
}

func decodeCreateMarketRequest(r *http.Request) (dto.CreateMarketRequest, error) {
	var req dto.CreateMarketRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return dto.CreateMarketRequest{}, fmt.Errorf("Error reading request body")
	}
	return req, nil
}

func sanitizeMarketRequest(securityService *security.SecurityService, req dto.CreateMarketRequest) (dto.CreateMarketRequest, error) {
	if securityService == nil {
		return dto.CreateMarketRequest{}, errSecurityServiceUnavailable
	}
	marketInput := security.MarketInput{
		Title:       req.QuestionTitle,
		Description: req.Description,
		EndTime:     req.ResolutionDateTime.String(),
	}

	sanitizedInput, err := securityService.ValidateAndSanitizeMarketInput(marketInput)
	if err != nil {
		return dto.CreateMarketRequest{}, err
	}

	req.QuestionTitle = sanitizedInput.Title
	req.Description = sanitizedInput.Description
	return req, nil
}

func writeCreateMarketSanitizationError(w http.ResponseWriter, err error) {
	if errors.Is(err, errSecurityServiceUnavailable) {
		writeInternalError(w)
		return
	}
	_ = handlers.WriteFailure(w, http.StatusBadRequest, handlers.ReasonInvalidRequest)
}

func toDomainCreateRequest(req dto.CreateMarketRequest) dmarkets.MarketCreateRequest {
	return dmarkets.MarketCreateRequest{
		QuestionTitle:      req.QuestionTitle,
		Description:        req.Description,
		OutcomeType:        req.OutcomeType,
		ResolutionDateTime: req.ResolutionDateTime,
		YesLabel:           req.YesLabel,
		NoLabel:            req.NoLabel,
	}
}

func toCreateMarketResponse(market *dmarkets.Market) dto.CreateMarketResponse {
	return dto.CreateMarketResponse{
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
}

func currentUserOrError(w http.ResponseWriter, r *http.Request, auth authsvc.Authenticator) (*dusers.User, *authsvc.AuthError) {
	if auth == nil {
		logger.LogError("CreateMarket", "CurrentUser", errors.New("authentication service unavailable"))
		writeInternalError(w)
		return nil, &authsvc.AuthError{Kind: authsvc.ErrorKindServiceUnavailable, Message: "authentication service unavailable"}
	}
	user, httperr := auth.CurrentUser(r)
	if httperr != nil {
		logger.LogWarn("CreateMarket", "CurrentUser", httperr.Message)
		writeAuthError(w, httperr)
		return nil, httperr
	}
	return user, nil
}

func writeCreateMarketError(w http.ResponseWriter, err error) {
	writeCreateError(w, err)
}

func logCreateMarketFailure(username string, err error) {
	message := fmt.Sprintf("Create market failed for user %s: %v", username, err)

	switch err {
	case dmarkets.ErrUserNotFound,
		dmarkets.ErrInsufficientBalance,
		dmarkets.ErrInvalidQuestionLength,
		dmarkets.ErrInvalidDescriptionLength,
		dmarkets.ErrInvalidLabel,
		dmarkets.ErrInvalidResolutionTime:
		logger.LogWarn("CreateMarket", "CreateMarket", message)
	default:
		logger.LogError("CreateMarket", "CreateMarket", err)
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
