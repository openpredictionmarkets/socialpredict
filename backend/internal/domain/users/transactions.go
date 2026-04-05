package users

type balanceAdjustment interface {
	Apply(balance int64, amount int64) int64
}

type balanceAdjustmentFunc func(balance int64, amount int64) int64

func (f balanceAdjustmentFunc) Apply(balance int64, amount int64) int64 {
	return f(balance, amount)
}

type TransactionType = string

// Transaction types supported when adjusting user balances.
const (
	TransactionWin    TransactionType = "WIN"
	TransactionRefund TransactionType = "REFUND"
	TransactionSale   TransactionType = "SALE"
	TransactionBuy    TransactionType = "BUY"
	TransactionFee    TransactionType = "FEE"
)

var transactionBalanceAdjustments = map[TransactionType]balanceAdjustment{
	TransactionWin:    balanceAdjustmentFunc(creditBalance),
	TransactionRefund: balanceAdjustmentFunc(creditBalance),
	TransactionSale:   balanceAdjustmentFunc(creditBalance),
	TransactionBuy:    balanceAdjustmentFunc(debitBalance),
	TransactionFee:    balanceAdjustmentFunc(debitBalance),
}

func creditBalance(balance int64, amount int64) int64 {
	return balance + amount
}

func debitBalance(balance int64, amount int64) int64 {
	return balance - amount
}

func applyTransactionBalance(balance int64, amount int64, transactionType string) (int64, error) {
	adjust, ok := transactionBalanceAdjustments[TransactionType(transactionType)]
	if !ok {
		return 0, ErrInvalidTransactionType
	}
	return adjust.Apply(balance, amount), nil
}
