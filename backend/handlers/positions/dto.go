package positions

import dmarkets "socialpredict/internal/domain/markets"

// userPositionResponse defines the JSON shape returned to clients.
type userPositionResponse struct {
	Username         string `json:"username"`
	MarketID         int64  `json:"marketId"`
	YesSharesOwned   int64  `json:"yesSharesOwned"`
	NoSharesOwned    int64  `json:"noSharesOwned"`
	Value            int64  `json:"value"`
	TotalSpent       int64  `json:"totalSpent"`
	TotalSpentInPlay int64  `json:"totalSpentInPlay"`
	IsResolved       bool   `json:"isResolved"`
	ResolutionResult string `json:"resolutionResult"`
}

func newUserPositionResponse(pos *dmarkets.UserPosition) userPositionResponse {
	if pos == nil {
		return userPositionResponse{}
	}

	return userPositionResponse{
		Username:         pos.Username,
		MarketID:         pos.MarketID,
		YesSharesOwned:   pos.YesSharesOwned,
		NoSharesOwned:    pos.NoSharesOwned,
		Value:            pos.Value,
		TotalSpent:       pos.TotalSpent,
		TotalSpentInPlay: pos.TotalSpentInPlay,
		IsResolved:       pos.IsResolved,
		ResolutionResult: pos.ResolutionResult,
	}
}
