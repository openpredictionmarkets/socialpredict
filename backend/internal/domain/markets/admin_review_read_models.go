package markets

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"
)

const (
	adminReviewDefaultLimit  = 50
	adminReviewMaxLimit      = 200
	adminReviewMaxQueryRunes = 100
)

// AdminMarketReviewFilters are the explicit read-model query inputs for the
// admin market review queue. The queue groups child markets before pagination.
type AdminMarketReviewFilters struct {
	Status string
	Query  string
	Limit  int
	Offset int
}

type AdminMarketReviewRow struct {
	RowKey        string
	IsMarketGroup bool
	Market        *Market
	Group         *MarketGroup
	Children      []*Market
	Tags          []MarketTag
}

type AdminMarketReviewPage struct {
	Rows   []AdminMarketReviewRow
	Total  int
	Limit  int
	Offset int
}

// AdminMarketReviewRepository owns grouped admin market queue pagination.
// Implementations must group binary child markets before count/limit/offset are
// applied so frontend queues never need to infer aggregate rows.
type AdminMarketReviewRepository interface {
	ListAdminMarketReviewRows(ctx context.Context, filters AdminMarketReviewFilters) (*AdminMarketReviewPage, error)
}

// AdminDescriptionAmendmentReviewFilters are the explicit read-model query
// inputs for admin amendment review. Grouped market amendments are collapsed
// before search totals and pagination are returned.
type AdminDescriptionAmendmentReviewFilters struct {
	MarketID int64
	Status   string
	Query    string
	Limit    int
	Offset   int
}

type AdminDescriptionAmendmentReviewRow struct {
	RowKey                 string
	IsMarketGroupAmendment bool
	Amendment              MarketDescriptionAmendment
	ChildAmendments        []MarketDescriptionAmendment
}

type AdminDescriptionAmendmentReviewPage struct {
	Rows   []AdminDescriptionAmendmentReviewRow
	Total  int
	Limit  int
	Offset int
}

// AdminAnswerAdditionReviewFilters are the explicit read-model query inputs
// for grouped-market answer additions.
type AdminAnswerAdditionReviewFilters struct {
	GroupID int64
	Status  string
	Query   string
	Limit   int
	Offset  int
}

type AdminAnswerAdditionReviewPage struct {
	Rows   []MarketGroupAnswerAddition
	Total  int
	Limit  int
	Offset int
}

// MarketDescriptionAmendmentAdminReviewRepository supplies status-scoped raw
// amendment candidates for the grouped admin read model. It is intentionally a
// separate review seam so generic amendment list filters do not grow hidden
// unbounded-pagination semantics.
type MarketDescriptionAmendmentAdminReviewRepository interface {
	ListMarketDescriptionAmendmentReviewCandidates(ctx context.Context, filters AdminDescriptionAmendmentReviewFilters) ([]MarketDescriptionAmendment, int, error)
}

// MarketGroupAnswerAdditionAdminReviewRepository owns query/count pagination for
// answer-option review rows.
type MarketGroupAnswerAdditionAdminReviewRepository interface {
	ListMarketGroupAnswerAdditionsForAdminReview(ctx context.Context, filters AdminAnswerAdditionReviewFilters) ([]MarketGroupAnswerAddition, int, error)
}

func (s *Service) ListAdminMarketReviewRows(ctx context.Context, filters AdminMarketReviewFilters) (*AdminMarketReviewPage, error) {
	filters = normalizeAdminMarketReviewFilters(filters)
	if !validAdminMarketReviewStatus(filters.Status) {
		return nil, ErrInvalidInput
	}
	if err := validateAdminReviewQuery(filters.Query); err != nil {
		return nil, err
	}

	repo, ok := s.repo.(AdminMarketReviewRepository)
	if !ok {
		return nil, ErrInvalidInput
	}
	page, err := repo.ListAdminMarketReviewRows(ctx, filters)
	if err != nil {
		return nil, err
	}
	if page == nil {
		page = &AdminMarketReviewPage{}
	}
	page.Limit = filters.Limit
	page.Offset = filters.Offset
	return page, nil
}

