package markets

import (
	"time"

	"socialpredict/internal/domain/boundary"
)

// MarketAccountingSnapshot is a display/read-model snapshot. It must not be
// used as transaction truth for buy/sell/order settlement decisions.
type MarketAccountingSnapshot struct {
	MarketID            int64
	GeneratedAt         time.Time
	ProbabilityChanges  []ProbabilityPoint
	LastProbability     float64
	NetBetVolume        int64
	MarketDust          int64
	VolumeWithDust      int64
	UserCount           int
	BetCount            int
	LastProcessedBetID  uint
	LastProcessedBetAt  time.Time
	Source              string
	TransactionSafeRead bool
}

type MarketAccountingSnapshotCalculator struct {
	probabilities ProbabilityEngine
	metrics       MetricsCalculator
	clock         Clock
}

// NewMarketAccountingSnapshotCalculator builds a calculator for display-safe
// market accounting snapshots. The resulting snapshots are read models, not
// transaction truth.
func NewMarketAccountingSnapshotCalculator(probabilities ProbabilityEngine, metrics MetricsCalculator, clock Clock) MarketAccountingSnapshotCalculator {
	return MarketAccountingSnapshotCalculator{
		probabilities: probabilityEngineOrDefault(probabilities),
		metrics:       metricsCalculatorOrDefault(metrics),
		clock:         clockOrDefault(clock),
	}
}

func (c MarketAccountingSnapshotCalculator) Calculate(market *Market, bets []boundary.Bet) MarketAccountingSnapshot {
	c = NewMarketAccountingSnapshotCalculator(c.probabilities, c.metrics, c.clock)

	marketID := int64(0)
	createdAt := c.clock.Now()
	if market != nil {
		marketID = market.ID
		createdAt = market.CreatedAt
	}

	probabilityChanges := c.probabilities.Calculate(createdAt, bets)
	probabilityPoints := make([]ProbabilityPoint, len(probabilityChanges))
	for i, change := range probabilityChanges {
		probabilityPoints[i] = ProbabilityPoint{
			Probability: change.Probability,
			Timestamp:   change.Timestamp,
		}
	}

	lastProbability := 0.0
	if len(probabilityPoints) > 0 {
		lastProbability = probabilityPoints[len(probabilityPoints)-1].Probability
	}

	lastBetID, lastBetAt := lastProcessedBet(bets)

	return MarketAccountingSnapshot{
		MarketID:            marketID,
		GeneratedAt:         c.clock.Now(),
		ProbabilityChanges:  probabilityPoints,
		LastProbability:     lastProbability,
		NetBetVolume:        c.metrics.Volume(bets),
		MarketDust:          c.metrics.Dust(bets),
		VolumeWithDust:      c.metrics.VolumeWithDust(bets),
		UserCount:           countUniqueUsers(bets),
		BetCount:            len(bets),
		LastProcessedBetID:  lastBetID,
		LastProcessedBetAt:  lastBetAt,
		Source:              "read_model",
		TransactionSafeRead: false,
	}
}

func lastProcessedBet(bets []boundary.Bet) (uint, time.Time) {
	var lastID uint
	var lastAt time.Time
	for _, bet := range bets {
		if bet.ID > lastID {
			lastID = uint(bet.ID)
		}
		if bet.PlacedAt.After(lastAt) {
			lastAt = bet.PlacedAt
		}
	}
	return lastID, lastAt
}
