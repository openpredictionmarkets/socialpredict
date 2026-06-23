package markets_test

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	dmarkets "socialpredict/internal/domain/markets"
	dusers "socialpredict/internal/domain/users"
	rmarkets "socialpredict/internal/repository/markets"
	"socialpredict/models"
	"socialpredict/models/modelstesting"

	"gorm.io/gorm"
)

func TestGormRepositoryCreateMarketGroupRollsBackWhenGroupLinkFails(t *testing.T) {
	db := modelstesting.NewFakeDB(t)
	sqlDB, _ := db.DB()
	if sqlDB != nil {
		t.Cleanup(func() { _ = sqlDB.Close() })
	}

	creator := groupedTransactionModerator("group_creator", 1000)
	if err := db.Create(&creator).Error; err != nil {
		t.Fatalf("seed creator: %v", err)
	}

	createErr := errors.New("forced group member create failure")
	registerGroupedTransactionCreateFailure(t, db, "feature14_fail_group_member_create", func(tx *gorm.DB) bool {
		switch tx.Statement.Dest.(type) {
		case *[]models.MarketGroupMember, *models.MarketGroupMember:
			return true
		default:
			return false
		}
	}, createErr)

	repo := rmarkets.NewGormRepository(db)
	service := dmarkets.NewService(repo, nil, nil, dmarkets.Config{
		GameMode:                                "moderator",
		CreateMarketCost:                        10,
		MaximumDebtAllowed:                      500,
		MultipleChoiceBinaryHardAnswerSafetyCap: 50,
	})

	_, err := service.CreateMarketGroup(context.Background(), dmarkets.MarketGroupCreateRequest{
		QuestionTitle:              "Rollback group",
		Description:                "Rollback group",
		ResolutionDateTime:         time.Now().UTC().Add(48 * time.Hour),
		AnswerLabels:               []string{"A", "B", "C"},
		AutoApproveAnswerAdditions: false,
	}, creator.Username)
	if !errors.Is(err, createErr) {
		t.Fatalf("CreateMarketGroup error = %v, want forced create error", err)
	}

	assertGroupedTransactionUserBalance(t, db, creator.Username, 1000)
	assertGroupedTransactionMarketCount(t, db, "Rollback group%", 0)
	assertGroupedTransactionCount(t, db, &models.MarketGroup{}, 0)
	assertGroupedTransactionCount(t, db, &models.MarketGroupMember{}, 0)
}

func TestGormRepositoryApproveAnswerAdditionRollsBackWhenGroupLinkFails(t *testing.T) {
	db := modelstesting.NewFakeDB(t)
	sqlDB, _ := db.DB()
	if sqlDB != nil {
		t.Cleanup(func() { _ = sqlDB.Close() })
	}

	creator := groupedTransactionModerator("answer_creator", 1000)
	proposer := groupedTransactionModerator("answer_proposer", 1000)
	if err := db.Create(&creator).Error; err != nil {
		t.Fatalf("seed creator: %v", err)
	}
	if err := db.Create(&proposer).Error; err != nil {
		t.Fatalf("seed proposer: %v", err)
	}

	repo := rmarkets.NewGormRepository(db)
	group := seedGroupedTransactionGroup(t, db, repo, creator.Username, "Answer rollback", []string{"One", "Two"})
	addition, err := repo.CreateMarketGroupAnswerAddition(context.Background(), dmarkets.MarketGroupAnswerAddition{
		GroupID:      group.ID,
		GroupTitle:   group.QuestionTitle,
		AnswerLabel:  "Three",
		Status:       dmarkets.MarketGroupAnswerAdditionStatusPending,
		ProposedBy:   proposer.Username,
		AdditionCost: 4,
		CreatedAt:    time.Now().UTC(),
		UpdatedAt:    time.Now().UTC(),
	})
	if err != nil {
		t.Fatalf("seed answer addition: %v", err)
	}

	createErr := errors.New("forced added answer link failure")
	registerGroupedTransactionCreateFailure(t, db, "feature14_fail_added_answer_link", func(tx *gorm.DB) bool {
		switch tx.Statement.Dest.(type) {
		case *models.MarketGroupMember:
			return true
		default:
			return false
		}
	}, createErr)

	service := dmarkets.NewService(repo, nil, nil, dmarkets.Config{
		GameMode:                          "moderator",
		MaximumDebtAllowed:                500,
		MultipleChoiceBinaryAddAnswerCost: 4,
	})

	_, err = service.ApproveMarketGroupAnswerAddition(context.Background(), addition.ID, creator.Username, true)
	if !errors.Is(err, createErr) {
		t.Fatalf("ApproveMarketGroupAnswerAddition error = %v, want forced create error", err)
	}

	assertGroupedTransactionUserBalance(t, db, proposer.Username, 1000)
	assertGroupedTransactionMarketCount(t, db, "Answer rollback - Three", 0)
	assertGroupedTransactionGroupMemberCount(t, db, group.ID, 2)

	var refreshed models.MarketGroupAnswerAddition
	if err := db.First(&refreshed, addition.ID).Error; err != nil {
		t.Fatalf("reload addition: %v", err)
	}
	if refreshed.Status != dmarkets.MarketGroupAnswerAdditionStatusPending || refreshed.MarketID != 0 {
		t.Fatalf("expected pending addition after rollback, got status=%q market_id=%d", refreshed.Status, refreshed.MarketID)
	}
}

