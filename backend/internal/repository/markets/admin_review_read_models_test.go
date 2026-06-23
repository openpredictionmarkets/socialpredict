package markets

import (
	"context"
	"testing"
	"time"

	dmarkets "socialpredict/internal/domain/markets"
	"socialpredict/models"
	"socialpredict/models/modelstesting"

	"gorm.io/gorm"
)

func TestGormRepositoryListAdminMarketReviewRowsGroupsBeforePagination(t *testing.T) {
	db := modelstesting.NewFakeDB(t)
	repo := NewGormRepository(db)
	ctx := context.Background()
	now := time.Now().UTC().Truncate(time.Second)
	creator := seedDiscoveryUser(t, db, "admin_review_creator")

	seedDiscoveryGroup(t, db, discoveryGroupSeed{
		GroupID:     90,
		Title:       "Admin grouped market",
		Description: "Grouped admin review row",
		Creator:     creator,
		CreatedAt:   now.Add(10 * time.Minute),
		Answers: []discoveryAnswerSeed{
			{MarketID: 9001, Label: "Alpha", Title: "Admin grouped market - Alpha"},
			{MarketID: 9002, Label: "Beta", Title: "Admin grouped market - Beta"},
		},
	})
	standalone := seedDiscoveryMarket(t, db, 9003, creator, "Standalone admin market", now)

	firstPage, err := repo.ListAdminMarketReviewRows(ctx, dmarkets.AdminMarketReviewFilters{
		Status: dmarkets.MarketLifecyclePublished,
		Limit:  1,
		Offset: 0,
	})
	if err != nil {
		t.Fatalf("ListAdminMarketReviewRows first page: %v", err)
	}
	if firstPage.Total != 2 || len(firstPage.Rows) != 1 {
		t.Fatalf("expected one grouped row with total 2, got total=%d rows=%d", firstPage.Total, len(firstPage.Rows))
	}
	if !firstPage.Rows[0].IsMarketGroup || firstPage.Rows[0].Group == nil || firstPage.Rows[0].Group.ID != 90 || len(firstPage.Rows[0].Children) != 2 {
		t.Fatalf("expected first page grouped row, got %+v", firstPage.Rows[0])
	}

	secondPage, err := repo.ListAdminMarketReviewRows(ctx, dmarkets.AdminMarketReviewFilters{
		Status: dmarkets.MarketLifecyclePublished,
		Limit:  1,
		Offset: 1,
	})
	if err != nil {
		t.Fatalf("ListAdminMarketReviewRows second page: %v", err)
	}
	if secondPage.Total != 2 || len(secondPage.Rows) != 1 || secondPage.Rows[0].IsMarketGroup || secondPage.Rows[0].Market == nil || secondPage.Rows[0].Market.ID != standalone.ID {
		t.Fatalf("expected standalone row on second page, got total=%d rows=%+v", secondPage.Total, secondPage.Rows)
	}
}

