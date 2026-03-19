package notifications

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"socialpredict/middleware"
	"socialpredict/models"
	"socialpredict/util"

	"gorm.io/gorm"
)

// NotificationResponse is the public shape of a notification returned by the API.
type NotificationResponse struct {
	ID        uint      `json:"id"`
	Type      string    `json:"type"`
	MarketID  uint      `json:"marketId"`
	Message   string    `json:"message"`
	IsRead    bool      `json:"isRead"`
	CreatedAt time.Time `json:"createdAt"`
}

// ListNotificationsHandler returns the authenticated user's notifications, newest first.
func ListNotificationsHandler(w http.ResponseWriter, r *http.Request) {
	db := util.GetDB()
	user, httpErr := middleware.ValidateUserAndEnforcePasswordChangeGetUser(r, db)
	if httpErr != nil {
		http.Error(w, httpErr.Message, httpErr.StatusCode)
		return
	}

	var notifs []models.Notification
	if err := db.Where("username = ?", user.Username).
		Order("created_at desc").
		Find(&notifs).Error; err != nil {
		http.Error(w, "failed to fetch notifications", http.StatusInternalServerError)
		return
	}

	responses := make([]NotificationResponse, len(notifs))
	for i, n := range notifs {
		responses[i] = NotificationResponse{
			ID:        n.ID,
			Type:      n.Type,
			MarketID:  n.MarketID,
			Message:   n.Message,
			IsRead:    n.IsRead,
			CreatedAt: n.CreatedAt,
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(responses)
}

// UnreadCountHandler returns the number of unread notifications for the authenticated user.
func UnreadCountHandler(w http.ResponseWriter, r *http.Request) {
	db := util.GetDB()
	user, httpErr := middleware.ValidateUserAndEnforcePasswordChangeGetUser(r, db)
	if httpErr != nil {
		http.Error(w, httpErr.Message, httpErr.StatusCode)
		return
	}

	var count int64
	if err := db.Model(&models.Notification{}).
		Where("username = ? AND is_read = ?", user.Username, false).
		Count(&count).Error; err != nil {
		http.Error(w, "failed to count notifications", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]int64{"unreadCount": count})
}

// MarkAllReadHandler marks all of the authenticated user's notifications as read.
func MarkAllReadHandler(w http.ResponseWriter, r *http.Request) {
	db := util.GetDB()
	user, httpErr := middleware.ValidateUserAndEnforcePasswordChangeGetUser(r, db)
	if httpErr != nil {
		http.Error(w, httpErr.Message, httpErr.StatusCode)
		return
	}

	if err := db.Model(&models.Notification{}).
		Where("username = ? AND is_read = ?", user.Username, false).
		Update("is_read", true).Error; err != nil {
		http.Error(w, "failed to mark notifications read", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// CreateMarketResolutionNotifications generates a "market_resolved" notification for
// every unique user who placed a bet on the given market. Call this after resolution.
func CreateMarketResolutionNotifications(db *gorm.DB, market *models.Market) error {
	var bets []models.Bet
	if err := db.Where("market_id = ?", market.ID).Find(&bets).Error; err != nil {
		return fmt.Errorf("fetch bets for notifications: %w", err)
	}

	seen := make(map[string]bool)
	for _, bet := range bets {
		if seen[bet.Username] {
			continue
		}
		seen[bet.Username] = true

		msg := fmt.Sprintf(
			"Market \"%s\" has been resolved: %s",
			market.QuestionTitle,
			market.ResolutionResult,
		)
		notif := models.Notification{
			Username: bet.Username,
			Type:     "market_resolved",
			MarketID: market.ID,
			Message:  msg,
			IsRead:   false,
		}
		if err := db.Create(&notif).Error; err != nil {
			return fmt.Errorf("create notification for %s: %w", bet.Username, err)
		}
	}
	return nil
}
