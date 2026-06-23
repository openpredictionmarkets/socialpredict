package marketshandlers

import (
	"context"
	"net/http"

	"socialpredict/handlers"
	"socialpredict/handlers/markets/dto"
	dmarkets "socialpredict/internal/domain/markets"
	"socialpredict/internal/domain/readmodels"
)

type marketSummaryReadModelService interface {
	GetMarketSummaryReadModel(ctx context.Context, marketID int64) (*dmarkets.MarketSummaryReadModel, error)
}

type marketSummaryReadModelResponse struct {
	Market      dto.PublicMarketResponse        `json:"market"`
	Creator     *dto.CreatorResponse            `json:"creator"`
	Probability []dto.ProbabilityChangeResponse `json:"probabilityChanges"`
	NumUsers    int                             `json:"numUsers"`
	TotalVolume int64                           `json:"totalVolume"`
	MarketDust  int64                           `json:"marketDust"`
	Freshness   dto.Freshness                   `json:"freshness"`
}

func readModelFreshnessToResponse(freshness readmodels.Freshness) dto.Freshness {
	return dto.Freshness{
		GeneratedAt:            freshness.GeneratedAt,
		Source:                 freshness.Source,
		TargetFreshnessSeconds: freshness.TargetFreshnessSeconds,
		TransactionSafeRead:    freshness.TransactionSafeRead,
		IsStale:                freshness.IsStale,
		StaleReason:            freshness.StaleReason,
		MarkedStaleAt:          freshness.MarkedStaleAt,
	}
}

func (h *Handler) MarketSummaryReadModel(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeMethodNotAllowed(w)
		return
	}

	marketID, err := parseMarketIDFromRequest(r)
	if err != nil {
		writeInvalidRequest(w)
		return
	}

	service, ok := h.service.(marketSummaryReadModelService)
	if !ok {
		writeInternalError(w)
		return
	}

	summary, err := service.GetMarketSummaryReadModel(r.Context(), marketID)
	if err != nil {
		writeDetailsError(w, err)
		return
	}

	response := marketSummaryReadModelResponse{
		Market:      publicMarketResponseFromDomain(summary.Market),
		Creator:     creatorResponseFromSummary(summary.Creator),
		Probability: probabilityChangesToResponse(summary.Accounting.ProbabilityChanges),
		NumUsers:    summary.Accounting.UserCount,
		TotalVolume: summary.Accounting.VolumeWithDust,
		MarketDust:  summary.Accounting.MarketDust,
		Freshness:   readModelFreshnessToResponse(summary.Accounting.Freshness()),
	}
	_ = handlers.WriteResult(w, http.StatusOK, response)
}
