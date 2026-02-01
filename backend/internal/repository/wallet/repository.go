package wallet

import (
	"context"
	"errors"

	dwallet "socialpredict/internal/domain/wallet"
	"socialpredict/models"

	"gorm.io/gorm"
)

// GormRepository implements the wallet domain repository interface using GORM.
type GormRepository struct {
	db *gorm.DB
}

// NewGormRepository creates a new GORM-based wallet repository.
func NewGormRepository(db *gorm.DB) *GormRepository {
	return &GormRepository{db: db}
}

// GetBalance retrieves the current balance for a user.
func (r *GormRepository) GetBalance(ctx context.Context, username string) (int64, error) {
	var balance int64
	err := r.db.WithContext(ctx).
		Model(&models.User{}).
		Select("account_balance").
		Where("username = ?", username).
		Take(&balance).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, dwallet.ErrAccountNotFound
		}
		return 0, err
	}

	return balance, nil
}

// UpdateBalance sets a new balance for a user.
func (r *GormRepository) UpdateBalance(ctx context.Context, username string, newBalance int64) error {
	result := r.db.WithContext(ctx).
		Model(&models.User{}).
		Where("username = ?", username).
		Update("account_balance", newBalance)

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return dwallet.ErrAccountNotFound
	}

	return nil
}

// RecordTransaction persists a ledger entry for audit purposes.
func (r *GormRepository) RecordTransaction(ctx context.Context, entry *dwallet.LedgerEntry) error {
	dbEntry := domainToModel(entry)
	return r.db.WithContext(ctx).Create(&dbEntry).Error
}

// UpdateBalanceAndRecord atomically updates the balance and records the transaction.
func (r *GormRepository) UpdateBalanceAndRecord(ctx context.Context, username string, newBalance int64, entry *dwallet.LedgerEntry) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Update balance
		result := tx.Model(&models.User{}).
			Where("username = ?", username).
			Update("account_balance", newBalance)

		if result.Error != nil {
			return result.Error
		}

		if result.RowsAffected == 0 {
			return dwallet.ErrAccountNotFound
		}

		// Record ledger entry
		dbEntry := domainToModel(entry)
		if err := tx.Create(&dbEntry).Error; err != nil {
			return err
		}

		// Update entry ID from database
		entry.ID = dbEntry.ID

		return nil
	})
}

// domainToModel converts a domain LedgerEntry to a GORM model.
func domainToModel(entry *dwallet.LedgerEntry) models.LedgerEntry {
	return models.LedgerEntry{
		Username:  entry.Username,
		Amount:    entry.Amount,
		Type:      entry.Type,
		Balance:   entry.Balance,
		CreatedAt: entry.CreatedAt,
	}
}

// modelToDomain converts a GORM model to a domain LedgerEntry.
func modelToDomain(dbEntry *models.LedgerEntry) *dwallet.LedgerEntry {
	return &dwallet.LedgerEntry{
		ID:        dbEntry.ID,
		Username:  dbEntry.Username,
		Amount:    dbEntry.Amount,
		Type:      dbEntry.Type,
		Balance:   dbEntry.Balance,
		CreatedAt: dbEntry.CreatedAt,
	}
}

// Ensure GormRepository implements the wallet.Repository interface.
var _ dwallet.Repository = (*GormRepository)(nil)
