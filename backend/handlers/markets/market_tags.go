package marketshandlers

import (
	"context"
	"net/http"

	"socialpredict/handlers"
	"socialpredict/handlers/markets/dto"
	dmarkets "socialpredict/internal/domain/markets"
)

type marketTagLister interface {
	ListMarketTags(ctx context.Context, includeInactive bool) ([]dmarkets.MarketTag, error)
}

func ListMarketTagsHandler(svc marketTagLister) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			_ = handlers.WriteFailure(w, http.StatusMethodNotAllowed, handlers.ReasonMethodNotAllowed)
			return
		}
		if svc == nil {
			_ = handlers.WriteFailure(w, http.StatusInternalServerError, handlers.ReasonInternalError)
			return
		}
		tags, err := svc.ListMarketTags(r.Context(), false)
		if err != nil {
			_ = handlers.WriteFailure(w, http.StatusInternalServerError, handlers.ReasonInternalError)
			return
		}
		response := dto.MarketTagsResponse{
			Tags:  marketTagResponsesFromDomain(tags),
			Total: len(tags),
		}
		_ = handlers.WriteResult(w, http.StatusOK, response)
	}
}

func marketTagResponsesFromDomain(tags []dmarkets.MarketTag) []dto.MarketTagResponse {
	responses := make([]dto.MarketTagResponse, 0, len(tags))
	for _, tag := range tags {
		responses = append(responses, marketTagResponseFromDomain(tag))
	}
	return responses
}

func marketTagResponseFromDomain(tag dmarkets.MarketTag) dto.MarketTagResponse {
	return dto.MarketTagResponse{
		ID:          tag.ID,
		Slug:        tag.Slug,
		DisplayName: tag.DisplayName,
		Description: tag.Description,
		ColorKey:    tag.ColorKey,
		SortOrder:   tag.SortOrder,
		IsActive:    tag.IsActive,
	}
}
