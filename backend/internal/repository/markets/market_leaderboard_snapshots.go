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

const marketLeaderboardSnapshotKind = "market_leaderboard"

// GetMarketLeaderboardSnapshot returns a display-only market leaderboard
// snapshot. Missing snapshots are not errors.
func (r *GormRepository) GetMarketLeaderboardSnapshot(ctx context.Context, marketID int64) (*dmarkets.MarketLeaderboardSnapshot, error) {
	if marketID <= 0 {
		return nil, dmarkets.ErrInvalidInput
	}
	var snapshot models.AnalyticsReadModelSnapshot
	if err := r.db.WithContext(ctx).
		Where("snapshot_key = ?", marketLeaderboardSnapshotStorageKey(marketID)).
		First(&snapshot).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return modelMarketLeaderboardSnapshotToDomain(marketID, &snapshot)
}

// UpsertMarketLeaderboardSnapshot stores display-only market leaderboard rows.
func (r *GormRepository) UpsertMarketLeaderboardSnapshot(ctx context.Context, snapshot dmarkets.MarketLeaderboardSnapshot) error {
	if snapshot.MarketID <= 0 {
		return dmarkets.ErrInvalidInput
	}
	dbSnapshot, err := domainMarketLeaderboardSnapshotToModel(snapshot)
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

// MarkMarketLeaderboardSnapshotStale marks an existing leaderboard display
// snapshot stale after canonical market activity. Missing snapshots are ignored.
func (r *GormRepository) MarkMarketLeaderboardSnapshotStale(ctx context.Context, marketID int64, reason string) error {
	if marketID <= 0 {
		return dmarkets.ErrInvalidInput
	}
	now := time.Now().UTC()
	return r.db.WithContext(ctx).
		Model(&models.AnalyticsReadModelSnapshot{}).
		Where("snapshot_key = ?", marketLeaderboardSnapshotStorageKey(marketID)).
		Updates(map[string]interface{}{
			"is_stale":        true,
			"stale_reason":    reason,
			"marked_stale_at": now,
			"updated_at":      now,
		}).Error
}

func marketLeaderboardSnapshotStorageKey(marketID int64) string {
	return fmt.Sprintf("market_leaderboard:%d", marketID)
}

func domainMarketLeaderboardSnapshotToModel(snapshot dmarkets.MarketLeaderboardSnapshot) (models.AnalyticsReadModelSnapshot, error) {
	payload, err := json.Marshal(snapshot.Rows)
	if err != nil {
		return models.AnalyticsReadModelSnapshot{}, err
	}
	source := snapshot.Source
	if source == "" {
		source = "read_model"
	}
	return models.AnalyticsReadModelSnapshot{
		SnapshotKey:   marketLeaderboardSnapshotStorageKey(snapshot.MarketID),
		Kind:          marketLeaderboardSnapshotKind,
		PayloadJSON:   string(payload),
		GeneratedAt:   snapshot.GeneratedAt,
		Source:        source,
		IsStale:       snapshot.IsStale,
		StaleReason:   snapshot.StaleReason,
		MarkedStaleAt: timeFromPointer(snapshot.MarkedStaleAt),
	}, nil
}

func modelMarketLeaderboardSnapshotToDomain(marketID int64, snapshot *models.AnalyticsReadModelSnapshot) (*dmarkets.MarketLeaderboardSnapshot, error) {
	if snapshot == nil {
		return nil, nil
	}
	rows := []*dmarkets.LeaderboardRow{}
	if snapshot.PayloadJSON != "" {
		if err := json.Unmarshal([]byte(snapshot.PayloadJSON), &rows); err != nil {
			return nil, err
		}
	}
	return &dmarkets.MarketLeaderboardSnapshot{
		MarketID:            marketID,
		Rows:                rows,
		GeneratedAt:         snapshot.GeneratedAt,
		Source:              snapshot.Source,
		TransactionSafeRead: false,
		IsStale:             snapshot.IsStale,
		StaleReason:         snapshot.StaleReason,
		MarkedStaleAt:       timePointerFromValue(snapshot.MarkedStaleAt),
	}, nil
}

var _ dmarkets.MarketLeaderboardSnapshotRepository = (*GormRepository)(nil)
