package models

import (
	"time"

	"gorm.io/gorm"
)

type Poll struct {
	gorm.Model
	ID              uint       `json:"id" gorm:"primary_key"`
	CreatorUsername string     `json:"creatorUsername" gorm:"not null;index"`
	Question        string     `json:"question" gorm:"not null"`
	Description     string     `json:"description" gorm:"type:text"`
	IsClosed        bool       `json:"isClosed" gorm:"default:false"`
	ClosedAt        *time.Time `json:"closedAt"`
}

type PollVote struct {
	gorm.Model
	ID       uint   `json:"id" gorm:"primary_key"`
	PollID   uint   `json:"pollId" gorm:"not null;index;uniqueIndex:idx_poll_user_vote"`
	Username string `json:"username" gorm:"not null;index;uniqueIndex:idx_poll_user_vote"`
	Vote     string `json:"vote" gorm:"not null"` // "YES" or "NO"
}
