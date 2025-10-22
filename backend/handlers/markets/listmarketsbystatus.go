package marketshandlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"socialpredict/handlers/markets/dto"
	dmarkets "socialpredict/internal/domain/markets"
)

// ListMarketsStatusResponse defines the structure for filtered market responses
type ListMarketsStatusResponse struct {
	Markets []dto.MarketOverview `json:"markets"`
	Status  string               `json:"status"`
	Count   int                  `json:"count"`
}

// ListMarketsByStatusHandler creates a handler for listing markets by status using domain service
func ListMarketsByStatusHandler(svc dmarkets.ServiceInterface, statusName string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("ListMarketsByStatusHandler: Request received for status: %s", statusName)
		if r.Method != http.MethodGet {
			http.Error(w, "Method is not supported.", http.StatusMethodNotAllowed)
			return
		}

		// Parse query parameters for pagination
		limitStr := r.URL.Query().Get("limit")
		offsetStr := r.URL.Query().Get("offset")

		// Parse limit with default
		limit := 100
		if limitStr != "" {
			if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 {
				limit = parsedLimit
			}
		}

		// Parse offset with default
		offset := 0
		if offsetStr != "" {
			if parsedOffset, err := strconv.Atoi(offsetStr); err == nil && parsedOffset >= 0 {
				offset = parsedOffset
			}
		}

		// Build domain pagination
		page := dmarkets.Page{
			Limit:  limit,
			Offset: offset,
		}

		// Call domain service
		markets, err := svc.ListByStatus(r.Context(), statusName, page)
		if err != nil {
			// Map domain errors to HTTP status codes
			switch err {
			case dmarkets.ErrInvalidInput:
				http.Error(w, "Invalid status parameter", http.StatusBadRequest)
			case dmarkets.ErrMarketNotFound:
				http.Error(w, "No markets found", http.StatusNotFound)
			default:
				log.Printf("Error fetching markets for status %s: %v", statusName, err)
				http.Error(w, "Error fetching markets", http.StatusInternalServerError)
			}
			return
		}

		// Convert domain models to DTOs
		var marketOverviews []dto.MarketOverview
		for _, market := range markets {
			// Convert domain market to DTO market response
			marketResponse := dto.MarketResponse{
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

			// Create basic creator info - domain service should provide this
			creator := &dto.CreatorResponse{
				Username:      market.CreatorUsername,
				PersonalEmoji: "👤", // Default emoji - TODO: get from domain service
				DisplayName:   market.CreatorUsername,
			}

			// Create market overview with basic data
			// TODO: Complex calculations (bets, probabilities, volumes) should be moved to domain service
			marketOverview := dto.MarketOverview{
				Market:          marketResponse,
				Creator:         creator,
				LastProbability: 0.5, // TODO: Calculate in domain service
				NumUsers:        0,   // TODO: Calculate in domain service
				TotalVolume:     0,   // TODO: Calculate in domain service
			}
			marketOverviews = append(marketOverviews, marketOverview)
		}

		// Ensure empty array instead of null
		if marketOverviews == nil {
			marketOverviews = make([]dto.MarketOverview, 0)
		}

		// Build response
		response := ListMarketsStatusResponse{
			Markets: marketOverviews,
			Status:  statusName,
			Count:   len(marketOverviews),
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(response); err != nil {
			log.Printf("Error encoding response for status %s: %v", statusName, err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}

// ListActiveMarketsHandler handles HTTP requests for active markets
func ListActiveMarketsHandler(svc dmarkets.ServiceInterface) http.HandlerFunc {
	return ListMarketsByStatusHandler(svc, "active")
}

// ListClosedMarketsHandler handles HTTP requests for closed markets
func ListClosedMarketsHandler(svc dmarkets.ServiceInterface) http.HandlerFunc {
	return ListMarketsByStatusHandler(svc, "closed")
}

// ListResolvedMarketsHandler handles HTTP requests for resolved markets
func ListResolvedMarketsHandler(svc dmarkets.ServiceInterface) http.HandlerFunc {
	return ListMarketsByStatusHandler(svc, "resolved")
}
