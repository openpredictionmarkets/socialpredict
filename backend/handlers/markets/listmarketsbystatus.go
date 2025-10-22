package marketshandlers

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

	"socialpredict/handlers/markets/dto"
	dmarkets "socialpredict/internal/domain/markets"
	"socialpredict/models"
	"socialpredict/util"

	"gorm.io/gorm"
)

// getCreatorInfo fetches creator details from database
func getCreatorInfo(username string) *dto.CreatorResponse {
	var user models.User
	db := util.GetDB()

	// Query user by username
	if err := db.Where("username = ?", username).First(&user).Error; err != nil {
		// If user not found, return default creator info
		log.Printf("Creator user not found for username %s: %v", username, err)
		return &dto.CreatorResponse{
			Username:      username,
			PersonalEmoji: "👤", // Default emoji
			DisplayName:   username,
		}
	}

	// Return actual user data
	return &dto.CreatorResponse{
		Username:      user.Username,
		PersonalEmoji: user.PersonalEmoji,
		DisplayName:   user.DisplayName,
	}
}

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
		markets, err := svc.ListByStatus(context.Background(), statusName, page)
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

			// Get actual creator info from database
			creator := getCreatorInfo(market.CreatorUsername)

			// Create market overview with basic data
			// TODO: Complex calculations (bets, probabilities, volumes) should be moved to domain service
			marketOverview := dto.MarketOverview{
				Market:          marketResponse,
				Creator:         creator, // Proper creator object instead of nil
				LastProbability: 0.5,     // TODO: Calculate in domain service
				NumUsers:        0,       // TODO: Calculate in domain service
				TotalVolume:     0,       // TODO: Calculate in domain service
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

// COMPATIBILITY FUNCTIONS FOR LEGACY CODE (searchmarkets.go)
// These functions maintain backward compatibility for files not yet refactored
// They can be removed once all handlers are migrated to domain service pattern

// MarketFilterFunc defines the filtering logic for markets (legacy compatibility)
type MarketFilterFunc func(*gorm.DB) *gorm.DB

// ActiveMarketsFilter returns markets that are not resolved and have not yet reached their resolution date
func ActiveMarketsFilter(db *gorm.DB) *gorm.DB {
	now := time.Now()
	return db.Where("is_resolved = ? AND resolution_date_time > ?", false, now)
}

// ClosedMarketsFilter returns markets that are not resolved but have passed their resolution date
func ClosedMarketsFilter(db *gorm.DB) *gorm.DB {
	now := time.Now()
	return db.Where("is_resolved = ? AND resolution_date_time <= ?", false, now)
}

// ResolvedMarketsFilter returns markets that have been resolved
func ResolvedMarketsFilter(db *gorm.DB) *gorm.DB {
	return db.Where("is_resolved = ?", true)
}

// ListMarketsByStatus - backward compatibility function for tests
func ListMarketsByStatus(db *gorm.DB, filterFunc MarketFilterFunc) ([]dto.MarketOverview, error) {
	var markets []models.Market

	// Apply the filter and get markets from database
	if err := filterFunc(db).Find(&markets).Error; err != nil {
		return nil, err
	}

	// Convert to market overviews (simplified for testing)
	var marketOverviews []dto.MarketOverview
	for _, market := range markets {
		// Create a basic market response
		marketResponse := dto.MarketResponse{
			ID:                 market.ID,
			QuestionTitle:      market.QuestionTitle,
			Description:        market.Description,
			OutcomeType:        market.OutcomeType,
			ResolutionDateTime: market.ResolutionDateTime,
			CreatorUsername:    market.CreatorUsername,
			YesLabel:           market.YesLabel,
			NoLabel:            market.NoLabel,
			CreatedAt:          market.CreatedAt,
			UpdatedAt:          market.UpdatedAt,
		}

		// Create market overview with minimal data for testing
		marketOverview := dto.MarketOverview{
			Market:          marketResponse,
			Creator:         nil, // Simplified for testing
			LastProbability: 0.5, // Default value for testing
			NumUsers:        0,   // Default value for testing
			TotalVolume:     0,   // Default value for testing
		}
		marketOverviews = append(marketOverviews, marketOverview)
	}

	return marketOverviews, nil
}
