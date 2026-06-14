package markets_test

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	markets "socialpredict/internal/domain/markets"
	dusers "socialpredict/internal/domain/users"
	rmarkets "socialpredict/internal/repository/markets"
	"socialpredict/models"
	"socialpredict/models/modelstesting"
)

func seedAnswerAdditionGroup(t *testing.T, now time.Time) (*rmarkets.GormRepository, *markets.MarketGroup, []models.Market) {
	t.Helper()
	db := modelstesting.NewFakeDB(t)
	creator := modelstesting.GenerateUser("creator", 1000)
	creator.UserType = string(dusers.UserTypeModerator)
	creator.ModeratorStatus = string(dusers.ModeratorStatusActive)
	if err := db.Create(&creator).Error; err != nil {
		t.Fatalf("seed creator: %v", err)
	}

	children := []models.Market{
		modelstesting.GenerateMarket(501, creator.Username),
		modelstesting.GenerateMarket(502, creator.Username),
	}
	labels := []string{"Spain", "Canada"}
	for index := range children {
		children[index].QuestionTitle = "Spain vs Canada - " + labels[index]
		children[index].Description = "Child market"
		children[index].ResolutionDateTime = now.Add(48 * time.Hour)
		children[index].LifecycleStatus = markets.MarketLifecyclePublished
		children[index].YesLabel = "YES"
		children[index].NoLabel = "NO"
		children[index].StewardUsername = "steward"
		if err := db.Create(&children[index]).Error; err != nil {
			t.Fatalf("seed child market %d: %v", index, err)
		}
	}

	repo := rmarkets.NewGormRepository(db)
	group := &markets.MarketGroup{
		QuestionTitle:      "Spain vs Canada",
		Description:        "Grouped market",
		LifecycleStatus:    markets.MarketLifecyclePublished,
		ProposalCost:       10,
		CreatorUsername:    creator.Username,
		StewardUsername:    "steward",
		ResolutionDateTime: now.Add(48 * time.Hour),
	}
	members := []markets.MarketGroupMember{
		{MarketID: children[0].ID, AnswerLabel: labels[0], DisplayOrder: 0},
		{MarketID: children[1].ID, AnswerLabel: labels[1], DisplayOrder: 1},
	}
	if err := repo.CreateMarketGroup(context.Background(), group, members); err != nil {
		t.Fatalf("seed group: %v", err)
	}
	return repo, group, children
}

func answerAdditionUserService(t *testing.T, charged *[]int64) noopUserService {
	t.Helper()
	return newNoopUserService(func(service *noopUserService) {
		service.getPublicUserFunc = func(_ context.Context, username string) (*dusers.PublicUser, error) {
			return activeModeratorUser(username), nil
		}
		service.validateUserBalanceFunc = func(_ context.Context, username string, amount int64, maxDebt int64) error {
			if username != "proposer" {
				t.Fatalf("balance checked for %q, want proposer", username)
			}
			if amount <= 0 {
				t.Fatalf("balance amount = %d, want positive", amount)
			}
			return nil
		}
		service.applyTransactionFunc = func(_ context.Context, username string, amount int64, txType string) error {
			if username != "proposer" {
				t.Fatalf("transaction applied to %q, want proposer", username)
			}
			if txType != dusers.TransactionFee {
				t.Fatalf("transaction type = %q, want fee", txType)
			}
			*charged = append(*charged, amount)
			return nil
		}
	})
}

func flexibleAnswerAdditionUserService(charged *[]string) noopUserService {
	return newNoopUserService(func(service *noopUserService) {
		service.getPublicUserFunc = func(_ context.Context, username string) (*dusers.PublicUser, error) {
			return activeModeratorUser(username), nil
		}
		service.validateUserBalanceFunc = func(_ context.Context, username string, amount int64, maxDebt int64) error {
			*charged = append(*charged, "check:"+username)
			return nil
		}
		service.applyTransactionFunc = func(_ context.Context, username string, amount int64, txType string) error {
			*charged = append(*charged, "fee:"+username)
			return nil
		}
	})
}

