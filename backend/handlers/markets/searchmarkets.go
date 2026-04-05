package marketshandlers

import (
	"context"
	"errors"
	"log"
	"net/http"

	"socialpredict/handlers/markets/dto"
	dmarkets "socialpredict/internal/domain/markets"
	"socialpredict/security"
)

// SearchMarketsHandler handles HTTP requests for searching markets - HTTP-only with service injection
func SearchMarketsHandler(svc dmarkets.ServiceInterface) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		handleSearchMarkets(w, r, svc)
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
	query, queryErr := extractQuery(r)
	if queryErr != nil {
		return searchRequestParams{}, queryErr
	}

	sanitizedQuery, sanitizeErr := sanitizeQuery(query)
	if sanitizeErr != nil {
		return searchRequestParams{}, sanitizeErr
	}

	return buildSearchRequestParams(r, sanitizedQuery)
}

func extractQuery(r *http.Request) (string, *httpError) {
	for _, key := range []string{"query", "q"} {
		if value := r.URL.Query().Get(key); value != "" {
			return value, nil
		}
	}
	return "", &httpError{message: "Query parameter 'query' is required", statusCode: http.StatusBadRequest}
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
	return parseBoundedInt(rawLimit, 20, 1, 50)
}

func parseOffset(rawOffset string) int {
	return parseBoundedInt(rawOffset, 0, 0, int(^uint(0)>>1))
}

func buildSearchResponse(ctx context.Context, svc dmarkets.ServiceInterface, searchResults *dmarkets.SearchResults) (dto.SearchResponse, error) {
	primaryOverviews, err := buildSearchResultSet(ctx, svc, searchResults.PrimaryResults)
	if err != nil {
		log.Printf("Error building primary results: %v", err)
		return dto.SearchResponse{}, errors.New("Error building primary results")
	}

	fallbackOverviews, err := buildSearchResultSet(ctx, svc, searchResults.FallbackResults)
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

func buildSearchRequestParams(r *http.Request, query string) (searchRequestParams, *httpError) {
	filters, filtersErr := parseSearchFilters(r)
	if filtersErr != nil {
		return searchRequestParams{}, filtersErr
	}

	return searchRequestParams{
		query:   query,
		filters: filters,
	}, nil
}

func parseSearchFilters(r *http.Request) (dmarkets.SearchFilters, *httpError) {
	status, err := normalizeStatusParam(r.URL.Query().Get("status"))
	if err != nil {
		return dmarkets.SearchFilters{}, &httpError{message: err.Error(), statusCode: http.StatusBadRequest}
	}

	return dmarkets.SearchFilters{
		Status: status,
		Limit:  parseLimit(r.URL.Query().Get("limit")),
		Offset: parseOffset(r.URL.Query().Get("offset")),
	}, nil
}

func searchMarkets(ctx context.Context, svc dmarkets.ServiceInterface, params searchRequestParams) (*dmarkets.SearchResults, error) {
	return svc.SearchMarkets(ctx, params.query, params.filters)
}

func buildSearchResultSet(ctx context.Context, svc dmarkets.ServiceInterface, markets []*dmarkets.Market) ([]*dto.MarketOverviewResponse, error) {
	return buildMarketOverviewResponses(ctx, svc, markets)
}

func writeSearchResponse(w http.ResponseWriter, ctx context.Context, svc dmarkets.ServiceInterface, searchResults *dmarkets.SearchResults) error {
	response, err := buildSearchResponse(ctx, svc, searchResults)
	if err != nil {
		return err
	}

	if err := writeJSON(w, http.StatusOK, response); err != nil {
		log.Printf("Error encoding search response: %v", err)
		return err
	}

	return nil
}

func writeSearchError(w http.ResponseWriter, err error) {
	switch err {
	case dmarkets.ErrInvalidInput:
		http.Error(w, "Invalid search parameters", http.StatusBadRequest)
	default:
		log.Printf("Error searching markets: %v", err)
		http.Error(w, "Error searching markets", http.StatusInternalServerError)
	}
}

func handleSearchMarkets(w http.ResponseWriter, r *http.Request, svc dmarkets.ServiceInterface) {
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

	searchResults, err := searchMarkets(r.Context(), svc, params)
	if err != nil {
		writeSearchError(w, err)
		return
	}

	if err := writeSearchResponse(w, r.Context(), svc, searchResults); err != nil {
		log.Printf("Error writing search response: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
