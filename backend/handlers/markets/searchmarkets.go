package marketshandlers

import (
	"encoding/json"
	"log"
	"net/http"
	"socialpredict/handlers/marketpublicresponse"
	marketmath "socialpredict/handlers/math/market"
	"socialpredict/handlers/math/probabilities/wpam"
	"socialpredict/handlers/tradingdata"
	"socialpredict/handlers/users/publicuser"
	"socialpredict/models"
	"socialpredict/security"
	"socialpredict/util"
	"strconv"
	"strings"

	"gorm.io/gorm"
)

// MarketOverview represents backward compatibility type for market overview data
type MarketOverview struct {
	Market          marketpublicresponse.PublicResponseMarket `json:"market"`
	Creator         interface{}                               `json:"creator"`
	LastProbability float64                                   `json:"lastProbability"`
	NumUsers        int                                       `json:"numUsers"`
	TotalVolume     int64                                     `json:"totalVolume"`
}

// SearchMarketsResponse defines the structure for search results
type SearchMarketsResponse struct {
	PrimaryResults  []MarketOverview `json:"primaryResults"`
	FallbackResults []MarketOverview `json:"fallbackResults"`
	Query           string           `json:"query"`
	PrimaryStatus   string           `json:"primaryStatus"`
	PrimaryCount    int              `json:"primaryCount"`
	FallbackCount   int              `json:"fallbackCount"`
	TotalCount      int              `json:"totalCount"`
	FallbackUsed    bool             `json:"fallbackUsed"`
}