func TestProposeMarketGroupAnswerAdditionCreatesPendingWithoutCharge(t *testing.T) {
	now := marketsTestTime()
	repo, group, _ := seedAnswerAdditionGroup(t, now)
	charged := []int64{}
	service := markets.NewService(repo, answerAdditionUserService(t, &charged), newFixedClock(now), markets.Config{
		GameMode:                                "moderator",
		MultipleChoiceBinaryAddAnswerCost:       4,
		MultipleChoiceBinaryHardAnswerSafetyCap: 50,
	})

	addition, err := service.ProposeMarketGroupAnswerAddition(context.Background(), group.ID, "proposer", markets.MarketGroupAnswerAdditionRequest{
		AnswerLabel: "Draw",
	})
	if err != nil {
		t.Fatalf("ProposeMarketGroupAnswerAddition returned error: %v", err)
	}
	if addition.Status != markets.MarketGroupAnswerAdditionStatusPending || addition.MarketID != 0 {
		t.Fatalf("unexpected pending addition: %+v", addition)
	}
	if addition.AdditionCost != 4 || addition.ProposedBy != "proposer" {
		t.Fatalf("unexpected proposal audit/cost fields: %+v", addition)
	}
	if len(charged) != 0 {
		t.Fatalf("pending proposal should not charge proposer, charged=%v", charged)
	}
	reloaded, err := repo.GetMarketGroup(context.Background(), group.ID)
	if err != nil {
		t.Fatalf("reload group: %v", err)
	}
	if len(reloaded.Members) != 2 {
		t.Fatalf("pending proposal should not add child member, got %d members", len(reloaded.Members))
	}
}

func TestProposeMarketGroupAnswerAdditionByStewardApprovesImmediately(t *testing.T) {
	now := marketsTestTime()
	repo, group, _ := seedAnswerAdditionGroup(t, now)
	charged := []string{}
	service := markets.NewService(repo, flexibleAnswerAdditionUserService(&charged), newFixedClock(now), markets.Config{
		GameMode:                                "moderator",
		MultipleChoiceBinaryAddAnswerCost:       4,
		MultipleChoiceBinaryHardAnswerSafetyCap: 50,
		MaximumDebtAllowed:                      500,
	})

	addition, err := service.ProposeMarketGroupAnswerAddition(context.Background(), group.ID, "steward", markets.MarketGroupAnswerAdditionRequest{
		AnswerLabel: "Draw",
	})
	if err != nil {
		t.Fatalf("ProposeMarketGroupAnswerAddition returned error: %v", err)
	}
	if addition.Status != markets.MarketGroupAnswerAdditionStatusApproved || addition.MarketID <= 0 {
		t.Fatalf("steward proposal should approve immediately: %+v", addition)
	}
	if addition.ReviewedBy != "steward" {
		t.Fatalf("reviewed by = %q, want steward", addition.ReviewedBy)
	}
	if strings.Join(charged, ",") != "check:steward,fee:steward" {
		t.Fatalf("unexpected charge calls: %v", charged)
	}
}

func TestProposeMarketGroupAnswerAdditionGroupAutoApprovesOtherModerator(t *testing.T) {
	now := marketsTestTime()
	repo, group, _ := seedAnswerAdditionGroup(t, now)
	if _, err := repo.UpdateMarketGroupAnswerAdditionAutoApproval(context.Background(), group.ID, true, now); err != nil {
		t.Fatalf("enable group auto approval: %v", err)
	}
	charged := []string{}
	service := markets.NewService(repo, flexibleAnswerAdditionUserService(&charged), newFixedClock(now), markets.Config{
		GameMode:                                "moderator",
		MultipleChoiceBinaryAddAnswerCost:       4,
		MultipleChoiceBinaryHardAnswerSafetyCap: 50,
		MaximumDebtAllowed:                      500,
	})

	addition, err := service.ProposeMarketGroupAnswerAddition(context.Background(), group.ID, "proposer", markets.MarketGroupAnswerAdditionRequest{
		AnswerLabel: "Draw",
	})
	if err != nil {
		t.Fatalf("ProposeMarketGroupAnswerAddition returned error: %v", err)
	}
	if addition.Status != markets.MarketGroupAnswerAdditionStatusApproved || addition.MarketID <= 0 {
		t.Fatalf("group auto-approved proposal should create child market: %+v", addition)
	}
	if addition.ReviewedBy != markets.MarketGroupAnswerAdditionApprovedByAuto {
		t.Fatalf("reviewed by = %q, want auto approval", addition.ReviewedBy)
	}
	if strings.Join(charged, ",") != "check:proposer,fee:proposer" {
		t.Fatalf("unexpected charge calls: %v", charged)
	}
}

func TestMarketGroupAnswerAdditionReviewerCanApprovePending(t *testing.T) {
	now := marketsTestTime()
	repo, group, _ := seedAnswerAdditionGroup(t, now)
	charged := []string{}
	service := markets.NewService(repo, flexibleAnswerAdditionUserService(&charged), newFixedClock(now), markets.Config{
		GameMode:                                "moderator",
		MultipleChoiceBinaryAddAnswerCost:       4,
		MultipleChoiceBinaryHardAnswerSafetyCap: 50,
		MaximumDebtAllowed:                      500,
	})

	addition, err := service.ProposeMarketGroupAnswerAddition(context.Background(), group.ID, "proposer", markets.MarketGroupAnswerAdditionRequest{AnswerLabel: "Draw"})
	if err != nil {
		t.Fatalf("propose returned error: %v", err)
	}
	approved, err := service.ApproveMarketGroupAnswerAdditionForReviewer(context.Background(), addition.ID, "steward", true)
	if err != nil {
		t.Fatalf("ApproveMarketGroupAnswerAdditionForReviewer returned error: %v", err)
	}
	if approved.Status != markets.MarketGroupAnswerAdditionStatusApproved || approved.ReviewedBy != "steward" {
		t.Fatalf("unexpected approved addition: %+v", approved)
	}
	if strings.Join(charged, ",") != "check:proposer,fee:proposer" {
		t.Fatalf("unexpected charge calls: %v", charged)
	}
}

