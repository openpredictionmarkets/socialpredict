package markets

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	dmarkets "socialpredict/internal/domain/markets"
	"socialpredict/models"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const marketPositionsSnapshotKind = "market_positions"

// GetMarketPositionsSnapshot returns a display-only market positions snapshot.
// Missing snapshots are not errors.
func (r *GormRepository) GetMarketPositionsSnapshot(ctx context.Context, marketID int64) (*dmarkets.MarketPositionsSnapshot, error) {
	if marketID <= 0 {
		return nil, dmarkets.ErrInvalidInput
	}
	var snapshot models.AnalyticsReadModelSnapshot
	if err := r.db.WithContext(ctx).
		Where("snapshot_key = ?", marketPositionsSnapshotStorageKey(marketID)).
		First(&snapshot).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return modelMarketPositionsSnapshotToDomain(marketID, &snapshot)
}

// UpsertMarketPositionsSnapshot stores display-only market positions rows.
func (r *GormRepository) UpsertMarketPositionsSnapshot(ctx context.Context, snapshot dmarkets.MarketPositionsSnapshot) error {
	if snapshot.MarketID <= 0 {
		return dmarkets.ErrInvalidInput
	}
	dbSnapshot, err := domainMarketPositionsSnapshotToModel(snapshot)
	if err != nil {
		return err
	}
	return r.db.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: "snapshot_key"}},
			DoUpdates: clause.AssignmentColumns([]string{
				"kind",
				"payload_json",
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

// MarkMarketPositionsSnapshotStale marks an existing positions display snapshot
// stale after canonical market activity. Missing snapshots are ignored.
func (r *GormRepository) MarkMarketPositionsSnapshotStale(ctx context.Context, marketID int64, reason string) error {
	if marketID <= 0 {
		return dmarkets.ErrInvalidInput
	}
	now := time.Now().UTC()
	return r.db.WithContext(ctx).
		Model(&models.AnalyticsReadModelSnapshot{}).
		Where("snapshot_key = ?", marketPositionsSnapshotStorageKey(marketID)).
		Updates(map[string]interface{}{
			"is_stale":        true,
			"stale_reason":    reason,
			"marked_stale_at": now,
			"updated_at":      now,
		}).Error
}

func marketPositionsSnapshotStorageKey(marketID int64) string {
	return fmt.Sprintf("market_positions:%d", marketID)
}

func domainMarketPositionsSnapshotToModel(snapshot dmarkets.MarketPositionsSnapshot) (models.AnalyticsReadModelSnapshot, error) {
	payload, err := json.Marshal(snapshot.Positions.Normalize())
	if err != nil {
		return models.AnalyticsReadModelSnapshot{}, err
	}
	source := snapshot.Source
	if source == "" {
		source = "read_model"
	}
	return models.AnalyticsReadModelSnapshot{
		SnapshotKey:   marketPositionsSnapshotStorageKey(snapshot.MarketID),
		Kind:          marketPositionsSnapshotKind,
		PayloadJSON:   string(payload),
		GeneratedAt:   snapshot.GeneratedAt,
		Source:        source,
		IsStale:       snapshot.IsStale,
		StaleReason:   snapshot.StaleReason,
		MarkedStaleAt: timeFromPointer(snapshot.MarkedStaleAt),
	}, nil
}

func modelMarketPositionsSnapshotToDomain(marketID int64, snapshot *models.AnalyticsReadModelSnapshot) (*dmarkets.MarketPositionsSnapshot, error) {
	if snapshot == nil {
		return nil, nil
	}
	positions := dmarkets.MarketPositions{}
	if snapshot.PayloadJSON != "" {
		if err := json.Unmarshal([]byte(snapshot.PayloadJSON), &positions); err != nil {
			return nil, err
		}
	}
	return &dmarkets.MarketPositionsSnapshot{
		MarketID:            marketID,
		Positions:           positions.Normalize(),
		GeneratedAt:         snapshot.GeneratedAt,
		Source:              snapshot.Source,
		TransactionSafeRead: false,
		IsStale:             snapshot.IsStale,
		StaleReason:         snapshot.StaleReason,
		MarkedStaleAt:       timePointerFromValue(snapshot.MarkedStaleAt),
	}, nil
}

var _ dmarkets.MarketPositionsSnapshotRepository = (*GormRepository)(nil)
