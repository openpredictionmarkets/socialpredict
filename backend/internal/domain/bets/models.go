package bets

import (
	"time"

	"socialpredict/models"
)

type betModelBuilder interface {
	Build() *models.Bet
}

type betResultAssembler[T any] interface {
	Assemble() *T
}

type modelBetBuilder struct {
	username string
	marketID uint
	amount   int64
	outcome  string
	placedAt time.Time
}

func (b modelBetBuilder) Build() *models.Bet {
	return &models.Bet{
		Username: b.username,
		MarketID: b.marketID,
		Amount:   b.amount,
		Outcome:  b.outcome,
		PlacedAt: b.placedAt,
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
func (r PlaceRequest) NewBet(outcome string, placedAt time.Time) *models.Bet {
	return modelBetBuilder{
		username: r.Username,
		marketID: r.MarketID,
		amount:   r.Amount,
		outcome:  outcome,
		placedAt: placedAt,
	}.Build()
}

// PlacedBet represents the bet that was successfully recorded.
type PlacedBet struct {
	Username string
	MarketID uint
	Amount   int64
	Outcome  string
	PlacedAt time.Time
}

type placedBetAssembler struct {
	target *PlacedBet
	bet    *models.Bet
}

func (a placedBetAssembler) Assemble() *PlacedBet {
	if a.bet == nil {
		return nil
	}

	target := a.target
	if target == nil {
		target = &PlacedBet{}
	}

	*target = PlacedBet{
		Username: a.bet.Username,
		MarketID: a.bet.MarketID,
		Amount:   a.bet.Amount,
		Outcome:  a.bet.Outcome,
		PlacedAt: a.bet.PlacedAt,
	}
	return target
}

// FromModel copies the persisted bet fields into the domain result shape.
func (p *PlacedBet) FromModel(bet *models.Bet) *PlacedBet {
	return placedBetAssembler{target: p, bet: bet}.Assemble()
}

// SellRequest represents a request to sell shares for credits.
type SellRequest struct {
	Username string
	MarketID uint
	Amount   int64 // credits requested
	Outcome  string
}

// NewSaleBet builds the persisted ledger entry for a share sale.
func (r SellRequest) NewSaleBet(outcome string, sharesSold int64, placedAt time.Time) *models.Bet {
	return modelBetBuilder{
		username: r.Username,
		marketID: r.MarketID,
		amount:   -sharesSold,
		outcome:  outcome,
		placedAt: placedAt,
	}.Build()
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

type sellResultAssembler struct {
	target        *SellResult
	req           SellRequest
	outcome       string
	sale          SaleQuote
	transactionAt time.Time
}

func (a sellResultAssembler) Assemble() *SellResult {
	target := a.target
	if target == nil {
		target = &SellResult{}
	}

	*target = SellResult{
		Username:      a.req.Username,
		MarketID:      a.req.MarketID,
		SharesSold:    a.sale.SharesToSell,
		SaleValue:     a.sale.SaleValue,
		Dust:          a.sale.Dust,
		Outcome:       a.outcome,
		TransactionAt: a.transactionAt,
	}
	return target
}

// Build populates the sell result from the request and calculated sale quote.
func (r *SellResult) Build(req SellRequest, outcome string, sale SaleQuote, transactionAt time.Time) *SellResult {
	return sellResultAssembler{
		target:        r,
		req:           req,
		outcome:       outcome,
		sale:          sale,
		transactionAt: transactionAt,
	}.Assemble()
}