func (s *Service) ListAdminMarketDescriptionAmendmentRows(ctx context.Context, filters AdminDescriptionAmendmentReviewFilters) (*AdminDescriptionAmendmentReviewPage, error) {
	filters = normalizeAdminDescriptionAmendmentReviewFilters(filters)
	if !validDescriptionAmendmentReviewStatus(filters.Status) {
		return nil, ErrInvalidInput
	}
	if err := validateAdminReviewQuery(filters.Query); err != nil {
		return nil, err
	}
	repo, ok := s.repo.(MarketDescriptionAmendmentAdminReviewRepository)
	if !ok {
		return nil, ErrInvalidInput
	}
	items, total, err := repo.ListMarketDescriptionAmendmentReviewCandidates(ctx, filters)
	if err != nil {
		return nil, err
	}
	items = s.hydrateDescriptionAmendmentContext(ctx, items)
	rows := groupedAdminDescriptionAmendmentRows(items)
	return &AdminDescriptionAmendmentReviewPage{
		Rows:   rows,
		Total:  total,
		Limit:  filters.Limit,
		Offset: filters.Offset,
	}, nil
}

func (s *Service) ListAdminMarketGroupAnswerAdditionRows(ctx context.Context, filters AdminAnswerAdditionReviewFilters) (*AdminAnswerAdditionReviewPage, error) {
	filters = normalizeAdminAnswerAdditionReviewFilters(filters)
	if !validAnswerAdditionReviewStatus(filters.Status) {
		return nil, ErrInvalidInput
	}
	if err := validateAdminReviewQuery(filters.Query); err != nil {
		return nil, err
	}
	repo, ok := s.repo.(MarketGroupAnswerAdditionAdminReviewRepository)
	if !ok {
		return nil, ErrInvalidInput
	}
	rows, total, err := repo.ListMarketGroupAnswerAdditionsForAdminReview(ctx, filters)
	if err != nil {
		return nil, err
	}
	return &AdminAnswerAdditionReviewPage{
		Rows:   rows,
		Total:  total,
		Limit:  filters.Limit,
		Offset: filters.Offset,
	}, nil
}

func normalizeAdminMarketReviewFilters(filters AdminMarketReviewFilters) AdminMarketReviewFilters {
	filters.Status = NormalizeLifecycleStatus(filters.Status)
	if filters.Status == "" {
		filters.Status = MarketLifecycleProposed
	}
	filters.Query = strings.TrimSpace(filters.Query)
	filters.Limit = normalizeAdminReviewLimit(filters.Limit, adminReviewDefaultLimit, 100)
	filters.Offset = normalizeAdminReviewOffset(filters.Offset)
	return filters
}

func normalizeAdminDescriptionAmendmentReviewFilters(filters AdminDescriptionAmendmentReviewFilters) AdminDescriptionAmendmentReviewFilters {
	filters.Status = NormalizeDescriptionAmendmentStatus(filters.Status)
	filters.Query = strings.TrimSpace(filters.Query)
	filters.Limit = normalizeAdminReviewLimit(filters.Limit, adminReviewDefaultLimit, adminReviewMaxLimit)
	filters.Offset = normalizeAdminReviewOffset(filters.Offset)
	return filters
}

func normalizeAdminAnswerAdditionReviewFilters(filters AdminAnswerAdditionReviewFilters) AdminAnswerAdditionReviewFilters {
	filters.Status = NormalizeMarketGroupAnswerAdditionStatus(filters.Status)
	filters.Query = strings.TrimSpace(filters.Query)
	filters.Limit = normalizeAdminReviewLimit(filters.Limit, adminReviewDefaultLimit, adminReviewMaxLimit)
	filters.Offset = normalizeAdminReviewOffset(filters.Offset)
	return filters
}

func normalizeAdminReviewLimit(limit int, fallback int, max int) int {
	if limit <= 0 {
		return fallback
	}
	if limit > max {
		return max
	}
	return limit
}

func normalizeAdminReviewOffset(offset int) int {
	if offset < 0 {
		return 0
	}
	return offset
}

func validateAdminReviewQuery(query string) error {
	if len([]rune(strings.TrimSpace(query))) > adminReviewMaxQueryRunes {
		return ErrInvalidInput
	}
	return nil
}

func validAdminMarketReviewStatus(status string) bool {
	switch status {
	case MarketStatusAll, MarketLifecycleProposed, MarketLifecyclePublished, MarketLifecycleRejected, MarketLifecycleClosed, MarketLifecycleResolved:
		return true
	default:
		return false
	}
}

