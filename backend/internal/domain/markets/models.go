package markets

import (
	"strings"
	"time"

	"socialpredict/models"
)

// Market represents the core market domain model
type Market struct {
	ID                      int64
	QuestionTitle           string
	Description             string
	OutcomeType             string
	ResolutionDateTime      time.Time
	FinalResolutionDateTime time.Time
	ResolutionResult        string
	CreatorUsername         string
	YesLabel                string
	NoLabel                 string
	Status                  string
	CreatedAt               time.Time
	UpdatedAt               time.Time
	InitialProbability      float64
	UTCOffset               int
}

// CreatedBy reports whether the market belongs to the supplied creator username.
func (m *Market) CreatedBy(username string) bool {
	if m == nil {
		return false
	}
	return strings.TrimSpace(m.CreatorUsername) == strings.TrimSpace(username)
}

// IsResolved reports whether the market is in the resolved state.
func (m *Market) IsResolved() bool {
	if m == nil {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(m.Status), "resolved")
}

// MarketCreateRequest represents the request to create a new market
type MarketCreateRequest struct {
	QuestionTitle      string
	Description        string
	OutcomeType        string
	ResolutionDateTime time.Time
	YesLabel           string
	NoLabel            string
}

// HasCustomLabels reports whether the create request includes either custom label.
func (r MarketCreateRequest) HasCustomLabels() bool {
	return strings.TrimSpace(r.YesLabel) != "" || strings.TrimSpace(r.NoLabel) != ""
}

// UserPosition represents a user's holdings within a market.
type UserPosition struct {
	Username         string
	MarketID         int64
	YesSharesOwned   int64
	NoSharesOwned    int64
	Value            int64
	TotalSpent       int64
	TotalSpentInPlay int64
	IsResolved       bool
	ResolutionResult string
}

// MarketPositions aggregates user positions for a market.
type MarketPositions []*UserPosition

// Normalize returns a non-nil collection for callers that treat empty and nil identically.
func (p MarketPositions) Normalize() MarketPositions {
	if p == nil {
		return MarketPositions{}
	}
	return p
}

// Bet represents a wager placed within a market.
type Bet struct {
	ID        uint
	Username  string
	MarketID  uint
	Amount    int64
	Outcome   string
	PlacedAt  time.Time
	CreatedAt time.Time
}

func modelBetFromDomain(bet *Bet) models.Bet {
	if bet == nil {
		return models.Bet{}
	}

	return models.Bet{
		Username: bet.Username,
		MarketID: bet.MarketID,
		Amount:   bet.Amount,
		Outcome:  bet.Outcome,
		PlacedAt: bet.PlacedAt,
	}
}

// ToModelBet converts the domain bet into the shared model shape used by math services.
func (b Bet) ToModelBet() models.Bet {
	return modelBetFromDomain(&b)
}

// ToModelBets converts a domain bet slice into the shared model representation.
func ToModelBets(bets []*Bet) []models.Bet {
	if len(bets) == 0 {
		return []models.Bet{}
	}

	modelBets := make([]models.Bet, len(bets))
	for i, bet := range bets {
		modelBets[i] = modelBetFromDomain(bet)
	}

	return modelBets
}

// PayoutPosition captures the resolved valuation per user for distribution.
type PayoutPosition struct {
	Username string
	Value    int64
}

// IsPayable reports whether the payout position results in a positive transfer.
func (p *PayoutPosition) IsPayable() bool {
	return p != nil && p.Value > 0
}

// PublicMarket represents the public view of a market.
type PublicMarket struct {
	ID                      int64
	QuestionTitle           string
	Description             string
	OutcomeType             string
	ResolutionDateTime      time.Time
	FinalResolutionDateTime time.Time
	UTCOffset               int
	IsResolved              bool
	ResolutionResult        string
	InitialProbability      float64
	CreatorUsername         string
	CreatedAt               time.Time
	YesLabel                string
	NoLabel                 string
}

// Resolved reports whether the public market already has a terminal resolution.
func (m *PublicMarket) Resolved() bool {
	return m != nil && m.IsResolved
}

func copyPublicMarket(target *PublicMarket, market *Market) *PublicMarket {
	if market == nil {
		return nil
	}

	if target == nil {
		target = &PublicMarket{}
	}

	*target = PublicMarket{
		ID:                      market.ID,
		QuestionTitle:           market.QuestionTitle,
		Description:             market.Description,
		OutcomeType:             market.OutcomeType,
		ResolutionDateTime:      market.ResolutionDateTime,
		FinalResolutionDateTime: market.FinalResolutionDateTime,
		UTCOffset:               market.UTCOffset,
		IsResolved:              market.IsResolved(),
		ResolutionResult:        market.ResolutionResult,
		InitialProbability:      market.InitialProbability,
		CreatorUsername:         market.CreatorUsername,
		CreatedAt:               market.CreatedAt,
		YesLabel:                market.YesLabel,
		NoLabel:                 market.NoLabel,
	}

	return target
}

// FromMarket copies the public fields from a domain market into the public view model.
func (m *PublicMarket) FromMarket(market *Market) *PublicMarket {
	return copyPublicMarket(m, market)
}
