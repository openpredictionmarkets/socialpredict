package dto

import (
	"time"
)

// CreateMarketRequest represents the HTTP request body for creating a market
type CreateMarketRequest struct {
	QuestionTitle      string    `json:"questionTitle" validate:"required,max=160"`
	Description        string    `json:"description" validate:"max=2000"`
	OutcomeType        string    `json:"outcomeType" validate:"required"`
	ResolutionDateTime time.Time `json:"resolutionDateTime" validate:"required"`
	YesLabel           string    `json:"yesLabel" validate:"omitempty,max=20"`
	NoLabel            string    `json:"noLabel" validate:"omitempty,max=20"`
}

// UpdateLabelsRequest represents the HTTP request body for updating market labels
type UpdateLabelsRequest struct {
	YesLabel string `json:"yesLabel" validate:"required,min=1,max=20"`
	NoLabel  string `json:"noLabel" validate:"required,min=1,max=20"`
}

// ListMarketsQueryParams represents query parameters for listing markets
type ListMarketsQueryParams struct {
	Status    string `form:"status"`
	CreatedBy string `form:"created_by"`
	Limit     int    `form:"limit"`
	Offset    int    `form:"offset"`
}

// SearchMarketsQueryParams represents query parameters for searching markets
type SearchMarketsQueryParams struct {
	Query  string `form:"q" validate:"required"`
	Status string `form:"status"`
	Limit  int    `form:"limit"`
	Offset int    `form:"offset"`
}

// ResolveMarketRequest represents the HTTP request body for resolving a market
type ResolveMarketRequest struct {
	Resolution string `json:"resolution" validate:"required,oneof=yes no"`
}
