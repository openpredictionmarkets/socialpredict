package bets

import (
	"time"

	"socialpredict/internal/domain/boundary"
)

func newPlacedBoundaryBet(req PlaceRequest, outcome string, placedAt time.Time) *boundary.Bet {
	return &boundary.Bet{
		Username: req.Username,
		MarketID: req.MarketID,
		Amount:   req.Amount,
		Outcome:  outcome,
		PlacedAt: placedAt,
	}
}

func newSoldBoundaryBet(req SellRequest, outcome string, sharesSold int64, placedAt time.Time) *boundary.Bet {
	return &boundary.Bet{
		Username: req.Username,
		MarketID: req.MarketID,
		Amount:   -sharesSold,
		Outcome:  outcome,
		PlacedAt: placedAt,
	}
}

// PlaceRequest captures the inputs required to place a buy bet.
type PlaceRequest struct {
	Username string
	MarketID uint
	Amount   int64
	Outcome  string
}

// NewBet builds the persisted bet record for a place request.
func (r PlaceRequest) NewBet(outcome string, placedAt time.Time) *boundary.Bet {
	return newPlacedBoundaryBet(r, outcome, placedAt)
}

// PlacedBet represents the bet that was successfully recorded.
type PlacedBet struct {
	Username string
	MarketID uint
	Amount   int64
	Outcome  string
	PlacedAt time.Time
}

func copyPlacedBet(target *PlacedBet, bet *boundary.Bet) *PlacedBet {
	if bet == nil {
		return nil
	}
	if target == nil {
		target = &PlacedBet{}
	}

	*target = PlacedBet{
		Username: bet.Username,
		MarketID: bet.MarketID,
		Amount:   bet.Amount,
		Outcome:  bet.Outcome,
		PlacedAt: bet.PlacedAt,
	}
	return target
}

// FromBoundary copies the persisted bet fields into the domain result shape.
func (p *PlacedBet) FromBoundary(bet *boundary.Bet) *PlacedBet {
	return copyPlacedBet(p, bet)
}

// FromModel preserves the legacy naming while reading from the boundary record.
func (p *PlacedBet) FromModel(bet *boundary.Bet) *PlacedBet {
	return p.FromBoundary(bet)
}

// SellRequest represents a request to sell shares for credits.
type SellRequest struct {
	Username string
	MarketID uint
	Amount   int64 // credits requested
	Outcome  string
}

// NewSaleBet builds the persisted ledger entry for a share sale.
func (r SellRequest) NewSaleBet(outcome string, sharesSold int64, placedAt time.Time) *boundary.Bet {
	return newSoldBoundaryBet(r, outcome, sharesSold, placedAt)
}

// SellResult summarises the sale that occurred.
type SellResult struct {
	Username      string
	MarketID      uint
	SharesSold    int64
	SaleValue     int64
	Dust          int64
	Outcome       string
	TransactionAt time.Time
}

func buildSellResult(target *SellResult, req SellRequest, outcome string, sale SaleQuote, transactionAt time.Time) *SellResult {
	if target == nil {
		target = &SellResult{}
	}

	*target = SellResult{
		Username:      req.Username,
		MarketID:      req.MarketID,
		SharesSold:    sale.SharesToSell,
		SaleValue:     sale.SaleValue,
		Dust:          sale.Dust,
		Outcome:       outcome,
		TransactionAt: transactionAt,
	}
	return target
}

// Build populates the sell result from the request and calculated sale quote.
func (r *SellResult) Build(req SellRequest, outcome string, sale SaleQuote, transactionAt time.Time) *SellResult {
	return buildSellResult(r, req, outcome, sale, transactionAt)
}
