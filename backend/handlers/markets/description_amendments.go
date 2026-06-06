package marketshandlers

import (
	"context"
	"encoding/json"
	"net/http"

	"socialpredict/handlers/markets/dto"
	dmarkets "socialpredict/internal/domain/markets"
)

type descriptionAmendmentProposer interface {
	ProposeMarketDescriptionAmendment(ctx context.Context, marketID int64, actorUsername string, req dmarkets.MarketDescriptionAmendmentRequest) (*dmarkets.MarketDescriptionAmendment, error)
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
