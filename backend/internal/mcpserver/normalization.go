package mcpserver

import (
	"strconv"
	"strings"

	dmarkets "socialpredict/internal/domain/markets"
)

const (
	defaultLimit = 20
	maxLimit     = 100
)

type StatusFilter struct {
	Canonical string
	Filter    string
}

func NormalizeStatus(raw string) (StatusFilter, error) {
	value := strings.ToLower(strings.TrimSpace(raw))
	switch value {
	case "", "all":
		return StatusFilter{Canonical: "all", Filter: ""}, nil
	case "open", "active":
		return StatusFilter{Canonical: dmarkets.MarketStatusActive, Filter: dmarkets.MarketStatusActive}, nil
	case "closed":
		return StatusFilter{Canonical: dmarkets.MarketStatusClosed, Filter: dmarkets.MarketStatusClosed}, nil
	case "resolved":
		return StatusFilter{Canonical: dmarkets.MarketStatusResolved, Filter: dmarkets.MarketStatusResolved}, nil
	default:
		return StatusFilter{}, &ToolError{Code: "validation_error", Message: "status must be active, open, closed, resolved, or all"}
	}
}

func NormalizeTagSlug(raw string) (string, error) {
	slug := dmarkets.NormalizeMarketTagSlug(raw)
	if slug == "" {
		return "", nil
	}
	if _, err := dmarkets.NormalizeMarketTagSlugs([]string{slug}); err != nil {
		return "", &ToolError{Code: "validation_error", Message: "tagSlug is invalid"}
	}
	return slug, nil
}

func NormalizeID(id int64, name string) (int64, error) {
	if id <= 0 {
		return 0, &ToolError{Code: "validation_error", Message: name + " must be a positive integer"}
	}
	return id, nil
}

func NormalizeOutcome(raw string) (string, error) {
	switch strings.ToUpper(strings.TrimSpace(raw)) {
	case "YES":
		return "YES", nil
	case "NO":
		return "NO", nil
	default:
		return "", &ToolError{Code: "validation_error", Message: "outcome must be YES or NO"}
	}
}

func NormalizePage(limit int, offset int) dmarkets.Page {
	if limit <= 0 {
		limit = defaultLimit
	}
	if limit > maxLimit {
		limit = maxLimit
	}
	if offset < 0 {
		offset = 0
	}
	return dmarkets.Page{Limit: limit, Offset: offset}
}

func PositiveInt64(raw string, name string) (int64, error) {
	id, err := strconv.ParseInt(strings.TrimSpace(raw), 10, 64)
	if err != nil {
		return 0, &ToolError{Code: "validation_error", Message: name + " must be a positive integer"}
	}
	return NormalizeID(id, name)
}
