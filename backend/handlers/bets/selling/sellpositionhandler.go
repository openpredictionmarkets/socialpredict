package sellbetshandlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"socialpredict/handlers"
	"socialpredict/handlers/authhttp"
	"socialpredict/handlers/bets/dto"
	bets "socialpredict/internal/domain/bets"
	dmarkets "socialpredict/internal/domain/markets"
	dusers "socialpredict/internal/domain/users"
	authsvc "socialpredict/internal/service/auth"
	"socialpredict/logger"
)

type readModelInvalidator interface {
	InvalidateAfterMarketTransaction(ctx context.Context, username string, marketID int64, reason string) error
}

// SellPositionHandler returns an HTTP handler that delegates sales to the bets service.
func SellPositionHandler(betsSvc bets.ServiceInterface, usersSvc dusers.ServiceInterface) http.HandlerFunc {
	return SellPositionHandlerWithInvalidator(betsSvc, usersSvc, nil)
}

// SellPositionHandlerWithInvalidator returns an HTTP handler that delegates
// sales and then marks display read models stale after a successful write.
func SellPositionHandlerWithInvalidator(betsSvc bets.ServiceInterface, usersSvc dusers.ServiceInterface, invalidator readModelInvalidator) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			_ = handlers.WriteFailure(w, http.StatusMethodNotAllowed, handlers.ReasonMethodNotAllowed)
			return
		}

		user, authErr := authsvc.ValidateUserAndEnforcePasswordChangeGetUser(r, usersSvc)
		if authErr != nil {
			_ = authhttp.WriteFailure(w, authErr)
			return
		}

		req, decodeErr := decodeSellRequest(r)
		if decodeErr != nil {
			_ = handlers.WriteFailure(w, http.StatusBadRequest, handlers.ReasonInvalidRequest)
			return
		}

		result, err := betsSvc.Sell(r.Context(), toSellRequest(req, user.Username))
		if err != nil {
			handleSellError(w, err)
			return
		}

		writeSellResponse(w, result)
		if invalidator != nil {
			if err := invalidator.InvalidateAfterMarketTransaction(r.Context(), result.Username, int64(result.MarketID), "sale_accepted"); err != nil {
				logger.LogError("SellPosition", "InvalidateReadModels", err)
			}
		}
	}
}

// SellQuoteHandler returns an HTTP handler that previews a sale without settlement.
func SellQuoteHandler(betsSvc bets.ServiceInterface, usersSvc dusers.ServiceInterface) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			_ = handlers.WriteFailure(w, http.StatusMethodNotAllowed, handlers.ReasonMethodNotAllowed)
			return
		}

		user, authErr := authsvc.ValidateUserAndEnforcePasswordChangeGetUser(r, usersSvc)
		if authErr != nil {
			_ = authhttp.WriteFailure(w, authErr)
			return
		}

		req, decodeErr := decodeSellRequest(r)
		if decodeErr != nil {
			_ = handlers.WriteFailure(w, http.StatusBadRequest, handlers.ReasonInvalidRequest)
			return
		}

		result, err := betsSvc.QuoteSell(r.Context(), toSellRequest(req, user.Username))
		if err != nil {
			handleSellError(w, err)
			return
		}

		writeSellQuoteResponse(w, result)
	}
}

func decodeSellRequest(r *http.Request) (dto.SellBetRequest, error) {
	var req dto.SellBetRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return dto.SellBetRequest{}, errors.New("Invalid request body")
	}
	return req, nil
}

func toSellRequest(req dto.SellBetRequest, username string) bets.SellRequest {
	return bets.SellRequest{
		Username: username,
		MarketID: req.MarketID,
		Amount:   req.Amount,
		Outcome:  req.Outcome,
	}
}

