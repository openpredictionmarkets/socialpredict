package users

type balanceAdjustment func(balance int64, amount int64) int64

// Transaction types supported when adjusting user balances.
const (
	TransactionWin    = "WIN"
	TransactionRefund = "REFUND"
	TransactionSale   = "SALE"
	TransactionBuy    = "BUY"
	TransactionFee    = "FEE"
)

var transactionBalanceAdjustments = map[string]balanceAdjustment{
	TransactionWin:    creditBalance,
	TransactionRefund: creditBalance,
	TransactionSale:   creditBalance,
	TransactionBuy:    debitBalance,
	TransactionFee:    debitBalance,
}

func creditBalance(balance int64, amount int64) int64 {
	return balance + amount
}

func debitBalance(balance int64, amount int64) int64 {
	return balance - amount
}

func applyTransactionBalance(balance int64, amount int64, transactionType string) (int64, error) {
	adjust, ok := transactionBalanceAdjustments[transactionType]
	if !ok {
		return 0, ErrInvalidTransactionType
	}
	return adjust(balance, amount), nil
}
