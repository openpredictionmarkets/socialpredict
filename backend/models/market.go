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

type MarketTag struct {
	gorm.Model
	ID          int64  `json:"id" gorm:"primary_key"`
	Slug        string `json:"slug" gorm:"not null;uniqueIndex;size:64"`
	DisplayName string `json:"displayName" gorm:"not null;size:120"`
	Description string `json:"description,omitempty" gorm:"type:text"`
	ColorKey    string `json:"colorKey,omitempty" gorm:"size:40"`
	SortOrder   int    `json:"sortOrder" gorm:"not null;default:0;index"`
	IsActive    bool   `json:"isActive" gorm:"not null;default:true;index"`
	CreatedBy   string `json:"createdBy,omitempty" gorm:"index"`
}

type MarketTagAssignment struct {
	gorm.Model
	ID         int64  `json:"id" gorm:"primary_key"`
	MarketID   int64  `json:"marketId" gorm:"not null;uniqueIndex:uniq_market_tag_assignment;index"`
	TagID      int64  `json:"tagId" gorm:"not null;uniqueIndex:uniq_market_tag_assignment;index"`
	AssignedBy string `json:"assignedBy,omitempty" gorm:"index"`
	Source     string `json:"source,omitempty" gorm:"not null;default:moderator_create;index"`
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

type MarketDescriptionAmendment struct {
	gorm.Model
	ID              int64      `json:"id" gorm:"primary_key"`
	MarketID        int64      `json:"marketId" gorm:"not null;uniqueIndex:uniq_market_description_amendment_version;index:idx_market_description_amendments_market_status_version"`
	Version         int        `json:"version" gorm:"not null;uniqueIndex:uniq_market_description_amendment_version;index:idx_market_description_amendments_market_status_version"`
	Body            string     `json:"body" gorm:"type:text;not null"`
	BodyFormat      string     `json:"bodyFormat" gorm:"not null;default:markdown_lite;size:32"`
	Status          string     `json:"status" gorm:"not null;default:pending;index:idx_market_description_amendments_market_status_version;index:idx_market_description_amendments_status_created"`
	CreatedBy       string     `json:"createdBy" gorm:"not null;index"`
	ApprovedBy      string     `json:"approvedBy,omitempty" gorm:"index"`
	ApprovedAt      *time.Time `json:"approvedAt,omitempty"`
	RejectedBy      string     `json:"rejectedBy,omitempty" gorm:"index"`
	RejectedAt      *time.Time `json:"rejectedAt,omitempty"`
	RejectionReason string     `json:"rejectionReason,omitempty" gorm:"type:text"`
	SubmitReason    string     `json:"submitReason,omitempty" gorm:"type:text"`
}

type MarketGovernanceSettings struct {
	gorm.Model
	ID                               uint   `json:"id" gorm:"primaryKey"`
	AutoApproveDescriptionAmendments bool   `json:"autoApproveDescriptionAmendments" gorm:"not null;default:false"`
	AutoApproveMarketProposals       bool   `json:"autoApproveMarketProposals" gorm:"not null;default:false"`
	Version                          uint   `json:"version" gorm:"not null;default:1"`
	UpdatedBy                        string `json:"updatedBy,omitempty" gorm:"size:64"`
}