func TestGormRepositoryListDescriptionAmendmentReviewCandidatesGroupsBeforePagination(t *testing.T) {
	db := modelstesting.NewFakeDB(t)
	repo := NewGormRepository(db)
	ctx := context.Background()
	now := time.Now().UTC().Truncate(time.Second)

	creator := modelstesting.GenerateUser("review_model_creator", 1000)
	if err := db.Create(&creator).Error; err != nil {
		t.Fatalf("seed creator: %v", err)
	}

	childA := seedAdminReviewMarket(t, db, 9101, creator.Username, "Child answer alpha", "Alpha contract", now)
	childB := seedAdminReviewMarket(t, db, 9102, creator.Username, "Child answer beta", "Beta contract", now)
	standalone := seedAdminReviewMarket(t, db, 9201, creator.Username, "Orchard standalone", "Standalone contract", now.Add(-time.Minute))
	ignored := seedAdminReviewMarket(t, db, 9301, creator.Username, "Football standalone", "Other contract", now.Add(-2*time.Minute))

	group := models.MarketGroup{
		ID:                 91,
		QuestionTitle:      "Orchard grouped question",
		Description:        "Parent contract text",
		GroupType:          dmarkets.MarketGroupTypeMultipleChoiceBinary,
		ProbabilityPolicy:  dmarkets.MarketGroupProbabilityPolicyIndependentBinary,
		ResolutionPolicy:   dmarkets.MarketGroupResolutionPolicyIndependentChildren,
		LifecycleStatus:    dmarkets.MarketLifecyclePublished,
		ProposalCost:       10,
		CreatorUsername:    creator.Username,
		StewardUsername:    creator.Username,
		ResolutionDateTime: now.Add(30 * 24 * time.Hour),
	}
	group.CreatedAt = now
	group.UpdatedAt = now
	if err := db.Create(&group).Error; err != nil {
		t.Fatalf("seed group: %v", err)
	}
	members := []models.MarketGroupMember{
		{GroupID: group.ID, MarketID: childA.ID, AnswerLabel: "Apples", DisplayOrder: 0},
		{GroupID: group.ID, MarketID: childB.ID, AnswerLabel: "Pears", DisplayOrder: 1},
	}
	if err := db.Create(&members).Error; err != nil {
		t.Fatalf("seed members: %v", err)
	}

	amendments := []models.MarketDescriptionAmendment{
		{MarketID: childA.ID, Version: 2, Body: "Clarify grouped fruit contract", Status: dmarkets.DescriptionAmendmentStatusPending, CreatedBy: creator.Username, SubmitReason: "same batch"},
		{MarketID: childB.ID, Version: 2, Body: "Clarify grouped fruit contract", Status: dmarkets.DescriptionAmendmentStatusPending, CreatedBy: creator.Username, SubmitReason: "same batch"},
		{MarketID: standalone.ID, Version: 2, Body: "Clarify standalone orchard contract", Status: dmarkets.DescriptionAmendmentStatusPending, CreatedBy: creator.Username},
		{MarketID: ignored.ID, Version: 2, Body: "Clarify unrelated contract", Status: dmarkets.DescriptionAmendmentStatusPending, CreatedBy: creator.Username},
	}
	for index := range amendments {
		amendments[index].CreatedAt = now.Add(time.Duration(-index) * time.Minute)
		amendments[index].UpdatedAt = amendments[index].CreatedAt
	}
	if err := db.Create(&amendments).Error; err != nil {
		t.Fatalf("seed amendments: %v", err)
	}

	firstPage, total, err := repo.ListMarketDescriptionAmendmentReviewCandidates(ctx, dmarkets.AdminDescriptionAmendmentReviewFilters{
		Status: dmarkets.DescriptionAmendmentStatusPending,
		Query:  "orchard",
		Limit:  1,
		Offset: 0,
	})
	if err != nil {
		t.Fatalf("ListMarketDescriptionAmendmentReviewCandidates first page: %v", err)
	}
	if total != 2 {
		t.Fatalf("expected 2 grouped admin rows, got %d", total)
	}
	if len(firstPage) != 2 {
		t.Fatalf("expected first page to include both grouped child amendments, got %d", len(firstPage))
	}
	if firstPage[0].MarketID != childA.ID || firstPage[1].MarketID != childB.ID {
		t.Fatalf("expected grouped child amendments on first page, got %+v", firstPage)
	}

	secondPage, total, err := repo.ListMarketDescriptionAmendmentReviewCandidates(ctx, dmarkets.AdminDescriptionAmendmentReviewFilters{
		Status: dmarkets.DescriptionAmendmentStatusPending,
		Query:  "orchard",
		Limit:  1,
		Offset: 1,
	})
	if err != nil {
		t.Fatalf("ListMarketDescriptionAmendmentReviewCandidates second page: %v", err)
	}
	if total != 2 || len(secondPage) != 1 || secondPage[0].MarketID != standalone.ID {
		t.Fatalf("expected standalone amendment on second page with total 2, got total=%d rows=%+v", total, secondPage)
	}
}