func TestMarketGroupAnswerAdditionReviewerRejectsNonSteward(t *testing.T) {
	now := marketsTestTime()
	repo, group, _ := seedAnswerAdditionGroup(t, now)
	charged := []string{}
	service := markets.NewService(repo, flexibleAnswerAdditionUserService(&charged), newFixedClock(now), markets.Config{
		GameMode:                                "moderator",
		MultipleChoiceBinaryAddAnswerCost:       4,
		MultipleChoiceBinaryHardAnswerSafetyCap: 50,
	})

	addition, err := service.ProposeMarketGroupAnswerAddition(context.Background(), group.ID, "proposer", markets.MarketGroupAnswerAdditionRequest{AnswerLabel: "Draw"})
	if err != nil {
		t.Fatalf("propose returned error: %v", err)
	}
	_, err = service.ApproveMarketGroupAnswerAdditionForReviewer(context.Background(), addition.ID, "othermod", true)
	if !errors.Is(err, markets.ErrUnauthorized) {
		t.Fatalf("non-steward approve error = %v, want ErrUnauthorized", err)
	}
	if len(charged) != 0 {
		t.Fatalf("unauthorized review should not charge proposer, got %v", charged)
	}
}

func TestUpdateMarketGroupAnswerAdditionSettingsRequiresSteward(t *testing.T) {
	now := marketsTestTime()
	repo, group, _ := seedAnswerAdditionGroup(t, now)
	service := markets.NewService(repo, flexibleAnswerAdditionUserService(&[]string{}), newFixedClock(now), markets.Config{
		GameMode: "moderator",
	})

	updated, err := service.UpdateMarketGroupAnswerAdditionSettings(context.Background(), group.ID, "steward", true)
	if err != nil {
		t.Fatalf("UpdateMarketGroupAnswerAdditionSettings returned error: %v", err)
	}
	if !updated.AutoApproveAnswerAdditions {
		t.Fatalf("expected auto approve answer additions setting to be enabled")
	}

	_, err = service.UpdateMarketGroupAnswerAdditionSettings(context.Background(), group.ID, "othermod", false)
	if !errors.Is(err, markets.ErrUnauthorized) {
		t.Fatalf("non-steward update error = %v, want ErrUnauthorized", err)
	}
}

func TestProposeMarketGroupAnswerAdditionAutoApprovesCreatesChildAndSharedAmendments(t *testing.T) {
	now := marketsTestTime()
	repo, group, children := seedAnswerAdditionGroup(t, now)
	charged := []int64{}
	service := markets.NewService(repo, answerAdditionUserService(t, &charged), newFixedClock(now), markets.Config{
		GameMode:                                "moderator",
		MultipleChoiceBinaryAddAnswerCost:       4,
		MultipleChoiceBinaryHardAnswerSafetyCap: 50,
		MaximumDebtAllowed:                      500,
	})
	settings, err := service.GetMarketGovernanceSettings(context.Background())
	if err != nil {
		t.Fatalf("GetMarketGovernanceSettings returned error: %v", err)
	}
	enabled := true
	if _, err := service.UpdateMarketGovernanceSettings(context.Background(), markets.MarketGovernanceSettingsUpdate{
		AutoApproveMarketGroupAnswers: &enabled,
		Version:                       settings.Version,
		UpdatedBy:                     "admin",
	}); err != nil {
		t.Fatalf("enable answer auto approval: %v", err)
	}

	addition, err := service.ProposeMarketGroupAnswerAddition(context.Background(), group.ID, "proposer", markets.MarketGroupAnswerAdditionRequest{
		AnswerLabel: "Draw",
	})
	if err != nil {
		t.Fatalf("ProposeMarketGroupAnswerAddition returned error: %v", err)
	}
	if addition.Status != markets.MarketGroupAnswerAdditionStatusApproved || addition.MarketID <= 0 {
		t.Fatalf("unexpected approved addition: %+v", addition)
	}
	if addition.ReviewedBy != markets.MarketGroupAnswerAdditionApprovedByAuto || addition.ReviewedAt == nil {
		t.Fatalf("unexpected auto approval audit: %+v", addition)
	}
	if len(charged) != 1 || charged[0] != 4 {
		t.Fatalf("approved answer should charge proposal-time add-answer cost once, got %v", charged)
	}

	addedMarket, err := repo.GetByID(context.Background(), addition.MarketID)
	if err != nil {
		t.Fatalf("reload child market: %v", err)
	}
	if addedMarket.QuestionTitle != "Spain vs Canada - Draw" || addedMarket.ProposalCost != 0 {
		t.Fatalf("unexpected added child market: %+v", addedMarket)
	}
	if addedMarket.LifecycleStatus != markets.MarketLifecyclePublished || addedMarket.CurrentStewardUsername() != "steward" {
		t.Fatalf("added child should inherit published lifecycle and group steward: %+v", addedMarket)
	}

	reloaded, err := repo.GetMarketGroup(context.Background(), group.ID)
	if err != nil {
		t.Fatalf("reload group: %v", err)
	}
	if len(reloaded.Members) != 3 || reloaded.Members[2].AnswerLabel != "Draw" || reloaded.Members[2].MarketID != addition.MarketID {
		t.Fatalf("added child member not appended in group: %+v", reloaded.Members)
	}

	for _, marketID := range []int64{children[0].ID, children[1].ID, addition.MarketID} {
		amendments, err := repo.ListMarketDescriptionAmendments(context.Background(), markets.MarketDescriptionAmendmentFilters{
			MarketID: marketID,
			Status:   markets.DescriptionAmendmentStatusApproved,
		})
		if err != nil {
			t.Fatalf("list amendments for market %d: %v", marketID, err)
		}
		if len(amendments) != 1 {
			t.Fatalf("market %d amendments = %d, want 1", marketID, len(amendments))
		}
		body := amendments[0].Body
		if !strings.Contains(body, "Added answer option **Draw**") || !strings.Contains(body, "by [@proposer](/user/proposer)") {
			t.Fatalf("market %d generated amendment body missing addition audit link: %q", marketID, body)
		}
	}
}