// SearchMarketsHandler handles HTTP requests for searching markets
func SearchMarketsHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("SearchMarketsHandler: Request received")
	if r.Method != http.MethodGet {
		http.Error(w, "Method is not supported.", http.StatusMethodNotAllowed)
		return
	}

	db := util.GetDB()

	// Get and validate query parameters
	query := r.URL.Query().Get("query")
	status := r.URL.Query().Get("status")
	limitStr := r.URL.Query().Get("limit")

	// Validate and sanitize input
	if query == "" {
		http.Error(w, "Query parameter is required", http.StatusBadRequest)
		return
	}

	// Sanitize the search query
	sanitizer := security.NewSanitizer()
	sanitizedQuery, err := sanitizer.SanitizeMarketTitle(query)
	if err != nil {
		log.Printf("SearchMarketsHandler: Sanitization failed for query '%s': %v", query, err)
		http.Error(w, "Invalid search query: "+err.Error(), http.StatusBadRequest)
		return
	}
	if len(sanitizedQuery) > 100 {
		http.Error(w, "Query too long (max 100 characters)", http.StatusBadRequest)
		return
	}

	log.Printf("SearchMarketsHandler: Original query: '%s', Sanitized query: '%s'", query, sanitizedQuery)

	// Default values
	if status == "" {
		status = "all"
	}
	limit := 20
	if limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 && parsedLimit <= 50 {
			limit = parsedLimit
		}
	}

	// Perform the search
	searchResponse, err := SearchMarkets(db, sanitizedQuery, status, limit)
	if err != nil {
		log.Printf("Error searching markets: %v", err)
		http.Error(w, "Error searching markets", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(searchResponse); err != nil {
		log.Printf("Error encoding search response: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// SearchMarkets performs the actual search logic with fallback
func SearchMarkets(db *gorm.DB, query, status string, limit int) (*SearchMarketsResponse, error) {
	log.Printf("SearchMarkets: Searching for '%s' in status '%s'", query, status)

	// Get the appropriate filter function for the primary search
	var primaryFilter MarketFilterFunc
	var statusName string

	switch status {
	case "active":
		primaryFilter = ActiveMarketsFilter
		statusName = "active"
	case "closed":
		primaryFilter = ClosedMarketsFilter
		statusName = "closed"
	case "resolved":
		primaryFilter = ResolvedMarketsFilter
		statusName = "resolved"
	default:
		primaryFilter = func(db *gorm.DB) *gorm.DB {
			return db // No status filter for "all"
		}
		statusName = "all"
	}

	// Search within the primary status
	primaryResults, err := searchMarketsWithFilter(db, query, primaryFilter, limit)
	if err != nil {
		return nil, err
	}

	primaryOverviews, err := convertToMarketOverviews(db, primaryResults)
	if err != nil {
		return nil, err
	}

	response := &SearchMarketsResponse{
		PrimaryResults:  primaryOverviews,
		FallbackResults: []MarketOverview{},
		Query:           query,
		PrimaryStatus:   statusName,
		PrimaryCount:    len(primaryOverviews),
		FallbackCount:   0,
		TotalCount:      len(primaryOverviews),
		FallbackUsed:    false,
	}

	// If we have 5 or fewer primary results and we're not already searching "all", search all markets
	if len(primaryOverviews) <= 5 && status != "all" {
		log.Printf("SearchMarkets: Primary results ≤5, searching all markets for fallback")

		// Search all markets
		allFilter := func(db *gorm.DB) *gorm.DB {
			return db // No status filter
		}
		allResults, err := searchMarketsWithFilter(db, query, allFilter, limit*2) // Get more for filtering
		if err != nil {
			return nil, err
		}

		// Filter out markets that are already in primary results
		primaryIDs := make(map[int64]bool)
		for _, market := range primaryResults {
			primaryIDs[market.ID] = true
		}

		var fallbackResults []models.Market
		for _, market := range allResults {
			if !primaryIDs[market.ID] {
				fallbackResults = append(fallbackResults, market)
				if len(fallbackResults) >= limit {
					break
				}
			}
		}

		if len(fallbackResults) > 0 {
			fallbackOverviews, err := convertToMarketOverviews(db, fallbackResults)
			if err != nil {
				return nil, err
			}

			response.FallbackResults = fallbackOverviews
			response.FallbackCount = len(fallbackOverviews)
			response.TotalCount = response.PrimaryCount + response.FallbackCount
			response.FallbackUsed = true
		}
	}

	return response, nil
}

// searchMarketsWithFilter performs the database search with the given filter
func searchMarketsWithFilter(db *gorm.DB, searchQuery string, filterFunc MarketFilterFunc, limit int) ([]models.Market, error) {
	var markets []models.Market

	// Create the search query - search in both title and description
	searchTerm := "%" + strings.ToLower(searchQuery) + "%"
	log.Printf("searchMarketsWithFilter: searchTerm = '%s'", searchTerm)

	// Build the query with filter
	query := filterFunc(db).Where("LOWER(question_title) LIKE ? OR LOWER(description) LIKE ?", searchTerm, searchTerm).
		Order("created_at DESC").
		Limit(limit)

	// Log the SQL query for debugging
	log.Printf("searchMarketsWithFilter: Executing search query...")

	// Search in both question_title and description fields
	result := query.Find(&markets)

	if result.Error != nil {
		log.Printf("Error in searchMarketsWithFilter: %v", result.Error)
		return nil, result.Error
	}

	log.Printf("searchMarketsWithFilter: Found %d markets", len(markets))
	for i, market := range markets {
		log.Printf("  Market %d: ID=%d, Title='%s'", i+1, market.ID, market.QuestionTitle)
	}

	return markets, nil
}

// convertToMarketOverviews converts market models to MarketOverview structs
func convertToMarketOverviews(db *gorm.DB, markets []models.Market) ([]MarketOverview, error) {
	var marketOverviews []MarketOverview

	for _, market := range markets {
		// Get market data similar to listmarketsbystatus.go
		bets := tradingdata.GetBetsForMarket(db, uint(market.ID))
		probabilityChanges := wpam.CalculateMarketProbabilitiesWPAM(market.CreatedAt, bets)
		numUsers := models.GetNumMarketUsers(bets)
		marketVolume := marketmath.GetMarketVolume(bets)
		lastProbability := probabilityChanges[len(probabilityChanges)-1].Probability

		creatorInfo := publicuser.GetPublicUserInfo(db, market.CreatorUsername)

		// Get public response market
		marketIDStr := strconv.FormatUint(uint64(market.ID), 10)
		publicResponseMarket, err := marketpublicresponse.GetPublicResponseMarketByID(db, marketIDStr)
		if err != nil {
			log.Printf("Error getting public response market for ID %s: %v", marketIDStr, err)
			continue // Skip this market instead of failing the entire request
		}

		marketOverview := MarketOverview{
			Market:          publicResponseMarket,
			Creator:         creatorInfo,
			LastProbability: lastProbability,
			NumUsers:        numUsers,
			TotalVolume:     marketVolume,
		}
		marketOverviews = append(marketOverviews, marketOverview)
	}

	return marketOverviews, nil
}