func TestGormRepositoryResolveMarketGroupRollsBackChildResolutionsWhenParentUpdateFails(t *testing.T) {
	db := modelstesting.NewFakeDB(t)
	sqlDB, _ := db.DB()
	if sqlDB != nil {
		t.Cleanup(func() { _ = sqlDB.Close() })
	}

	creator := groupedTransactionModerator("resolve_creator", 1000)
	if err := db.Create(&creator).Error; err != nil {
		t.Fatalf("seed creator: %v", err)
	}

	repo := rmarkets.NewGormRepository(db)
	group := seedGroupedTransactionGroup(t, db, repo, creator.Username, "Resolution rollback", []string{"Yes", "No"})

	updateErr := errors.New("forced parent group resolution failure")
	registerGroupedTransactionUpdateFailure(t, db, "feature14_fail_group_resolve_update", func(tx *gorm.DB) bool {
		return tx.Statement.Schema != nil && tx.Statement.Schema.Name == "MarketGroup"
	}, updateErr)

	service := dmarkets.NewService(repo, nil, nil, dmarkets.Config{
		GameMode:           "moderator",
		InitialBetFee:      1,
		MaximumDebtAllowed: 500,
	})

	_, err := service.ResolveMarketGroup(context.Background(), group.ID, dmarkets.MarketGroupResolveRequest{
		Mode:            dmarkets.MarketGroupResolveModeExclusiveYes,
		WinningMarketID: group.Members[0].MarketID,
	}, creator.Username)
	if !errors.Is(err, updateErr) {
		t.Fatalf("ResolveMarketGroup error = %v, want forced update error", err)
	}

	for _, member := range group.Members {
		var child models.Market
		if err := db.First(&child, member.MarketID).Error; err != nil {
			t.Fatalf("reload child market %d: %v", member.MarketID, err)
		}
		if child.IsResolved || child.LifecycleStatus == dmarkets.MarketLifecycleResolved || child.ResolutionResult != "" {
			t.Fatalf("child market %d should not be resolved after rollback: %+v", member.MarketID, child)
		}
	}

	var refreshedGroup models.MarketGroup
	if err := db.First(&refreshedGroup, group.ID).Error; err != nil {
		t.Fatalf("reload group: %v", err)
	}
	if refreshedGroup.LifecycleStatus != dmarkets.MarketLifecyclePublished {
		t.Fatalf("group lifecycle = %q, want published after rollback", refreshedGroup.LifecycleStatus)
	}
}

func groupedTransactionModerator(username string, balance int64) models.User {
	user := modelstesting.GenerateUser(username, balance)
	user.UserType = string(dusers.UserTypeModerator)
	user.ModeratorStatus = string(dusers.ModeratorStatusActive)
	return user
}

