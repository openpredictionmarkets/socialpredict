package marketshandlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"socialpredict/handlers/markets/dto"
	dmarkets "socialpredict/internal/domain/markets"
	"socialpredict/models"
	"socialpredict/util"

	"github.com/gorilla/mux"
)

// getCreatorInfoForDetails fetches creator details from database for market details
func getCreatorInfoForDetails(username string) *dto.CreatorResponse {
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

// MarketDetailsHandler handles requests for detailed market information
func MarketDetailsHandler(svc dmarkets.ServiceInterface) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 1. Parse HTTP parameters
		vars := mux.Vars(r)
		marketIdStr := vars["marketId"]

		marketId, err := strconv.ParseInt(marketIdStr, 10, 64)
		if err != nil {
			http.Error(w, "Invalid market ID", http.StatusBadRequest)
			return
		}

		// 2. Call domain service to get market details
		details, err := svc.GetMarketDetails(r.Context(), marketId)
		if err != nil {
			// 3. Map domain errors to HTTP status codes
			switch err {
			case dmarkets.ErrMarketNotFound:
				http.Error(w, "Market not found", http.StatusNotFound)
			case dmarkets.ErrInvalidInput:
				http.Error(w, "Invalid market ID", http.StatusBadRequest)
			default:
				http.Error(w, "Internal server error", http.StatusInternalServerError)
			}
			return
		}

		// 4. Get actual creator info from database
		creatorUsername := details.Market.CreatorUsername
		creator := getCreatorInfoForDetails(creatorUsername)

		// 5. Convert domain model to response DTO
		response := dto.MarketDetailsResponse{
			MarketID:           marketId,
			Market:             details.Market,
			Creator:            creator, // Proper creator object instead of nil/interface{}
			ProbabilityChanges: details.ProbabilityChanges,
			NumUsers:           details.NumUsers,
			TotalVolume:        details.TotalVolume,
			MarketDust:         details.MarketDust,
		}

		// 5. Return response
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}
}
