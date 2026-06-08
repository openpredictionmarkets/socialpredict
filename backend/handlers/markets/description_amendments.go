package marketshandlers

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"

	"socialpredict/handlers/markets/dto"
	dmarkets "socialpredict/internal/domain/markets"
)

type descriptionAmendmentProposer interface {
	ProposeMarketDescriptionAmendment(ctx context.Context, marketID int64, actorUsername string, req dmarkets.MarketDescriptionAmendmentRequest) (*dmarkets.MarketDescriptionAmendment, error)
}

type descriptionAmendmentLister interface {
	ListMarketDescriptionAmendments(ctx context.Context, filters dmarkets.MarketDescriptionAmendmentFilters) ([]dmarkets.MarketDescriptionAmendment, error)
}

func (h *Handler) ProposeDescriptionAmendment(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeMethodNotAllowed(w)
		return
	}
	if h.auth == nil {
		writeInternalError(w)
		return
	}
	user, authErr := h.auth.CurrentUser(r)
	if authErr != nil {
		writeAuthError(w, authErr)
		return
	}
	marketID, err := parseMarketIDFromRequest(r)
	if err != nil {
		writeInvalidRequest(w)
		return
	}
	var req dto.MarketDescriptionAmendmentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeInvalidRequest(w)
		return
	}
	proposer, ok := h.service.(descriptionAmendmentProposer)
	if !ok {
		writeInternalError(w)
		return
	}
	amendment, err := proposer.ProposeMarketDescriptionAmendment(r.Context(), marketID, user.Username, dmarkets.MarketDescriptionAmendmentRequest{
		Body:         req.Body,
		BodyFormat:   req.BodyFormat,
		SubmitReason: req.SubmitReason,
	})
	if err != nil {
		writeMarketActionError(w, err)
		return
	}
	_ = writeJSON(w, http.StatusCreated, descriptionAmendmentsToResponse([]dmarkets.MarketDescriptionAmendment{*amendment})[0])
}

func (h *Handler) ListMyDescriptionAmendments(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeMethodNotAllowed(w)
		return
	}
	if h.auth == nil {
		writeInternalError(w)
		return
	}
	user, authErr := h.auth.CurrentUser(r)
	if authErr != nil {
		writeAuthError(w, authErr)
		return
	}
	lister, ok := h.service.(descriptionAmendmentLister)
	if !ok {
		writeInternalError(w)
		return
	}
	query := r.URL.Query()
	limit := boundedQueryInt(query.Get("limit"), 50, 1, 200)
	offset := boundedQueryInt(query.Get("offset"), 0, 0, 100000)
	items, err := lister.ListMarketDescriptionAmendments(r.Context(), dmarkets.MarketDescriptionAmendmentFilters{
		Status:    query.Get("status"),
		MarketID:  parsePositiveInt64(query.Get("marketId")),
		CreatedBy: user.Username,
		Limit:     limit,
		Offset:    offset,
	})
	if err != nil {
		writeMarketActionError(w, err)
		return
	}
	_ = writeJSON(w, http.StatusOK, struct {
		Amendments []dto.MarketDescriptionAmendmentResponse `json:"amendments"`
		Total      int                                      `json:"total"`
	}{
		Amendments: descriptionAmendmentsToResponse(items),
		Total:      len(items),
	})
}

func parsePositiveInt64(raw string) int64 {
	if raw == "" {
		return 0
	}
	value, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || value < 0 {
		return 0
	}
	return value
}
