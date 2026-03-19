package adminhandlers

import (
	"encoding/csv"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"socialpredict/middleware"
	"socialpredict/models"
	"socialpredict/util"
)

// ExportBetsCSVHandler streams all bets as CSV. Admin-only.
func ExportBetsCSVHandler(w http.ResponseWriter, r *http.Request) {
	db := util.GetDB()
	if err := middleware.ValidateAdminToken(r, db); err != nil {
		http.Error(w, "Forbidden: "+err.Error(), http.StatusForbidden)
		return
	}

	var bets []models.Bet
	if err := db.Order("id asc").Find(&bets).Error; err != nil {
		http.Error(w, "failed to fetch bets", http.StatusInternalServerError)
		return
	}

	filename := fmt.Sprintf("bets_%s.csv", time.Now().UTC().Format("20060102_150405"))
	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", `attachment; filename="`+filename+`"`)

	wr := csv.NewWriter(w)
	_ = wr.Write([]string{"id", "username", "market_id", "action", "outcome", "amount", "placed_at"})
	for _, b := range bets {
		action := "BUY"
		if b.Amount < 0 {
			action = "SELL"
		}
		_ = wr.Write([]string{
			strconv.FormatUint(uint64(b.ID), 10),
			b.Username,
			strconv.FormatUint(uint64(b.MarketID), 10),
			action,
			b.Outcome,
			strconv.FormatInt(b.Amount, 10),
			b.PlacedAt.UTC().Format(time.RFC3339),
		})
	}
	wr.Flush()
}

// ExportMarketsCSVHandler streams all markets as CSV. Admin-only.
func ExportMarketsCSVHandler(w http.ResponseWriter, r *http.Request) {
	db := util.GetDB()
	if err := middleware.ValidateAdminToken(r, db); err != nil {
		http.Error(w, "Forbidden: "+err.Error(), http.StatusForbidden)
		return
	}

	var markets []models.Market
	if err := db.Order("id asc").Find(&markets).Error; err != nil {
		http.Error(w, "failed to fetch markets", http.StatusInternalServerError)
		return
	}

	filename := fmt.Sprintf("markets_%s.csv", time.Now().UTC().Format("20060102_150405"))
	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", `attachment; filename="`+filename+`"`)

	wr := csv.NewWriter(w)
	_ = wr.Write([]string{
		"id", "creator_username", "question_title", "description",
		"outcome_type", "is_resolved", "resolution_result",
		"initial_probability", "resolution_date_time", "final_resolution_date_time",
	})
	for _, m := range markets {
		_ = wr.Write([]string{
			strconv.FormatInt(m.ID, 10),
			m.CreatorUsername,
			m.QuestionTitle,
			m.Description,
			m.OutcomeType,
			strconv.FormatBool(m.IsResolved),
			m.ResolutionResult,
			fmt.Sprintf("%.4f", m.InitialProbability),
			m.ResolutionDateTime.UTC().Format(time.RFC3339),
			m.FinalResolutionDateTime.UTC().Format(time.RFC3339),
		})
	}
	wr.Flush()
}

// ExportUsersCSVHandler streams public user data as CSV. Admin-only. Passwords excluded.
func ExportUsersCSVHandler(w http.ResponseWriter, r *http.Request) {
	db := util.GetDB()
	if err := middleware.ValidateAdminToken(r, db); err != nil {
		http.Error(w, "Forbidden: "+err.Error(), http.StatusForbidden)
		return
	}

	var users []models.User
	if err := db.Order("id asc").Find(&users).Error; err != nil {
		http.Error(w, "failed to fetch users", http.StatusInternalServerError)
		return
	}

	filename := fmt.Sprintf("users_%s.csv", time.Now().UTC().Format("20060102_150405"))
	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", `attachment; filename="`+filename+`"`)

	wr := csv.NewWriter(w)
	_ = wr.Write([]string{
		"username", "display_name", "user_type",
		"account_balance", "initial_account_balance",
		"created_at",
	})
	for _, u := range users {
		_ = wr.Write([]string{
			u.Username,
			u.DisplayName,
			u.UserType,
			strconv.FormatInt(u.AccountBalance, 10),
			strconv.FormatInt(u.InitialAccountBalance, 10),
			u.CreatedAt.UTC().Format(time.RFC3339),
		})
	}
	wr.Flush()
}