func seedGroupedTransactionGroup(t *testing.T, db *gorm.DB, repo *rmarkets.GormRepository, creatorUsername string, title string, labels []string) *dmarkets.MarketGroup {
	t.Helper()

	members := make([]dmarkets.MarketGroupMember, 0, len(labels))
	for index, label := range labels {
		child := modelstesting.GenerateMarket(0, creatorUsername)
		child.QuestionTitle = title + " - " + label
		child.Description = "Child market"
		child.ResolutionDateTime = time.Now().UTC().Add(48 * time.Hour)
		child.LifecycleStatus = dmarkets.MarketLifecyclePublished
		child.StewardUsername = creatorUsername
		if err := db.Create(&child).Error; err != nil {
			t.Fatalf("seed child market %q: %v", label, err)
		}
		members = append(members, dmarkets.MarketGroupMember{
			MarketID:     child.ID,
			AnswerLabel:  label,
			DisplayOrder: index,
		})
	}

	group := &dmarkets.MarketGroup{
		QuestionTitle:      title,
		Description:        title,
		LifecycleStatus:    dmarkets.MarketLifecyclePublished,
		ProposalCost:       10,
		CreatorUsername:    creatorUsername,
		StewardUsername:    creatorUsername,
		ResolutionDateTime: time.Now().UTC().Add(48 * time.Hour),
		CreatedAt:          time.Now().UTC(),
		UpdatedAt:          time.Now().UTC(),
	}
	if err := repo.CreateMarketGroup(context.Background(), group, members); err != nil {
		t.Fatalf("seed group: %v", err)
	}
	return group
}

func registerGroupedTransactionCreateFailure(t *testing.T, db *gorm.DB, name string, shouldFail func(*gorm.DB) bool, err error) {
	t.Helper()
	if err := db.Callback().Create().Before("gorm:create").Register(name, func(tx *gorm.DB) {
		if shouldFail(tx) {
			tx.AddError(err)
		}
	}); err != nil {
		t.Fatalf("register create callback: %v", err)
	}
}

func registerGroupedTransactionUpdateFailure(t *testing.T, db *gorm.DB, name string, shouldFail func(*gorm.DB) bool, err error) {
	t.Helper()
	if err := db.Callback().Update().Before("gorm:update").Register(name, func(tx *gorm.DB) {
		if shouldFail(tx) {
			tx.AddError(err)
		}
	}); err != nil {
		t.Fatalf("register update callback: %v", err)
	}
}

func assertGroupedTransactionUserBalance(t *testing.T, db *gorm.DB, username string, want int64) {
	t.Helper()
	var user models.User
	if err := db.Where("username = ?", username).First(&user).Error; err != nil {
		t.Fatalf("reload user %q: %v", username, err)
	}
	if user.AccountBalance != want {
		t.Fatalf("balance for %s = %d, want %d", username, user.AccountBalance, want)
	}
}

func assertGroupedTransactionMarketCount(t *testing.T, db *gorm.DB, titleLike string, want int64) {
	t.Helper()
	var count int64
	query := db.Model(&models.Market{})
	if strings.Contains(titleLike, "%") {
		query = query.Where("question_title LIKE ?", titleLike)
	} else {
		query = query.Where("question_title = ?", titleLike)
	}
	if err := query.Count(&count).Error; err != nil {
		t.Fatalf("count markets: %v", err)
	}
	if count != want {
		t.Fatalf("market count for %q = %d, want %d", titleLike, count, want)
	}
}

func assertGroupedTransactionGroupMemberCount(t *testing.T, db *gorm.DB, groupID int64, want int64) {
	t.Helper()
	var count int64
	if err := db.Model(&models.MarketGroupMember{}).Where("group_id = ?", groupID).Count(&count).Error; err != nil {
		t.Fatalf("count group members: %v", err)
	}
	if count != want {
		t.Fatalf("group member count = %d, want %d", count, want)
	}
}

func assertGroupedTransactionCount(t *testing.T, db *gorm.DB, model any, want int64) {
	t.Helper()
	var count int64
	if err := db.Model(model).Count(&count).Error; err != nil {
		t.Fatalf("count rows: %v", err)
	}
	if count != want {
		t.Fatalf("row count = %d, want %d", count, want)
	}
}
