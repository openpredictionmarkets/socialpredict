package markets_test

import (
	"context"
	"testing"
	"time"

	markets "socialpredict/internal/domain/markets"
)

func TestListAdminMarketReviewRowsUsesGroupedDiscoveryPage(t *testing.T) {
	now := marketsTestTime()
	group := &markets.MarketGroup{
		ID:              7,
		QuestionTitle:   "Grouped match",
		LifecycleStatus: markets.MarketLifecyclePublished,
		CreatorUsername: "moderator",
		Members: []markets.MarketGroupMember{
			{MarketID: 101, AnswerLabel: "Home"},
			{MarketID: 102, AnswerLabel: "Away"},
		},
		CreatedAt: now,
	}
	children := []*markets.Market{
		{ID: 101, QuestionTitle: "Grouped match - Home", LifecycleStatus: markets.MarketLifecyclePublished, Status: markets.MarketStatusActive, Tags: []markets.MarketTag{{ID: 1, Slug: "sports", DisplayName: "Sports"}}},
		{ID: 102, QuestionTitle: "Grouped match - Away", LifecycleStatus: markets.MarketLifecyclePublished, Status: markets.MarketStatusActive, Tags: []markets.MarketTag{{ID: 1, Slug: "sports", DisplayName: "Sports"}}},
	}
	repo := newProjectionRepo(func(repo *projectionRepo) {
		repo.listAdminMarketReviewRowsFunc = func(_ context.Context, filters markets.AdminMarketReviewFilters) (*markets.AdminMarketReviewPage, error) {
			if filters.Status != markets.MarketLifecyclePublished || filters.Query != "match" || filters.Limit != 10 || filters.Offset != 20 {
				t.Fatalf("filters = %+v", filters)
			}
			return &markets.AdminMarketReviewPage{
				Rows: []markets.AdminMarketReviewRow{{
					RowKey:        "group:7",
					IsMarketGroup: true,
					Market:        children[0],
					Group:         group,
					Children:      children,
					Tags:          []markets.MarketTag{{ID: 1, Slug: "sports", DisplayName: "Sports"}},
				}},
				Total:  3,
				Limit:  filters.Limit,
				Offset: filters.Offset,
			}, nil
		}
	})
	service := markets.NewService(repo, newNoopUserService(), newFixedClock(now), markets.Config{})

	page, err := service.ListAdminMarketReviewRows(context.Background(), markets.AdminMarketReviewFilters{
		Status: markets.MarketLifecyclePublished,
		Query:  "match",
		Limit:  10,
		Offset: 20,
	})
	if err != nil {
		t.Fatalf("ListAdminMarketReviewRows returned error: %v", err)
	}
	if page.Total != 3 || page.Limit != 10 || page.Offset != 20 || len(page.Rows) != 1 {
		t.Fatalf("unexpected page: %+v", page)
	}
	row := page.Rows[0]
	if row.RowKey != "group:7" || !row.IsMarketGroup || row.Group != group || len(row.Children) != 2 || len(row.Tags) != 1 || row.Tags[0].Slug != "sports" {
		t.Fatalf("unexpected grouped row: %+v", row)
	}
}

