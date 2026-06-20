package analytics

import (
	"context"
	"errors"
	"time"

	"socialpredict/models"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// GetUserFinancialMetricSnapshot returns an authenticated display/read-model
// financial snapshot for username. A missing snapshot is not an error.
func (r *GormRepository) GetUserFinancialMetricSnapshot(ctx context.Context, username string) (*UserFinancialMetricSnapshot, error) {
	if username == "" {
		return nil, errors.New("username is required")
	}
	db, err := r.dbWithContext(ctx)
	if err != nil {
		return nil, err
	}

	var snapshot models.UserFinancialMetricSnapshot
	if err := db.Where("username = ?", username).First(&snapshot).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return modelUserFinancialMetricSnapshotToDomain(&snapshot), nil
}

// UpsertUserFinancialMetricSnapshot stores authenticated display/read-model
// financial metrics. Transaction paths must continue using canonical user,
// market, and bet state instead of these snapshots.
func (r *GormRepository) UpsertUserFinancialMetricSnapshot(ctx context.Context, snapshot UserFinancialMetricSnapshot) error {
	if snapshot.Username == "" {
		return errors.New("username is required")
	}
	db, err := r.dbWithContext(ctx)
	if err != nil {
		return err
	}

	dbSnapshot := domainUserFinancialMetricSnapshotToModel(snapshot)
	return db.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "username"}},
		DoUpdates: clause.AssignmentColumns([]string{
			"account_balance",
			"maximum_debt_allowed",
			"amount_in_play",
			"amount_borrowed",
			"retained_earnings",
			"equity",
			"trading_profits",
			"work_profits",
			"unrealized_work_income",
			"unrealized_work_profits",
			"total_profits",
			"amount_in_play_active",
			"total_spent",
			"total_spent_in_play",
			"realized_profits",
			"potential_profits",
			"realized_value",
			"potential_value",
			"position_count",
			"generated_at",
			"source",
			"is_stale",
			"stale_reason",
			"marked_stale_at",
			"updated_at",
		}),
	}).Create(&dbSnapshot).Error
}

// MarkUserFinancialMetricSnapshotStale marks an existing authenticated display
// snapshot stale after a canonical user-affecting mutation. Missing snapshots
// are ignored because refresh can recreate them later.
func (r *GormRepository) MarkUserFinancialMetricSnapshotStale(ctx context.Context, username string, reason string) error {
	if username == "" {
		return errors.New("username is required")
	}
	db, err := r.dbWithContext(ctx)
	if err != nil {
		return err
	}
	now := time.Now().UTC()
	return db.Model(&models.UserFinancialMetricSnapshot{}).
		Where("username = ?", username).
		Updates(map[string]interface{}{
			"is_stale":        true,
			"stale_reason":    reason,
			"marked_stale_at": now,
			"updated_at":      now,
		}).Error
}

func domainUserFinancialMetricSnapshotToModel(snapshot UserFinancialMetricSnapshot) models.UserFinancialMetricSnapshot {
	source := snapshot.Source
	if source == "" {
		source = "read_model"
	}
	financial := snapshot.Financial
	return models.UserFinancialMetricSnapshot{
		Username:              snapshot.Username,
		AccountBalance:        financial.AccountBalance,
		MaximumDebtAllowed:    financial.MaximumDebtAllowed,
		AmountInPlay:          financial.AmountInPlay,
		AmountBorrowed:        financial.AmountBorrowed,
		RetainedEarnings:      financial.RetainedEarnings,
		Equity:                financial.Equity,
		TradingProfits:        financial.TradingProfits,
		WorkProfits:           financial.WorkProfits,
		UnrealizedWorkIncome:  financial.UnrealizedWorkIncome,
		UnrealizedWorkProfits: financial.UnrealizedWorkProfits,
		TotalProfits:          financial.TotalProfits,
		AmountInPlayActive:    financial.AmountInPlayActive,
		TotalSpent:            financial.TotalSpent,
		TotalSpentInPlay:      financial.TotalSpentInPlay,
		RealizedProfits:       financial.RealizedProfits,
		PotentialProfits:      financial.PotentialProfits,
		RealizedValue:         financial.RealizedValue,
		PotentialValue:        financial.PotentialValue,
		PositionCount:         snapshot.PositionCount,
		GeneratedAt:           snapshot.GeneratedAt,
		Source:                source,
		IsStale:               snapshot.IsStale,
		StaleReason:           snapshot.StaleReason,
		MarkedStaleAt:         userFinancialTimeFromPointer(snapshot.MarkedStaleAt),
	}
}

func modelUserFinancialMetricSnapshotToDomain(snapshot *models.UserFinancialMetricSnapshot) *UserFinancialMetricSnapshot {
	if snapshot == nil {
		return nil
	}
	return &UserFinancialMetricSnapshot{
		Username:      snapshot.Username,
		GeneratedAt:   snapshot.GeneratedAt,
		PositionCount: snapshot.PositionCount,
		Financial: FinancialSnapshot{
			AccountBalance:        snapshot.AccountBalance,
			MaximumDebtAllowed:    snapshot.MaximumDebtAllowed,
			AmountInPlay:          snapshot.AmountInPlay,
			AmountBorrowed:        snapshot.AmountBorrowed,
			RetainedEarnings:      snapshot.RetainedEarnings,
			Equity:                snapshot.Equity,
			TradingProfits:        snapshot.TradingProfits,
			WorkProfits:           snapshot.WorkProfits,
			UnrealizedWorkIncome:  snapshot.UnrealizedWorkIncome,
			UnrealizedWorkProfits: snapshot.UnrealizedWorkProfits,
			TotalProfits:          snapshot.TotalProfits,
			AmountInPlayActive:    snapshot.AmountInPlayActive,
			TotalSpent:            snapshot.TotalSpent,
			TotalSpentInPlay:      snapshot.TotalSpentInPlay,
			RealizedProfits:       snapshot.RealizedProfits,
			PotentialProfits:      snapshot.PotentialProfits,
			RealizedValue:         snapshot.RealizedValue,
			PotentialValue:        snapshot.PotentialValue,
		},
		Source:              snapshot.Source,
		TransactionSafeRead: false,
		IsStale:             snapshot.IsStale,
		StaleReason:         snapshot.StaleReason,
		MarkedStaleAt:       userFinancialTimePointerFromValue(snapshot.MarkedStaleAt),
	}
}

func userFinancialTimeFromPointer(value *time.Time) time.Time {
	if value == nil {
		return time.Time{}
	}
	return *value
}

func userFinancialTimePointerFromValue(value time.Time) *time.Time {
	if value.IsZero() {
		return nil
	}
	return &value
}
