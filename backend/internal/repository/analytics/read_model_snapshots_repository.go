package analytics

import (
	"context"
	"errors"
	"time"

	"socialpredict/models"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// GetAnalyticsReadModelSnapshot returns a display-only aggregate analytics
// snapshot. A missing snapshot is not an error.
func (r *GormRepository) GetAnalyticsReadModelSnapshot(ctx context.Context, key string) (*AnalyticsReadModelSnapshot, error) {
	if key == "" {
		return nil, errors.New("snapshot key is required")
	}
	db, err := r.dbWithContext(ctx)
	if err != nil {
		return nil, err
	}

	var snapshot models.AnalyticsReadModelSnapshot
	if err := db.Where("snapshot_key = ?", key).First(&snapshot).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return modelAnalyticsReadModelSnapshotToDomain(&snapshot), nil
}

// UpsertAnalyticsReadModelSnapshot stores a display-only aggregate analytics
// snapshot. Transaction paths must continue using canonical state.
func (r *GormRepository) UpsertAnalyticsReadModelSnapshot(ctx context.Context, snapshot AnalyticsReadModelSnapshot) error {
	if snapshot.Key == "" {
		return errors.New("snapshot key is required")
	}
	db, err := r.dbWithContext(ctx)
	if err != nil {
		return err
	}

	dbSnapshot := domainAnalyticsReadModelSnapshotToModel(snapshot)
	return db.Clauses(clause.OnConflict{
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
	}).Create(&dbSnapshot).Error
}

// MarkAnalyticsReadModelSnapshotStale marks an aggregate display snapshot stale
// after a canonical mutation. Missing snapshots are ignored.
func (r *GormRepository) MarkAnalyticsReadModelSnapshotStale(ctx context.Context, key string, reason string) error {
	if key == "" {
		return errors.New("snapshot key is required")
	}
	db, err := r.dbWithContext(ctx)
	if err != nil {
		return err
	}
	now := time.Now().UTC()
	return db.Model(&models.AnalyticsReadModelSnapshot{}).
		Where("snapshot_key = ?", key).
		Updates(map[string]interface{}{
			"is_stale":        true,
			"stale_reason":    reason,
			"marked_stale_at": now,
			"updated_at":      now,
		}).Error
}

func domainAnalyticsReadModelSnapshotToModel(snapshot AnalyticsReadModelSnapshot) models.AnalyticsReadModelSnapshot {
	source := snapshot.Source
	if source == "" {
		source = "read_model"
	}
	return models.AnalyticsReadModelSnapshot{
		SnapshotKey:   snapshot.Key,
		Kind:          snapshot.Kind,
		PayloadJSON:   string(snapshot.PayloadJSON),
		GeneratedAt:   snapshot.GeneratedAt,
		Source:        source,
		IsStale:       snapshot.IsStale,
		StaleReason:   snapshot.StaleReason,
		MarkedStaleAt: analyticsTimeFromPointer(snapshot.MarkedStaleAt),
	}
}

func modelAnalyticsReadModelSnapshotToDomain(snapshot *models.AnalyticsReadModelSnapshot) *AnalyticsReadModelSnapshot {
	if snapshot == nil {
		return nil
	}
	return &AnalyticsReadModelSnapshot{
		Key:                 snapshot.SnapshotKey,
		Kind:                snapshot.Kind,
		PayloadJSON:         []byte(snapshot.PayloadJSON),
		GeneratedAt:         snapshot.GeneratedAt,
		Source:              snapshot.Source,
		TransactionSafeRead: false,
		IsStale:             snapshot.IsStale,
		StaleReason:         snapshot.StaleReason,
		MarkedStaleAt:       analyticsTimePointerFromValue(snapshot.MarkedStaleAt),
	}
}

func analyticsTimeFromPointer(value *time.Time) time.Time {
	if value == nil {
		return time.Time{}
	}
	return *value
}

func analyticsTimePointerFromValue(value time.Time) *time.Time {
	if value.IsZero() {
		return nil
	}
	return &value
}

var _ AnalyticsReadModelSnapshotRepository = (*GormRepository)(nil)
