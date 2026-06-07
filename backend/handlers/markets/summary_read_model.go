package marketshandlers

import (
	"context"
	"net/http"
	"time"

	"socialpredict/handlers"
	"socialpredict/handlers/markets/dto"
	dmarkets "socialpredict/internal/domain/markets"
	"socialpredict/internal/domain/readmodels"
	"socialpredict/logger"
)

type marketAccountingSnapshotRefresher interface {
	RefreshMarketAccountingSnapshot(ctx context.Context, marketID int64) (*dmarkets.MarketAccountingSnapshot, error)
}

type marketSummaryReadModelResponse struct {
	Market      dto.PublicMarketResponse        `json:"market"`
	Creator     *dto.CreatorResponse            `json:"creator"`
	Probability []dto.ProbabilityChangeResponse `json:"probabilityChanges"`
	NumUsers    int                             `json:"numUsers"`
	TotalVolume int64                           `json:"totalVolume"`
	MarketDust  int64                           `json:"marketDust"`
	Freshness   readModelFreshnessResponse      `json:"freshness"`
}

type readModelFreshnessResponse struct {
	GeneratedAt            time.Time  `json:"generatedAt"`
	Source                 string     `json:"source"`
	TargetFreshnessSeconds int        `json:"targetFreshnessSeconds"`
	TransactionSafeRead    bool       `json:"transactionSafeRead"`
	IsStale                bool       `json:"isStale"`
	StaleReason            string     `json:"staleReason,omitempty"`
	MarkedStaleAt          *time.Time `json:"markedStaleAt,omitempty"`
}

func readModelFreshnessToResponse(freshness readmodels.Freshness) readModelFreshnessResponse {
	return readModelFreshnessResponse{
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

	details, err := h.service.GetMarketDetails(r.Context(), marketID)
	if err != nil {
		writeDetailsError(w, err)
		return
	}

	freshness := readmodels.NewFreshness(details.Market.UpdatedAt, "live", dmarkets.MarketAccountingSnapshotTargetFreshness, false)
	if refresher, ok := h.service.(marketAccountingSnapshotRefresher); ok {
		if snapshot, refreshErr := refresher.RefreshMarketAccountingSnapshot(r.Context(), marketID); refreshErr == nil && snapshot != nil {
			freshness = snapshot.Freshness()
		} else if refreshErr != nil {
			logger.LogError("MarketSummaryReadModel", "RefreshMarketAccountingSnapshot", refreshErr)
		}
	}

	response := marketSummaryReadModelResponse{
		Market:      publicMarketResponseFromDomain(details.Market),
		Creator:     creatorResponseFromSummary(details.Creator),
		Probability: probabilityChangesToResponse(details.ProbabilityChanges),
		NumUsers:    details.NumUsers,
		TotalVolume: details.TotalVolume,
		MarketDust:  details.MarketDust,
		Freshness:   readModelFreshnessToResponse(freshness),
	}
	_ = handlers.WriteResult(w, http.StatusOK, response)
}
