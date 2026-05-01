package marketshandlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"

	"socialpredict/handlers/markets/dto"
	dmarkets "socialpredict/internal/domain/markets"
	authsvc "socialpredict/internal/service/auth"
	"socialpredict/logger"

	"github.com/golang-jwt/jwt/v4"
	"github.com/gorilla/mux"
)

func ResolveMarketHandler(svc dmarkets.ServiceInterface) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		marketId, req, err := parseResolveRequest(r)
		if err != nil {
			logger.LogWarn("ResolveMarket", "ParseResolveRequest", err.Error())
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		username, err := extractUsernameFromRequest(r)
		if err != nil {
			logger.LogWarn("ResolveMarket", "ExtractUsernameFromRequest", err.Error())
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}

		if err := svc.ResolveMarket(r.Context(), marketId, req.Resolution, username); err != nil {
			writeResolveError(w, err)
			logResolveMarketFailure(marketId, username, err)
			return
		}

		logger.LogInfo("ResolveMarket", "ResolveMarket", fmt.Sprintf("Resolved market %d by user %s", marketId, username))
		w.WriteHeader(http.StatusNoContent)
	}
}

func parseResolveRequest(r *http.Request) (int64, dto.ResolveMarketRequest, error) {
	var req dto.ResolveMarketRequest

	marketIdStr := mux.Vars(r)["marketId"]
	marketId, err := strconv.ParseInt(marketIdStr, 10, 64)
	if err != nil {
		return 0, req, fmt.Errorf("Invalid market ID")
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return 0, req, fmt.Errorf("Invalid request body")
	}
	return marketId, req, nil
}

func writeResolveError(w http.ResponseWriter, err error) {
	switch err {
	case dmarkets.ErrMarketNotFound:
		http.Error(w, "Market not found", http.StatusNotFound)
	case dmarkets.ErrUnauthorized:
		http.Error(w, "User is not the creator of the market", http.StatusForbidden)
	case dmarkets.ErrInvalidState:
		http.Error(w, "Market is already resolved", http.StatusConflict)
	case dmarkets.ErrInvalidInput:
		http.Error(w, "Invalid resolution outcome", http.StatusBadRequest)
	default:
		logger.LogError("ResolveMarket", "writeResolveError", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

func logResolveMarketFailure(marketID int64, username string, err error) {
	message := fmt.Sprintf("Market %d for user %s: %v", marketID, username, err)

	switch err {
	case dmarkets.ErrMarketNotFound, dmarkets.ErrUnauthorized, dmarkets.ErrInvalidState, dmarkets.ErrInvalidInput:
		logger.LogWarn("ResolveMarket", "ResolveMarket", message)
	default:
		logger.LogError("ResolveMarket", "ResolveMarket", err)
	}
}

func extractUsernameFromRequest(r *http.Request) (string, error) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return "", errors.New("authorization header required")
	}

	tokenString := strings.TrimPrefix(authHeader, "Bearer ")
	token, err := jwt.ParseWithClaims(tokenString, &authsvc.UserClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(os.Getenv("JWT_SIGNING_KEY")), nil
	})
	if err != nil || !token.Valid {
		return "", errors.New("invalid token")
	}

	claims, ok := token.Claims.(*authsvc.UserClaims)
	if !ok || claims.Username == "" {
		return "", errors.New("invalid token claims")
	}

	return claims.Username, nil
}
