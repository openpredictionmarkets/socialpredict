package models

import (
	"time"

	"gorm.io/gorm"
)

// UserFinancialMetricSnapshot stores authenticated display/read-model user
// financial metrics. It is not transaction truth for balances or settlement.
type UserFinancialMetricSnapshot struct {
	gorm.Model
	ID                 uint      `json:"id" gorm:"primaryKey"`
	Username           string    `json:"username" gorm:"not null;uniqueIndex;size:64"`
	AccountBalance     int64     `json:"accountBalance" gorm:"not null;default:0"`
	MaximumDebtAllowed int64     `json:"maximumDebtAllowed" gorm:"not null;default:0"`
	AmountInPlay       int64     `json:"amountInPlay" gorm:"not null;default:0"`
	AmountBorrowed     int64     `json:"amountBorrowed" gorm:"not null;default:0"`
	RetainedEarnings   int64     `json:"retainedEarnings" gorm:"not null;default:0"`
	Equity             int64     `json:"equity" gorm:"not null;default:0"`
	TradingProfits     int64     `json:"tradingProfits" gorm:"not null;default:0"`
	WorkProfits        int64     `json:"workProfits" gorm:"not null;default:0"`
	TotalProfits       int64     `json:"totalProfits" gorm:"not null;default:0"`
	AmountInPlayActive int64     `json:"amountInPlayActive" gorm:"not null;default:0"`
	TotalSpent         int64     `json:"totalSpent" gorm:"not null;default:0"`
	TotalSpentInPlay   int64     `json:"totalSpentInPlay" gorm:"not null;default:0"`
	RealizedProfits    int64     `json:"realizedProfits" gorm:"not null;default:0"`
	PotentialProfits   int64     `json:"potentialProfits" gorm:"not null;default:0"`
	RealizedValue      int64     `json:"realizedValue" gorm:"not null;default:0"`
	PotentialValue     int64     `json:"potentialValue" gorm:"not null;default:0"`
	PositionCount      int       `json:"positionCount" gorm:"not null;default:0"`
	GeneratedAt        time.Time `json:"generatedAt" gorm:"not null;index"`
	Source             string    `json:"source" gorm:"not null;default:read_model;size:32"`
}