func handleSellError(w http.ResponseWriter, err error) {
	var dustErr bets.ErrDustCapExceeded
	if errors.As(err, &dustErr) {
		_ = handlers.WriteFailureWithDetails(
			w,
			http.StatusUnprocessableEntity,
			handlers.ReasonDustCapExceeded,
			"Sale would create too much dust. Dust is the small rounding remainder retained by the market when sale proceeds are rounded to whole shares. Submit a different credit amount and try again.",
			map[string]any{
				"dust":    dustErr.Requested,
				"maxDust": dustErr.Cap,
				"hint":    "Try lowering or adjusting the requested credit amount so the rounding dust is within the configured cap.",
			},
		)
		return
	}

	var projectionErr bets.SaleProjectionNotExecutableError
	if errors.As(err, &projectionErr) {
		_ = handlers.WriteFailureWithDetails(
			w,
			http.StatusUnprocessableEntity,
			handlers.ReasonInsufficientShares,
			bets.ProjectionInexecutableSaleMessage,
			saleProjectionDetails(projectionErr.Details),
		)
		return
	}

	switch {
	case errors.Is(err, bets.ErrInvalidOutcome), errors.Is(err, bets.ErrInvalidAmount):
		_ = handlers.WriteFailure(w, http.StatusBadRequest, handlers.ReasonValidationFailed)
	case errors.Is(err, bets.ErrMarketClosed):
		_ = handlers.WriteFailure(w, http.StatusConflict, handlers.ReasonMarketClosed)
	case errors.Is(err, bets.ErrNoSellableShares):
		_ = handlers.WriteFailureWithDetails(
			w,
			http.StatusUnprocessableEntity,
			handlers.ReasonNoPosition,
			bets.NoSellableSharesMessage,
			map[string]any{
				"hint": "A later buy from another user is required before this value can be sold.",
			},
		)
	case errors.Is(err, bets.ErrNoPosition):
		_ = handlers.WriteFailureWithDetails(
			w,
			http.StatusUnprocessableEntity,
			handlers.ReasonNoPosition,
			bets.NoSellableSharesMessage,
			map[string]any{
				"hint": "A later buy from another user is required before this value can be sold.",
			},
		)
	case errors.Is(err, bets.ErrInsufficientShares):
		_ = handlers.WriteFailureWithDetails(
			w,
			http.StatusUnprocessableEntity,
			handlers.ReasonInsufficientShares,
			bets.InsufficientSellableSharesMessage,
			map[string]any{
				"hint": "Use a smaller sale amount or wait for another user's later buy to unlock the newest value.",
			},
		)
	case errors.Is(err, dmarkets.ErrMarketNotFound):
		_ = handlers.WriteFailure(w, http.StatusNotFound, handlers.ReasonMarketNotFound)
	default:
		_ = handlers.WriteFailure(w, http.StatusInternalServerError, handlers.ReasonInternalError)
	}
}

func saleProjectionDetails(details bets.SaleProjectionDetails) map[string]any {
	return map[string]any{
		"outcome":                      details.Outcome,
		"requestedCredits":             details.RequestedCredits,
		"positionValue":                details.PositionValue,
		"positionOutcomeShares":        details.PositionOutcomeShares,
		"nominalUnlockedValue":         details.NominalUnlockedValue,
		"nominalUnlockedOutcomeShares": details.NominalUnlockedOutcomeShares,
		"projectedPositionValue":       details.ProjectedPositionValue,
		"projectedOutcomeShares":       details.ProjectedOutcomeShares,
		"executableSaleValue":          details.ExecutableSaleValue,
		"hint":                         "This Position has value, but the requested Sale Order does not currently reduce the backend-projected position enough to pay credits safely.",
	}
}

func writeSellResponse(w http.ResponseWriter, result *bets.SellResult) {
	response := dto.SellBetResponse{
		Username:      result.Username,
		MarketID:      result.MarketID,
		SharesSold:    result.SharesSold,
		SaleValue:     result.SaleValue,
		Dust:          result.Dust,
		NetProceeds:   result.NetProceeds,
		Outcome:       result.Outcome,
		TransactionAt: result.TransactionAt,
	}

	_ = handlers.WriteResult(w, http.StatusCreated, response)
}

func writeSellQuoteResponse(w http.ResponseWriter, result *bets.SellQuoteResult) {
	response := dto.SellQuoteResponse{
		Username:          result.Username,
		MarketID:          result.MarketID,
		Outcome:           result.Outcome,
		RequestedCredits:  result.RequestedCredits,
		SharesSold:        result.SharesSold,
		SaleValue:         result.SaleValue,
		Dust:              result.Dust,
		NetProceeds:       result.NetProceeds,
		MaxDust:           result.MaxDust,
		ValuePerShare:     result.ValuePerShare,
		DustCapCoverage:   result.DustCapCoverage,
		Allowed:           result.Allowed,
		SuggestedAmounts:  result.SuggestedAmounts,
		Message:           result.Message,
		QuotedAt:          result.QuotedAt,
		DustCapExceeded:   result.DustCapExceeded,
		DustCapExceededBy: result.DustCapExceededBy,
	}

	_ = handlers.WriteResult(w, http.StatusOK, response)
}
