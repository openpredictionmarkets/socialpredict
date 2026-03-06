package marketshandlers

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"

	"socialpredict/handlers/markets/dto"
	dmarkets "socialpredict/internal/domain/markets"
	"socialpredict/security"
)

// SearchMarketsHandler handles HTTP requests for searching markets - HTTP-only with service injection
func SearchMarketsHandler(svc dmarkets.ServiceInterface) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("SearchMarketsHandler: Request received")
		if r.Method != http.MethodGet {
			http.Error(w, "Method is not supported.", http.StatusMethodNotAllowed)
			return
		}

		params, clientErr := parseSearchRequest(r)
		if clientErr != nil {
			http.Error(w, clientErr.message, clientErr.statusCode)
			return
		}

		// ms, err := h.service.SearchMarkets(r.Context(), q, f)
		searchResults, err := svc.SearchMarkets(r.Context(), params.query, params.filters)
		if err != nil {
			// Map errors
			switch err {
			case dmarkets.ErrInvalidInput:
				http.Error(w, "Invalid search parameters", http.StatusBadRequest)
			default:
				log.Printf("Error searching markets: %v", err)
				http.Error(w, "Error searching markets", http.StatusInternalServerError)
			}
			return
		}

		response, buildErr := buildSearchResponse(r.Context(), svc, searchResults)
		if buildErr != nil {
			http.Error(w, buildErr.Error(), http.StatusInternalServerError)
			return
		}

		// Encode response
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(response); err != nil {
			log.Printf("Error encoding search response: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}

type searchRequestParams struct {
	query   string
	filters dmarkets.SearchFilters
}

type httpError struct {
	message    string
	statusCode int
}

func parseSearchRequest(r *http.Request) (searchRequestParams, *httpError) {
	query := extractQuery(r)
	status, statusErr := normalizeStatusParam(r.URL.Query().Get("status"))
	if statusErr != nil {
		return searchRequestParams{}, &httpError{message: statusErr.Error(), statusCode: http.StatusBadRequest}
	}
	if query == "" {
		return searchRequestParams{}, &httpError{message: "Query parameter 'query' is required", statusCode: http.StatusBadRequest}
	}

	sanitizedQuery, sanitizeErr := sanitizeQuery(query)
	if sanitizeErr != nil {
		return searchRequestParams{}, sanitizeErr
	}

	filters := dmarkets.SearchFilters{
		Status: status, // Can be empty, "active", "closed", "resolved", or "all"
		Limit:  parseLimit(r.URL.Query().Get("limit")),
		Offset: parseOffset(r.URL.Query().Get("offset")),
	}

	return searchRequestParams{
		query:   sanitizedQuery,
		filters: filters,
	}, nil
}

func extractQuery(r *http.Request) string {
	query := r.URL.Query().Get("query")
	if query == "" {
		query = r.URL.Query().Get("q")
	}
	return query
}

func sanitizeQuery(query string) (string, *httpError) {
	sanitizer := security.NewSanitizer()
	sanitizedQuery, err := sanitizer.SanitizeMarketTitle(query)
	if err != nil {
		log.Printf("SearchMarketsHandler: Sanitization failed for query '%s': %v", query, err)
		return "", &httpError{message: "Invalid search query: " + err.Error(), statusCode: http.StatusBadRequest}
	}
	if len(sanitizedQuery) > 100 {
		return "", &httpError{message: "Query too long (max 100 characters)", statusCode: http.StatusBadRequest}
	}
	return sanitizedQuery, nil
}

func parseLimit(rawLimit string) int {
	limit := 20 // Default
	if rawLimit == "" {
		return limit
	}

	parsedLimit, err := strconv.Atoi(rawLimit)
	if err != nil || parsedLimit <= 0 || parsedLimit > 50 {
		return limit
	}
	return parsedLimit
}

func parseOffset(rawOffset string) int {
	if rawOffset == "" {
		return 0
	}
	parsedOffset, err := strconv.Atoi(rawOffset)
	if err != nil || parsedOffset < 0 {
		return 0
	}
	return parsedOffset
}

func buildSearchResponse(ctx context.Context, svc dmarkets.ServiceInterface, searchResults *dmarkets.SearchResults) (dto.SearchResponse, error) {
	primaryOverviews, err := buildMarketOverviewResponses(ctx, svc, searchResults.PrimaryResults)
	if err != nil {
		log.Printf("Error building primary results: %v", err)
		return dto.SearchResponse{}, errors.New("Error building primary results")
	}

	fallbackOverviews, err := buildMarketOverviewResponses(ctx, svc, searchResults.FallbackResults)
	if err != nil {
		log.Printf("Error building fallback results: %v", err)
		return dto.SearchResponse{}, errors.New("Error building fallback results")
	}

	return dto.SearchResponse{
		PrimaryResults:  primaryOverviews,
		FallbackResults: fallbackOverviews,
		Query:           searchResults.Query,
		PrimaryStatus:   searchResults.PrimaryStatus,
		PrimaryCount:    searchResults.PrimaryCount,
		FallbackCount:   searchResults.FallbackCount,
		TotalCount:      searchResults.TotalCount,
		FallbackUsed:    searchResults.FallbackUsed,
	}, nil
}
