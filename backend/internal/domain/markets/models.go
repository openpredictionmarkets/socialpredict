package markets

import (
	"strings"
	"time"

	"socialpredict/models"
)

type modelBetAssembler interface {
	Assemble() models.Bet
}

type publicMarketAssembler interface {
	Assemble() *PublicMarket
}

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

type domainBetAssembler struct {
	bet *Bet
}

func (a domainBetAssembler) Assemble() models.Bet {
	if a.bet == nil {
		return models.Bet{}
	}

	return models.Bet{
		Username: a.bet.Username,
		MarketID: a.bet.MarketID,
		Amount:   a.bet.Amount,
		Outcome:  a.bet.Outcome,
		PlacedAt: a.bet.PlacedAt,
	}
}

// ToModelBet converts the domain bet into the shared model shape used by math services.
func (b Bet) ToModelBet() models.Bet {
	return domainBetAssembler{bet: &b}.Assemble()
}

// ToModelBets converts a domain bet slice into the shared model representation.
func ToModelBets(bets []*Bet) []models.Bet {
	if len(bets) == 0 {
		return []models.Bet{}
	}

	modelBets := make([]models.Bet, len(bets))
	for i, bet := range bets {
		modelBets[i] = domainBetAssembler{bet: bet}.Assemble()
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

type marketPublicAssembler struct {
	target *PublicMarket
	market *Market
}

func (a marketPublicAssembler) Assemble() *PublicMarket {
	if a.market == nil {
		return nil
	}

	target := a.target
	if target == nil {
		target = &PublicMarket{}
	}

	*target = PublicMarket{
		ID:                      a.market.ID,
		QuestionTitle:           a.market.QuestionTitle,
		Description:             a.market.Description,
		OutcomeType:             a.market.OutcomeType,
		ResolutionDateTime:      a.market.ResolutionDateTime,
		FinalResolutionDateTime: a.market.FinalResolutionDateTime,
		UTCOffset:               a.market.UTCOffset,
		IsResolved:              a.market.IsResolved(),
		ResolutionResult:        a.market.ResolutionResult,
		InitialProbability:      a.market.InitialProbability,
		CreatorUsername:         a.market.CreatorUsername,
		CreatedAt:               a.market.CreatedAt,
		YesLabel:                a.market.YesLabel,
		NoLabel:                 a.market.NoLabel,
	}

	return target
}

// FromMarket copies the public fields from a domain market into the public view model.
func (m *PublicMarket) FromMarket(market *Market) *PublicMarket {
	return marketPublicAssembler{target: m, market: market}.Assemble()
}
