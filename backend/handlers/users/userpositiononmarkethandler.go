package usershandlers

import (
	"net/http"

	"github.com/gorilla/mux"

	positionshandlers "socialpredict/handlers/positions"
	dmarkets "socialpredict/internal/domain/markets"
	"socialpredict/middleware"
	"socialpredict/util"
)

// UserMarketPositionHandlerWithService returns an HTTP handler that resolves the authenticated user's
// position in the given market by delegating to the shared positions handler.
func UserMarketPositionHandlerWithService(svc dmarkets.ServiceInterface) http.HandlerFunc {
	positionsHandler := positionshandlers.MarketUserPositionHandlerWithService(svc)

	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method is not supported.", http.StatusMethodNotAllowed)
			return
		}

		db := util.GetDB()
		user, httperr := middleware.ValidateTokenAndGetUser(r, db)
		if httperr != nil {
			http.Error(w, "Invalid token: "+httperr.Error(), http.StatusUnauthorized)
			return
		}

		vars := mux.Vars(r)
		if vars == nil {
			vars = map[string]string{}
		}
		vars["username"] = user.Username
		r = mux.SetURLVars(r, vars)

		positionsHandler(w, r)
	}
}
