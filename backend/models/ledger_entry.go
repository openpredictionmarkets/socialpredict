package models

import "time"

// LedgerEntry records a balance-affecting transaction for audit purposes.
type LedgerEntry struct {
	ID        int64     `gorm:"primaryKey;autoIncrement"`
	Username  string    `gorm:"not null;index"`
	Amount    int64     `gorm:"not null"` // positive for credit, negative for debit
	Type      string    `gorm:"not null"` // WIN, REFUND, SALE, BUY, FEE
	Balance   int64     `gorm:"not null"` // balance after this transaction
	CreatedAt time.Time `gorm:"not null;index"`
}

// TableName specifies the table name for GORM.
func (LedgerEntry) TableName() string {
	return "ledger_entries"
}
