package middleware

import (
	"context"
	"net/http"
	"socialpredict/models"

	"socialpredict/util"
)

type contextKey string

const ContextAPIUser contextKey = "apiUser"

func APIKeyAuthMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			apiKey := r.Header.Get("X-API-Key")
			if apiKey == "" {
				http.Error(w, "Missing API Key", http.StatusUnauthorized)
				return
			}

			db := util.GetDB()

			var user models.User
			if err := db.Where("api_key = ?", apiKey).First(&user).Error; err != nil {
				http.Error(w, "Invalid API Key", http.StatusUnauthorized)
				return
			}

			// Store user in context if needed downstream
			ctx := context.WithValue(r.Context(), ContextAPIUser, &user)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func GetAPIUserFromContext(r *http.Request) *models.User {
	user, ok := r.Context().Value(ContextAPIUser).(*models.User)
	if !ok {
		return nil
	}
	return user
}