func validDescriptionAmendmentReviewStatus(status string) bool {
	switch status {
	case DescriptionAmendmentStatusPending, DescriptionAmendmentStatusApproved, DescriptionAmendmentStatusRejected:
		return true
	default:
		return false
	}
}

func validAnswerAdditionReviewStatus(status string) bool {
	switch status {
	case MarketGroupAnswerAdditionStatusPending, MarketGroupAnswerAdditionStatusApproved, MarketGroupAnswerAdditionStatusRejected:
		return true
	default:
		return false
	}
}

func adminMarketReviewRowsFromDiscovery(rows []MarketDiscoveryRow) []AdminMarketReviewRow {
	out := make([]AdminMarketReviewRow, 0, len(rows))
	for _, row := range rows {
		if row.Group != nil && row.Group.ID > 0 {
			out = append(out, AdminMarketReviewRow{
				RowKey:        "group:" + strconv.FormatInt(row.Group.ID, 10),
				IsMarketGroup: true,
				Market:        row.Market,
				Group:         row.Group,
				Children:      row.Children,
				Tags:          uniqueMarketTagsFromMarkets(row.Children),
			})
			continue
		}
		if row.Market == nil {
			continue
		}
		out = append(out, AdminMarketReviewRow{
			RowKey: "market:" + strconv.FormatInt(row.Market.ID, 10),
			Market: row.Market,
			Tags:   row.Market.Tags,
		})
	}
	return out
}

func uniqueMarketTagsFromMarkets(markets []*Market) []MarketTag {
	seen := map[string]bool{}
	tags := []MarketTag{}
	for _, market := range markets {
		if market == nil {
			continue
		}
		for _, tag := range market.Tags {
			key := tag.Slug
			if key == "" {
				key = strconv.FormatInt(tag.ID, 10)
			}
			if key == "" || seen[key] {
				continue
			}
			seen[key] = true
			tags = append(tags, tag)
		}
	}
	return tags
}

func groupedAdminDescriptionAmendmentRows(items []MarketDescriptionAmendment) []AdminDescriptionAmendmentReviewRow {
	rows := []AdminDescriptionAmendmentReviewRow{}
	groups := map[descriptionAmendmentGroupKey]int{}
	for _, item := range items {
		group := item.MarketGroup
		if group == nil || group.ID <= 0 {
			rows = append(rows, AdminDescriptionAmendmentReviewRow{
				RowKey:    "amendment:" + strconv.FormatInt(item.ID, 10),
				Amendment: item,
			})
			continue
		}
		key := descriptionAmendmentGroupKey{
			GroupID:      group.ID,
			Version:      item.Version,
			Status:       item.Status,
			Body:         item.Body,
			CreatedBy:    item.CreatedBy,
			SubmitReason: item.SubmitReason,
		}
		if existingIndex, ok := groups[key]; ok {
			rows[existingIndex].ChildAmendments = append(rows[existingIndex].ChildAmendments, item)
			continue
		}
		representative := item
		representative.MarketTitle = group.QuestionTitle
		representative.MarketDescription = group.Description
		row := AdminDescriptionAmendmentReviewRow{
			RowKey:                 descriptionAmendmentGroupRowKey(key),
			IsMarketGroupAmendment: true,
			Amendment:              representative,
			ChildAmendments:        []MarketDescriptionAmendment{item},
		}
		groups[key] = len(rows)
		rows = append(rows, row)
	}
	return rows
}

type descriptionAmendmentGroupKey struct {
	GroupID      int64
	Version      int
	Status       string
	Body         string
	CreatedBy    string
	SubmitReason string
}

func descriptionAmendmentGroupRowKey(key descriptionAmendmentGroupKey) string {
	hashInput := fmt.Sprintf("%d\x00%d\x00%s\x00%s\x00%s\x00%s", key.GroupID, key.Version, key.Status, key.CreatedBy, key.Body, key.SubmitReason)
	sum := sha256.Sum256([]byte(hashInput))
	return "group-amendment:" + strconv.FormatInt(key.GroupID, 10) + ":" + hex.EncodeToString(sum[:8])
}