func TestGormRepositoryListAnswerAdditionsForAdminReviewSearchesAndCounts(t *testing.T) {
	db := modelstesting.NewFakeDB(t)
	repo := NewGormRepository(db)
	ctx := context.Background()
	now := time.Now().UTC().Truncate(time.Second)

	creator := modelstesting.GenerateUser("answer_review_creator", 1000)
	if err := db.Create(&creator).Error; err != nil {
		t.Fatalf("seed creator: %v", err)
	}
	group := models.MarketGroup{
		ID:                 92,
		QuestionTitle:      "Favorite tree",
		Description:        "Tree answer options",
		GroupType:          dmarkets.MarketGroupTypeMultipleChoiceBinary,
		ProbabilityPolicy:  dmarkets.MarketGroupProbabilityPolicyIndependentBinary,
		ResolutionPolicy:   dmarkets.MarketGroupResolutionPolicyIndependentChildren,
		LifecycleStatus:    dmarkets.MarketLifecyclePublished,
		CreatorUsername:    creator.Username,
		StewardUsername:    creator.Username,
		ResolutionDateTime: now.Add(30 * 24 * time.Hour),
	}
	group.CreatedAt = now
	group.UpdatedAt = now
	if err := db.Create(&group).Error; err != nil {
		t.Fatalf("seed group: %v", err)
	}
	additions := []models.MarketGroupAnswerAddition{
		{GroupID: group.ID, AnswerLabel: "Cedar", Status: dmarkets.MarketGroupAnswerAdditionStatusPending, ProposedBy: "tree_mod", AdditionCost: 2},
		{GroupID: group.ID, AnswerLabel: "Pine", Status: dmarkets.MarketGroupAnswerAdditionStatusPending, ProposedBy: "tree_mod", AdditionCost: 2},
		{GroupID: group.ID, AnswerLabel: "Football", Status: dmarkets.MarketGroupAnswerAdditionStatusRejected, ProposedBy: "sports_mod", AdditionCost: 2},
	}
	for index := range additions {
		additions[index].CreatedAt = now.Add(time.Duration(-index) * time.Minute)
		additions[index].UpdatedAt = additions[index].CreatedAt
	}
	if err := db.Create(&additions).Error; err != nil {
		t.Fatalf("seed answer additions: %v", err)
	}

	rows, total, err := repo.ListMarketGroupAnswerAdditionsForAdminReview(ctx, dmarkets.AdminAnswerAdditionReviewFilters{
		Status: dmarkets.MarketGroupAnswerAdditionStatusPending,
		Query:  "tree",
		Limit:  1,
		Offset: 1,
	})
	if err != nil {
		t.Fatalf("ListMarketGroupAnswerAdditionsForAdminReview: %v", err)
	}
	if total != 2 {
		t.Fatalf("expected 2 matching pending additions, got %d", total)
	}
	if len(rows) != 1 || rows[0].AnswerLabel != "Pine" || rows[0].GroupTitle != "Favorite tree" || rows[0].MarketGroup == nil {
		t.Fatalf("expected paged Pine row with hydrated group, got %+v", rows)
	}
}

func TestGormRepositorySetMarketGroupTagsRewritesEveryChild(t *testing.T) {
	db := modelstesting.NewFakeDB(t)
	repo := NewGormRepository(db)
	ctx := context.Background()
	now := time.Now().UTC().Truncate(time.Second)
	creator := seedDiscoveryUser(t, db, "group_tag_creator")
	children := seedDiscoveryGroup(t, db, discoveryGroupSeed{
		GroupID:     93,
		Title:       "Grouped tags",
		Description: "Grouped tags",
		Creator:     creator,
		CreatedAt:   now,
		Answers: []discoveryAnswerSeed{
			{MarketID: 9301, Label: "A", Title: "Grouped tags - A"},
			{MarketID: 9302, Label: "B", Title: "Grouped tags - B"},
		},
	})
	oldTag := models.MarketTag{ID: 930, Slug: "old", DisplayName: "Old", IsActive: true}
	newTag := models.MarketTag{ID: 931, Slug: "new", DisplayName: "New", IsActive: true}
	if err := db.Create(&oldTag).Error; err != nil {
		t.Fatalf("seed old tag: %v", err)
	}
	if err := db.Create(&newTag).Error; err != nil {
		t.Fatalf("seed new tag: %v", err)
	}
	for _, child := range children {
		if err := db.Create(&models.MarketTagAssignment{MarketID: child.ID, TagID: oldTag.ID, AssignedBy: "old", Source: "old"}).Error; err != nil {
			t.Fatalf("seed old assignment: %v", err)
		}
	}

	tags, err := repo.SetMarketGroupTags(ctx, 93, []string{"new"}, "admin", dmarkets.MarketTagAssignmentSourceAdmin, now)
	if err != nil {
		t.Fatalf("SetMarketGroupTags returned error: %v", err)
	}
	if len(tags) != 1 || tags[0].Slug != "new" {
		t.Fatalf("unexpected returned tags: %+v", tags)
	}
	for _, child := range children {
		var assignments []models.MarketTagAssignment
		if err := db.Where("market_id = ?", child.ID).Find(&assignments).Error; err != nil {
			t.Fatalf("load assignments for child %d: %v", child.ID, err)
		}
		if len(assignments) != 1 || assignments[0].TagID != newTag.ID || assignments[0].AssignedBy != "admin" || assignments[0].Source != dmarkets.MarketTagAssignmentSourceAdmin {
			t.Fatalf("unexpected assignments for child %d: %+v", child.ID, assignments)
		}
	}
}

func seedAdminReviewMarket(t *testing.T, db *gorm.DB, id int64, creator string, title string, description string, createdAt time.Time) models.Market {
	t.Helper()
	market := modelstesting.GenerateMarket(id, creator)
	market.QuestionTitle = title
	market.Description = description
	market.LifecycleStatus = dmarkets.MarketLifecyclePublished
	market.CreatedAt = createdAt
	market.UpdatedAt = createdAt
	market.ResolutionDateTime = createdAt.Add(30 * 24 * time.Hour)
	if err := db.Create(&market).Error; err != nil {
		t.Fatalf("seed market %d: %v", id, err)
	}
	return market
}