func TestRejectMarketGroupAnswerAdditionDoesNotChargeOrCreateChild(t *testing.T) {
	now := marketsTestTime()
	repo, group, _ := seedAnswerAdditionGroup(t, now)
	charged := []int64{}
	service := markets.NewService(repo, answerAdditionUserService(t, &charged), newFixedClock(now), markets.Config{
		GameMode:                                "moderator",
		MultipleChoiceBinaryAddAnswerCost:       4,
		MultipleChoiceBinaryHardAnswerSafetyCap: 50,
	})
	addition, err := service.ProposeMarketGroupAnswerAddition(context.Background(), group.ID, "proposer", markets.MarketGroupAnswerAdditionRequest{
		AnswerLabel: "Draw",
	})
	if err != nil {
		t.Fatalf("ProposeMarketGroupAnswerAddition returned error: %v", err)
	}

	rejected, err := service.RejectMarketGroupAnswerAddition(context.Background(), addition.ID, "admin", "unclear answer")
	if err != nil {
		t.Fatalf("RejectMarketGroupAnswerAddition returned error: %v", err)
	}
	if rejected.Status != markets.MarketGroupAnswerAdditionStatusRejected || rejected.RejectionReason != "unclear answer" {
		t.Fatalf("unexpected rejected addition: %+v", rejected)
	}
	if len(charged) != 0 {
		t.Fatalf("rejected addition should not charge proposer, charged=%v", charged)
	}
	reloaded, err := repo.GetMarketGroup(context.Background(), group.ID)
	if err != nil {
		t.Fatalf("reload group: %v", err)
	}
	if len(reloaded.Members) != 2 {
		t.Fatalf("rejected addition should not add child member, got %d members", len(reloaded.Members))
	}
}

func TestProposeMarketGroupAnswerAdditionBlocksDuplicatePendingLabel(t *testing.T) {
	now := marketsTestTime()
	repo, group, _ := seedAnswerAdditionGroup(t, now)
	charged := []int64{}
	service := markets.NewService(repo, answerAdditionUserService(t, &charged), newFixedClock(now), markets.Config{
		GameMode:                                "moderator",
		MultipleChoiceBinaryAddAnswerCost:       4,
		MultipleChoiceBinaryHardAnswerSafetyCap: 50,
	})
	if _, err := service.ProposeMarketGroupAnswerAddition(context.Background(), group.ID, "proposer", markets.MarketGroupAnswerAdditionRequest{AnswerLabel: "Draw"}); err != nil {
		t.Fatalf("first proposal returned error: %v", err)
	}
	_, err := service.ProposeMarketGroupAnswerAddition(context.Background(), group.ID, "proposer", markets.MarketGroupAnswerAdditionRequest{AnswerLabel: " draw "})
	if !errors.Is(err, markets.ErrInvalidInput) {
		t.Fatalf("duplicate pending answer error = %v, want ErrInvalidInput", err)
	}
}
