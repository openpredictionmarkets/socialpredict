package models

import (
	"time"

	"gorm.io/gorm"
)

type Market struct {
	gorm.Model
	ID                      int64     `json:"id" gorm:"primary_key"`
	QuestionTitle           string    `json:"questionTitle" gorm:"not null"`
	Description             string    `json:"description" gorm:"not null"`
	OutcomeType             string    `json:"outcomeType" gorm:"not null"`
	ResolutionDateTime      time.Time `json:"resolutionDateTime" gorm:"not null"`
	FinalResolutionDateTime time.Time `json:"finalResolutionDateTime"`
	UTCOffset               int       `json:"utcOffset"`
	IsResolved              bool      `json:"isResolved"`
	ResolutionResult        string    `json:"resolutionResult"`
	InitialProbability      float64   `json:"initialProbability" gorm:"not null"`
	CreatorUsername         string    `json:"creatorUsername" gorm:"not null"`
	Creator                 User      `gorm:"foreignKey:CreatorUsername;references:Username"`
}

func (m *Market) IsClosed() bool {
	if !m.IsResolved {
		return false
	}

	// TODO: Decide if we should be using one or both of the below conditions
	if time.Now().Before(m.ResolutionDateTime) {
		return false
	}

	if time.Now().Before(m.FinalResolutionDateTime) {
		return false
	}

	return true

}
