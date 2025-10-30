package marketshandlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"strconv"
	"strings"

	"socialpredict/handlers/markets/dto"
	dmarkets "socialpredict/internal/domain/markets"
	"socialpredict/logging"
	"socialpredict/middleware"

	"github.com/golang-jwt/jwt/v4"
	"github.com/gorilla/mux"
)

func ResolveMarketHandler(svc dmarkets.ServiceInterface) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logging.LogMsg("Attempting to use ResolveMarketHandler.")

		// 1. Parse {id} path param
		vars := mux.Vars(r)
		marketIdStr := vars["marketId"]

		marketId, err := strconv.ParseInt(marketIdStr, 10, 64)
		if err != nil {
			http.Error(w, "Invalid market ID", http.StatusBadRequest)
			return
		}

		// 2. Parse body into dto.ResolveRequest{Result string}
		var req dto.ResolveMarketRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		username, err := extractUsernameFromRequest(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}

		// 4. Call domain service to resolve market
		err = svc.ResolveMarket(r.Context(), marketId, req.Resolution, username)
		if err != nil {
			// 5. Map errors (not found → 404, invalid → 400/409, forbidden → 403)
			switch err {
			case dmarkets.ErrMarketNotFound:
				http.Error(w, "Market not found", http.StatusNotFound)
			case dmarkets.ErrUnauthorized:
				http.Error(w, "User is not the creator of the market", http.StatusForbidden) // Changed to 403 per spec
			case dmarkets.ErrInvalidState:
				http.Error(w, "Market is already resolved", http.StatusConflict) // 409 Conflict
			case dmarkets.ErrInvalidInput:
				http.Error(w, "Invalid resolution outcome", http.StatusBadRequest) // 400 Bad Request
			default:
				logging.LogMsg("Error resolving market: " + err.Error())
				http.Error(w, "Internal server error", http.StatusInternalServerError)
			}
			return
		}

		// 6. w.WriteHeader(http.StatusNoContent) - per specification
		w.WriteHeader(http.StatusNoContent)
	}
}

func extractUsernameFromRequest(r *http.Request) (string, error) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return "", errors.New("authorization header required")
	}

	tokenString := strings.TrimPrefix(authHeader, "Bearer ")
	token, err := jwt.ParseWithClaims(tokenString, &middleware.UserClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(os.Getenv("JWT_SIGNING_KEY")), nil
	})
	if err != nil || !token.Valid {
		return "", errors.New("invalid token")
	}

	claims, ok := token.Claims.(*middleware.UserClaims)
	if !ok || claims.Username == "" {
		return "", errors.New("invalid token claims")
	}

	return claims.Username, nil
}
