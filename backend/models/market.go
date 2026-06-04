package models

import (
	"time"

	"gorm.io/gorm"
)

type Market struct {
	gorm.Model
	ID                      int64      `json:"id" gorm:"primary_key"`
	QuestionTitle           string     `json:"questionTitle" gorm:"not null"`
	Description             string     `json:"description" gorm:"not null"`
	OutcomeType             string     `json:"outcomeType" gorm:"not null"`
	ResolutionDateTime      time.Time  `json:"resolutionDateTime" gorm:"not null"`
	FinalResolutionDateTime time.Time  `json:"finalResolutionDateTime"`
	UTCOffset               int        `json:"utcOffset"`
	IsResolved              bool       `json:"isResolved"`
	ResolutionResult        string     `json:"resolutionResult"`
	InitialProbability      float64    `json:"initialProbability" gorm:"not null"`
	YesLabel                string     `json:"yesLabel" gorm:"default:YES"`
	NoLabel                 string     `json:"noLabel" gorm:"default:NO"`
	LifecycleStatus         string     `json:"lifecycleStatus" gorm:"not null;default:published;index"`
	ApprovedBy              string     `json:"approvedBy,omitempty" gorm:"index"`
	ApprovedAt              *time.Time `json:"approvedAt,omitempty"`
	RejectedBy              string     `json:"rejectedBy,omitempty" gorm:"index"`
	RejectedAt              *time.Time `json:"rejectedAt,omitempty"`
	RejectionReason         string     `json:"rejectionReason,omitempty" gorm:"type:text"`
	ProposalCost            int64      `json:"proposalCost" gorm:"not null;default:0"`
	CreatorUsername         string     `json:"creatorUsername" gorm:"not null"`
	Creator                 User       `gorm:"foreignKey:CreatorUsername;references:Username"`
	StewardUsername         string     `json:"stewardUsername" gorm:"index"`
}

type MarketStewardshipAudit struct {
	gorm.Model
	ID                  int64  `json:"id" gorm:"primary_key"`
	MarketID            int64  `json:"marketId" gorm:"not null;index"`
	FromStewardUsername string `json:"fromStewardUsername" gorm:"index"`
	ToStewardUsername   string `json:"toStewardUsername" gorm:"not null;index"`
	ActorUsername       string `json:"actorUsername" gorm:"not null;index"`
	Reason              string `json:"reason,omitempty" gorm:"type:text"`
}
