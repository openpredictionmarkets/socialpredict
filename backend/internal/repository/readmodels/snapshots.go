package readmodels

import (
	"context"
	"errors"
	"strings"
	"time"

	"socialpredict/models"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// Snapshot stores a display-only read-model payload. PayloadJSON is owned by
// the API/read-model boundary and must never be used for transaction decisions.
type Snapshot struct {
	Key           string
	Kind          string
	PayloadJSON   string
	GeneratedAt   time.Time
	Source        string
	IsStale       bool
	StaleReason   string
	MarkedStaleAt *time.Time
}

type GormRepository struct {
	db *gorm.DB
}

func NewGormRepository(db *gorm.DB) *GormRepository {
	return &GormRepository{db: db}
}

func (r *GormRepository) Get(ctx context.Context, key string) (*Snapshot, error) {
	if r == nil || r.db == nil || strings.TrimSpace(key) == "" {
		return nil, nil
	}
	var model models.AnalyticsReadModelSnapshot
	if err := r.db.WithContext(ctx).Where("snapshot_key = ?", key).First(&model).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return snapshotFromModel(&model), nil
}

func (r *GormRepository) Upsert(ctx context.Context, snapshot Snapshot) error {
	if r == nil || r.db == nil || strings.TrimSpace(snapshot.Key) == "" {
		return nil
	}
	model := modelFromSnapshot(snapshot)
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
		Create(&model).Error
}

// MarkMarketDiscoverySnapshotsStale marks all discovery/card snapshots stale.
// Discovery snapshots are page-level, so broad invalidation is safer than
// trying to prove which tag/status page contains a changed market.
func (r *GormRepository) MarkMarketDiscoverySnapshotsStale(ctx context.Context, reason string) error {
	if r == nil || r.db == nil {
		return nil
	}
	if strings.TrimSpace(reason) == "" {
		reason = "discovery_changed"
	}
	now := time.Now().UTC()
	updates := map[string]interface{}{
		"is_stale":        true,
		"stale_reason":    reason,
		"marked_stale_at": now,
		"updated_at":      now,
	}
	if marketDiscoveryStructuralChange(reason) {
		updates["generated_at"] = now.Add(-24 * time.Hour)
	}
	return r.db.WithContext(ctx).
		Model(&models.AnalyticsReadModelSnapshot{}).
		Where("snapshot_key LIKE ?", "market_discovery:%").
		Updates(updates).Error
}

func marketDiscoveryStructuralChange(reason string) bool {
	switch strings.TrimSpace(reason) {
	case "market_created",
		"market_group_created",
		"market_status_changed",
		"market_tags_changed",
		"tag_catalog_changed",
		"cms_page_changed",
		"cms_pins_changed":
		return true
	default:
		return false
	}
}

func snapshotFromModel(model *models.AnalyticsReadModelSnapshot) *Snapshot {
	if model == nil {
		return nil
	}
	var markedAt *time.Time
	if !model.MarkedStaleAt.IsZero() {
		value := model.MarkedStaleAt
		markedAt = &value
	}
	return &Snapshot{
		Key:           model.SnapshotKey,
		Kind:          model.Kind,
		PayloadJSON:   model.PayloadJSON,
		GeneratedAt:   model.GeneratedAt,
		Source:        model.Source,
		IsStale:       model.IsStale,
		StaleReason:   model.StaleReason,
		MarkedStaleAt: markedAt,
	}
}

func modelFromSnapshot(snapshot Snapshot) models.AnalyticsReadModelSnapshot {
	source := snapshot.Source
	if source == "" {
		source = "read_model"
	}
	var markedAt time.Time
	if snapshot.MarkedStaleAt != nil {
		markedAt = *snapshot.MarkedStaleAt
	}
	return models.AnalyticsReadModelSnapshot{
		SnapshotKey:   snapshot.Key,
		Kind:          snapshot.Kind,
		PayloadJSON:   snapshot.PayloadJSON,
		GeneratedAt:   snapshot.GeneratedAt,
		Source:        source,
		IsStale:       snapshot.IsStale,
		StaleReason:   snapshot.StaleReason,
		MarkedStaleAt: markedAt,
	}
}
