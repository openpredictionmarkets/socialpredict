package notifications

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"socialpredict/models"
	"socialpredict/models/modelstesting"
	"socialpredict/util"
)

func setupTest(t *testing.T) {
	t.Helper()
	db := modelstesting.NewFakeDB(t)
	origDB := util.DB
	util.DB = db
	t.Cleanup(func() { util.DB = origDB })
	t.Setenv("JWT_SIGNING_KEY", "test-secret-key-for-testing")
}

func TestListNotifications_Empty(t *testing.T) {
	setupTest(t)

	user := modelstesting.GenerateUser("alice", 0)
	user.MustChangePassword = false
	if err := util.DB.Create(&user).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}

	req := httptest.NewRequest("GET", "/v0/notifications", nil)
	req.Header.Set("Authorization", "Bearer "+modelstesting.GenerateValidJWT("alice"))
	rec := httptest.NewRecorder()

	ListNotificationsHandler(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var result []NotificationResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &result); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(result) != 0 {
		t.Fatalf("expected empty, got %d", len(result))
	}
}

func TestListNotifications_ReturnsOwnOnly(t *testing.T) {
	setupTest(t)

	alice := modelstesting.GenerateUser("alice", 0)
	alice.MustChangePassword = false
	bob := modelstesting.GenerateUser("bob", 0)
	bob.MustChangePassword = false
	util.DB.Create(&alice)
	util.DB.Create(&bob)

	// alice has 2 notifications, bob has 1
	util.DB.Create(&models.Notification{Username: "alice", Type: "market_resolved", Message: "M1 resolved", MarketID: 1})
	util.DB.Create(&models.Notification{Username: "alice", Type: "market_resolved", Message: "M2 resolved", MarketID: 2})
	util.DB.Create(&models.Notification{Username: "bob", Type: "market_resolved", Message: "M3 resolved", MarketID: 3})

	req := httptest.NewRequest("GET", "/v0/notifications", nil)
	req.Header.Set("Authorization", "Bearer "+modelstesting.GenerateValidJWT("alice"))
	rec := httptest.NewRecorder()

	ListNotificationsHandler(rec, req)

	var result []NotificationResponse
	json.Unmarshal(rec.Body.Bytes(), &result)
	if len(result) != 2 {
		t.Fatalf("expected 2 notifications for alice, got %d", len(result))
	}
}

func TestUnreadCount_CountsOnlyUnread(t *testing.T) {
	setupTest(t)

	alice := modelstesting.GenerateUser("alice", 0)
	alice.MustChangePassword = false
	util.DB.Create(&alice)

	util.DB.Create(&models.Notification{Username: "alice", Type: "market_resolved", Message: "unread 1", IsRead: false})
	util.DB.Create(&models.Notification{Username: "alice", Type: "market_resolved", Message: "unread 2", IsRead: false})
	util.DB.Create(&models.Notification{Username: "alice", Type: "market_resolved", Message: "already read", IsRead: true})

	req := httptest.NewRequest("GET", "/v0/notifications/unread", nil)
	req.Header.Set("Authorization", "Bearer "+modelstesting.GenerateValidJWT("alice"))
	rec := httptest.NewRecorder()

	UnreadCountHandler(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var result map[string]int64
	json.Unmarshal(rec.Body.Bytes(), &result)
	if result["unreadCount"] != 2 {
		t.Fatalf("expected unreadCount=2, got %d", result["unreadCount"])
	}
}

func TestMarkAllRead_MarksUnread(t *testing.T) {
	setupTest(t)

	alice := modelstesting.GenerateUser("alice", 0)
	alice.MustChangePassword = false
	util.DB.Create(&alice)

	util.DB.Create(&models.Notification{Username: "alice", Type: "market_resolved", Message: "msg1", IsRead: false})
	util.DB.Create(&models.Notification{Username: "alice", Type: "market_resolved", Message: "msg2", IsRead: false})

	req := httptest.NewRequest("PATCH", "/v0/notifications/read-all", nil)
	req.Header.Set("Authorization", "Bearer "+modelstesting.GenerateValidJWT("alice"))
	rec := httptest.NewRecorder()

	MarkAllReadHandler(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d: %s", rec.Code, rec.Body.String())
	}

	var count int64
	util.DB.Model(&models.Notification{}).Where("username = ? AND is_read = ?", "alice", false).Count(&count)
	if count != 0 {
		t.Fatalf("expected 0 unread after mark-all-read, got %d", count)
	}
}

func TestCreateMarketResolutionNotifications_OnePerUser(t *testing.T) {
	setupTest(t)

	alice := modelstesting.GenerateUser("alice", 100)
	alice.MustChangePassword = false
	bob := modelstesting.GenerateUser("bob", 100)
	bob.MustChangePassword = false
	util.DB.Create(&alice)
	util.DB.Create(&bob)

	market := modelstesting.GenerateMarket(1, "alice")
	market.IsResolved = true
	market.ResolutionResult = "YES"
	util.DB.Create(&market)

	// alice placed 2 bets, bob placed 1 — should produce 2 notifications (one per user)
	util.DB.Create(&modelstesting.GenerateBet(100, "YES", "alice", uint(market.ID), 0))
	util.DB.Create(&modelstesting.GenerateBet(50, "YES", "alice", uint(market.ID), 0))
	util.DB.Create(&modelstesting.GenerateBet(75, "NO", "bob", uint(market.ID), 0))

	if err := CreateMarketResolutionNotifications(util.DB, &market); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var notifs []models.Notification
	util.DB.Where("market_id = ?", market.ID).Find(&notifs)
	if len(notifs) != 2 {
		t.Fatalf("expected 2 notifications (one per unique bettor), got %d", len(notifs))
	}

	usernames := map[string]bool{}
	for _, n := range notifs {
		usernames[n.Username] = true
		if n.Type != "market_resolved" {
			t.Errorf("expected type market_resolved, got %s", n.Type)
		}
		if n.IsRead {
			t.Error("expected notification to start unread")
		}
	}
	if !usernames["alice"] || !usernames["bob"] {
		t.Error("expected notifications for both alice and bob")
	}
}