func TestListAdminMarketDescriptionAmendmentRowsGroupsSearchesAndPaginates(t *testing.T) {
	now := marketsTestTime()
	group := &markets.MarketGroup{
		ID:              8,
		QuestionTitle:   "Grouped orchard",
		Description:     "Grouped contract",
		LifecycleStatus: markets.MarketLifecyclePublished,
		Members: []markets.MarketGroupMember{
			{MarketID: 201, AnswerLabel: "Apples"},
			{MarketID: 202, AnswerLabel: "Pears"},
		},
	}
	marketByID := map[int64]*markets.Market{
		201: {ID: 201, QuestionTitle: "Grouped orchard - Apples", Description: "Apple contract"},
		202: {ID: 202, QuestionTitle: "Grouped orchard - Pears", Description: "Pear contract"},
		301: {ID: 301, QuestionTitle: "Solo football", Description: "Football contract"},
	}
	repo := newProjectionRepo(func(repo *projectionRepo) {
		repo.listDescriptionAmendmentReviewCandidatesFunc = func(_ context.Context, filters markets.AdminDescriptionAmendmentReviewFilters) ([]markets.MarketDescriptionAmendment, int, error) {
			if filters.Status != markets.DescriptionAmendmentStatusPending || filters.Query != "orchard" || filters.Limit != 1 || filters.Offset != 0 {
				t.Fatalf("filters = %+v", filters)
			}
			return []markets.MarketDescriptionAmendment{
				{ID: 1, MarketID: 201, Version: 2, Body: "Clarify grouped orchard", Status: markets.DescriptionAmendmentStatusPending, CreatedBy: "moderator", CreatedAt: now, UpdatedAt: now},
				{ID: 2, MarketID: 202, Version: 2, Body: "Clarify grouped orchard", Status: markets.DescriptionAmendmentStatusPending, CreatedBy: "moderator", CreatedAt: now, UpdatedAt: now},
			}, 1, nil
		}
		repo.getByIDFunc = func(_ context.Context, marketID int64) (*markets.Market, error) {
			if market := marketByID[marketID]; market != nil {
				return market, nil
			}
			return nil, markets.ErrMarketNotFound
		}
		repo.getMarketGroupForMarketFunc = func(_ context.Context, marketID int64) (*markets.MarketGroup, error) {
			if marketID == 201 || marketID == 202 {
				return group, nil
			}
			return nil, markets.ErrMarketGroupNotFound
		}
	})
	service := markets.NewService(repo, newNoopUserService(), newFixedClock(now), markets.Config{})

	page, err := service.ListAdminMarketDescriptionAmendmentRows(context.Background(), markets.AdminDescriptionAmendmentReviewFilters{
		Status: markets.DescriptionAmendmentStatusPending,
		Query:  "orchard",
		Limit:  1,
	})
	if err != nil {
		t.Fatalf("ListAdminMarketDescriptionAmendmentRows returned error: %v", err)
	}
	if page.Total != 1 || len(page.Rows) != 1 || page.Rows[0].RowKey == "" {
		t.Fatalf("unexpected page: %+v", page)
	}
	row := page.Rows[0]
	if !row.IsMarketGroupAmendment || row.Amendment.MarketTitle != "Grouped orchard" || len(row.ChildAmendments) != 2 {
		t.Fatalf("expected grouped amendment row, got %+v", row)
	}
}

func TestListAdminMarketGroupAnswerAdditionRowsUsesRepositorySearchAndTotal(t *testing.T) {
	now := time.Date(2026, 6, 20, 15, 0, 0, 0, time.UTC)
	repo := newProjectionRepo(func(repo *projectionRepo) {
		repo.listAnswerAdditionsForAdminReviewFunc = func(_ context.Context, filters markets.AdminAnswerAdditionReviewFilters) ([]markets.MarketGroupAnswerAddition, int, error) {
			if filters.Status != markets.MarketGroupAnswerAdditionStatusPending || filters.Query != "cedar" || filters.Limit != 5 || filters.Offset != 10 {
				t.Fatalf("filters = %+v", filters)
			}
			return []markets.MarketGroupAnswerAddition{{
				ID:          4,
				GroupID:     9,
				GroupTitle:  "Favorite tree",
				AnswerLabel: "Cedar",
				Status:      markets.MarketGroupAnswerAdditionStatusPending,
				CreatedAt:   now,
			}}, 12, nil
		}
	})
	service := markets.NewService(repo, newNoopUserService(), newFixedClock(now), markets.Config{})

	page, err := service.ListAdminMarketGroupAnswerAdditionRows(context.Background(), markets.AdminAnswerAdditionReviewFilters{
		Status: markets.MarketGroupAnswerAdditionStatusPending,
		Query:  "cedar",
		Limit:  5,
		Offset: 10,
	})
	if err != nil {
		t.Fatalf("ListAdminMarketGroupAnswerAdditionRows returned error: %v", err)
	}
	if page.Total != 12 || page.Limit != 5 || page.Offset != 10 || len(page.Rows) != 1 || page.Rows[0].AnswerLabel != "Cedar" {
		t.Fatalf("unexpected page: %+v", page)
	}
}
