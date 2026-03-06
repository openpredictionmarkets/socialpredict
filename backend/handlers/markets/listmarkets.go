package marketshandlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"socialpredict/handlers/markets/dto"
	dmarkets "socialpredict/internal/domain/markets"
)

// ListMarketsHandlerFactory creates an HTTP handler for listing markets with service injection
func ListMarketsHandlerFactory(svc dmarkets.ServiceInterface) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Println("ListMarketsHandler: Request received")
		if r.Method != http.MethodGet {
			http.Error(w, "Method is not supported.", http.StatusMethodNotAllowed)
			return
		}

		params, parseErr := parseListMarketsParams(r)
		if parseErr != nil {
			http.Error(w, parseErr.Error(), http.StatusBadRequest)
			return
		}

		markets, err := fetchMarkets(r, svc, params)
		if err != nil {
			writeListMarketsError(w, err)
			return
		}

		overviews, err := buildMarketOverviewResponses(r.Context(), svc, markets)
		if err != nil {
			log.Printf("Error building market overviews: %v", err)
			http.Error(w, "Error fetching markets", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(dto.ListMarketsResponse{
			Markets: overviews,
			Total:   len(overviews),
		}); err != nil {
			log.Printf("Error encoding response: %v", err)
			http.Error(w, "Error encoding response", http.StatusInternalServerError)
		}
	}
}

type listMarketsParams struct {
	status  string
	limit   int
	offset  int
	filters dmarkets.ListFilters
	page    dmarkets.Page
}

func parseListMarketsParams(r *http.Request) (listMarketsParams, error) {
	status, statusErr := normalizeStatusParam(r.URL.Query().Get("status"))
	if statusErr != nil {
		return listMarketsParams{}, statusErr
	}

	limit := parseListLimit(r.URL.Query().Get("limit"))
	offset := parseListOffset(r.URL.Query().Get("offset"))

	return listMarketsParams{
		status: status,
		limit:  limit,
		offset: offset,
		filters: dmarkets.ListFilters{
			Status: status,
			Limit:  limit,
			Offset: offset,
		},
		page: dmarkets.Page{
			Limit:  limit,
			Offset: offset,
		},
	}, nil
}

func parseListLimit(rawLimit string) int {
	limit := 50
	if rawLimit == "" {
		return limit
	}

	if parsedLimit, err := strconv.Atoi(rawLimit); err == nil && parsedLimit > 0 && parsedLimit <= 100 {
		return parsedLimit
	}
	return limit
}

func parseListOffset(rawOffset string) int {
	if rawOffset == "" {
		return 0
	}
	if parsedOffset, err := strconv.Atoi(rawOffset); err == nil && parsedOffset >= 0 {
		return parsedOffset
	}
	return 0
}

func fetchMarkets(r *http.Request, svc dmarkets.ServiceInterface, params listMarketsParams) ([]*dmarkets.Market, error) {
	if params.status != "" {
		return svc.ListByStatus(r.Context(), params.status, params.page)
	}
	return svc.ListMarkets(r.Context(), params.filters)
}

func writeListMarketsError(w http.ResponseWriter, err error) {
	switch err {
	case dmarkets.ErrInvalidInput:
		http.Error(w, "Invalid input parameters", http.StatusBadRequest)
	case dmarkets.ErrUnauthorized:
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
	default:
		log.Printf("Error fetching markets: %v", err)
		http.Error(w, "Error fetching markets", http.StatusInternalServerError)
	}
}
