package marketshandlers

import (
	"net/http"
	"strings"

	"socialpredict/handlers/markets/dto"
	dmarkets "socialpredict/internal/domain/markets"
	"socialpredict/logger"
	"socialpredict/security"
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
			Status:    status,
			CreatedBy: strings.TrimSpace(r.URL.Query().Get("created_by")),
			Limit:     page.Limit,
			Offset:    page.Offset,
		},
		page: page,
	}, nil
}

func parseListLimit(rawLimit string) int {
	return security.ParseBoundedIntParam(rawLimit, 50, 1, 100)
}

func parseListOffset(rawOffset string) int {
	return security.ParseBoundedIntParam(rawOffset, 0, 0, int(^uint(0)>>1))
}

func fetchMarkets(r *http.Request, svc dmarkets.ServiceInterface, params listMarketsParams) ([]*dmarkets.Market, error) {
	if params.filters.Status != "" {
		return fetchMarketsByStatus(r, svc, params)
	}
	return fetchAllMarkets(r, svc, params)
}

func writeListMarketsError(w http.ResponseWriter, err error) {
	if !isListMarketsKnownError(err) {
		logger.LogError("ListMarkets", "FetchMarkets", err)
	}
	writeListError(w, err)
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
		logger.LogError("ListMarkets", "BuildMarketOverviewResponses", err)
		return dto.ListMarketsResponse{}, err
	}

	return dto.ListMarketsResponse{
		Markets: overviews,
		Total:   len(overviews),
	}, nil
}

func isListMarketsKnownError(err error) bool {
	switch err {
	case dmarkets.ErrInvalidInput, dmarkets.ErrUnauthorized:
		return true
	default:
		return false
	}
}

func handleListMarkets(w http.ResponseWriter, r *http.Request, svc dmarkets.ServiceInterface) {
	if r.Method != http.MethodGet {
		writeMethodNotAllowed(w)
		return
	}

	params, parseErr := parseListMarketsParams(r)
	if parseErr != nil {
		writeInvalidRequest(w)
		return
	}

	markets, err := fetchMarkets(r, svc, params)
	if err != nil {
		writeListMarketsError(w, err)
		return
	}

	if err := writeListMarketsResponse(w, r, svc, markets); err != nil {
		if isRequestCanceled(err) {
			return
		}
		logger.LogError("ListMarkets", "WriteResponse", err)
		writeInternalError(w)
	}
}
