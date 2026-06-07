package markets

import (
	"context"
	"errors"
	"time"

	dmarkets "socialpredict/internal/domain/markets"
	"socialpredict/models"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// GetMarketAccountingSnapshot returns the latest durable display/read-model
// accounting snapshot for a market. A missing snapshot is not a market error.
func (r *GormRepository) GetMarketAccountingSnapshot(ctx context.Context, marketID int64) (*dmarkets.MarketAccountingSnapshot, error) {
	if marketID <= 0 {
		return nil, dmarkets.ErrInvalidInput
	}

	var snapshot models.MarketAccountingSnapshot
	if err := r.db.WithContext(ctx).
		Where("market_id = ?", marketID).
		First(&snapshot).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return modelAccountingSnapshotToDomain(&snapshot), nil
}

// UpsertMarketAccountingSnapshot stores a market accounting display snapshot.
// It intentionally persists read-model values only; transaction paths must
// continue reading canonical market/bet/user state.
func (r *GormRepository) UpsertMarketAccountingSnapshot(ctx context.Context, snapshot dmarkets.MarketAccountingSnapshot) error {
	if snapshot.MarketID <= 0 {
		return dmarkets.ErrInvalidInput
	}

	dbSnapshot := domainAccountingSnapshotToModel(snapshot)
	return r.db.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: "market_id"}},
			DoUpdates: clause.AssignmentColumns([]string{
				"last_probability",
				"net_bet_volume",
				"market_dust",
				"volume_with_dust",
				"user_count",
				"bet_count",
				"last_processed_bet_id",
				"last_processed_bet_at",
				"generated_at",
				"source",
				"is_stale",
				"stale_reason",
				"marked_stale_at",
				"updated_at",
			}),
		}).
		Create(&dbSnapshot).Error
}

// MarkMarketAccountingSnapshotStale marks an existing display snapshot stale
// after a canonical mutation. Missing snapshots are ignored because refresh can
// recreate them later.
func (r *GormRepository) MarkMarketAccountingSnapshotStale(ctx context.Context, marketID int64, reason string) error {
	if marketID <= 0 {
		return dmarkets.ErrInvalidInput
	}
	now := time.Now().UTC()
	return r.db.WithContext(ctx).
		Model(&models.MarketAccountingSnapshot{}).
		Where("market_id = ?", marketID).
		Updates(map[string]interface{}{
			"is_stale":        true,
			"stale_reason":    reason,
			"marked_stale_at": now,
			"updated_at":      now,
		}).Error
}

func domainAccountingSnapshotToModel(snapshot dmarkets.MarketAccountingSnapshot) models.MarketAccountingSnapshot {
	source := snapshot.Source
	if source == "" {
		source = "read_model"
	}
	return models.MarketAccountingSnapshot{
		MarketID:           snapshot.MarketID,
		LastProbability:    snapshot.LastProbability,
		NetBetVolume:       snapshot.NetBetVolume,
		MarketDust:         snapshot.MarketDust,
		VolumeWithDust:     snapshot.VolumeWithDust,
		UserCount:          snapshot.UserCount,
		BetCount:           snapshot.BetCount,
		LastProcessedBetID: snapshot.LastProcessedBetID,
		LastProcessedBetAt: snapshot.LastProcessedBetAt,
		GeneratedAt:        snapshot.GeneratedAt,
		Source:             source,
		IsStale:            snapshot.IsStale,
		StaleReason:        snapshot.StaleReason,
		MarkedStaleAt:      timeFromPointer(snapshot.MarkedStaleAt),
	}
}

func modelAccountingSnapshotToDomain(snapshot *models.MarketAccountingSnapshot) *dmarkets.MarketAccountingSnapshot {
	if snapshot == nil {
		return nil
	}
	return &dmarkets.MarketAccountingSnapshot{
		MarketID:            snapshot.MarketID,
		GeneratedAt:         snapshot.GeneratedAt,
		LastProbability:     snapshot.LastProbability,
		NetBetVolume:        snapshot.NetBetVolume,
		MarketDust:          snapshot.MarketDust,
		VolumeWithDust:      snapshot.VolumeWithDust,
		UserCount:           snapshot.UserCount,
		BetCount:            snapshot.BetCount,
		LastProcessedBetID:  snapshot.LastProcessedBetID,
		LastProcessedBetAt:  snapshot.LastProcessedBetAt,
		Source:              snapshot.Source,
		TransactionSafeRead: false,
		IsStale:             snapshot.IsStale,
		StaleReason:         snapshot.StaleReason,
		MarkedStaleAt:       timePointerFromValue(snapshot.MarkedStaleAt),
	}
}

func timeFromPointer(value *time.Time) time.Time {
	if value == nil {
		return time.Time{}
	}
	return *value
}

func timePointerFromValue(value time.Time) *time.Time {
	if value.IsZero() {
		return nil
	}
	return &value
}
