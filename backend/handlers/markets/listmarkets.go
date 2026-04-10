package marketshandlers

import (
	"log"
	"net/http"

	"socialpredict/handlers/markets/dto"
	dmarkets "socialpredict/internal/domain/markets"
)

// ListMarketsHandlerFactory creates an HTTP handler for listing markets with service injection
func ListMarketsHandlerFactory(svc dmarkets.ServiceInterface) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		handleListMarkets(w, r, svc)
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

	page := parseListPage(r)

	return listMarketsParams{
		status: status,
		limit:  page.Limit,
		offset: page.Offset,
		filters: dmarkets.ListFilters{
			Status: status,
			Limit:  page.Limit,
			Offset: page.Offset,
		},
		page: page,
	}, nil
}

func parseListLimit(rawLimit string) int {
	return parseBoundedInt(rawLimit, 50, 1, 100)
}

func parseListOffset(rawOffset string) int {
	return parseBoundedInt(rawOffset, 0, 0, int(^uint(0)>>1))
}

func fetchMarkets(r *http.Request, svc dmarkets.ServiceInterface, params listMarketsParams) ([]*dmarkets.Market, error) {
	if params.filters.Status != "" {
		return fetchMarketsByStatus(r, svc, params)
	}
	return fetchAllMarkets(r, svc, params)
}

func writeListMarketsError(w http.ResponseWriter, err error) {
	message, statusCode := listMarketsErrorResponse(err)
	if statusCode == http.StatusInternalServerError {
		log.Printf("Error fetching markets: %v", err)
	}
	http.Error(w, message, statusCode)
}

func parseListPage(r *http.Request) dmarkets.Page {
	return dmarkets.Page{
		Limit:  parseListLimit(r.URL.Query().Get("limit")),
		Offset: parseListOffset(r.URL.Query().Get("offset")),
	}
}

func fetchMarketsByStatus(r *http.Request, svc dmarkets.ServiceInterface, params listMarketsParams) ([]*dmarkets.Market, error) {
	return svc.ListByStatus(r.Context(), params.filters.Status, params.page)
}

func fetchAllMarkets(r *http.Request, svc dmarkets.ServiceInterface, params listMarketsParams) ([]*dmarkets.Market, error) {
	return svc.ListMarkets(r.Context(), params.filters)
}

func writeListMarketsResponse(w http.ResponseWriter, r *http.Request, svc dmarkets.ServiceInterface, markets []*dmarkets.Market) error {
	response, err := buildListMarketsResponse(r, svc, markets)
	if err != nil {
		return err
	}
	return writeJSON(w, http.StatusOK, response)
}

func buildListMarketsResponse(r *http.Request, svc dmarkets.ServiceInterface, markets []*dmarkets.Market) (dto.ListMarketsResponse, error) {
	overviews, err := buildMarketOverviewResponses(r.Context(), svc, markets)
	if err != nil {
		log.Printf("Error building market overviews: %v", err)
		return dto.ListMarketsResponse{}, err
	}

	return dto.ListMarketsResponse{
		Markets: overviews,
		Total:   len(overviews),
	}, nil
}

func listMarketsErrorResponse(err error) (string, int) {
	switch err {
	case dmarkets.ErrInvalidInput:
		return "Invalid input parameters", http.StatusBadRequest
	case dmarkets.ErrUnauthorized:
		return "Unauthorized", http.StatusUnauthorized
	default:
		return "Error fetching markets", http.StatusInternalServerError
	}
}

func handleListMarkets(w http.ResponseWriter, r *http.Request, svc dmarkets.ServiceInterface) {
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

	if err := writeListMarketsResponse(w, r, svc, markets); err != nil {
		log.Printf("Error encoding list markets response: %v", err)
		http.Error(w, "Error encoding response", http.StatusInternalServerError)
	}
}
