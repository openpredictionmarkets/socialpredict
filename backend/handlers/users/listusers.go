package usershandlers

import (
	"context"

	dusers "socialpredict/internal/domain/users"
)

// ListUserMarkets returns markets that the specified user participates in via the users service.
func ListUserMarkets(ctx context.Context, svc dusers.ServiceInterface, userID int64) ([]*dusers.UserMarket, error) {
	return svc.ListUserMarkets(ctx, userID)
}
