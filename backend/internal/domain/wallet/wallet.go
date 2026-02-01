package wallet

import "time"

// Account holds balance state for a user.
type Account struct {
	ID      int64
	UserID  int64
	Balance int64
}

// LedgerEntry captures a single balance-affecting operation for auditing.
type LedgerEntry struct {
	ID        int64
	Username  string
	Amount    int64
	Type      string // e.g. WIN, REFUND, SALE, BUY, FEE
	Balance   int64  // balance after this transaction
	CreatedAt time.Time
}

// Transaction types for balance operations.
const (
	TxWin    = "WIN"
	TxRefund = "REFUND"
	TxSale   = "SALE"
	TxBuy    = "BUY"
	TxFee    = "FEE"
)

// IsCreditType returns true if the transaction type adds funds.
func IsCreditType(txType string) bool {
	switch txType {
	case TxWin, TxRefund, TxSale:
		return true
	default:
		return false
	}
}

// IsDebitType returns true if the transaction type removes funds.
func IsDebitType(txType string) bool {
	switch txType {
	case TxBuy, TxFee:
		return true
	default:
		return false
	}
}
