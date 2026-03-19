package comments

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"socialpredict/middleware"
	"socialpredict/models"
	"socialpredict/util"
)

const maxCommentLength = 2000

type commentRequest struct {
	Content string `json:"content"`
}

// CommentResponse is the public representation of a comment returned by the API.
type CommentResponse struct {
	ID        uint      `json:"id"`
	MarketID  uint      `json:"marketId"`
	Username  string    `json:"username"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"createdAt"`
}

// ListCommentsHandler returns all comments for a given market, oldest first.
func ListCommentsHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	marketID, err := strconv.ParseUint(vars["marketId"], 10, 64)
	if err != nil {
		http.Error(w, "invalid market id", http.StatusBadRequest)
		return
	}

	db := util.GetDB()
	var comments []models.Comment
	if err := db.Where("market_id = ?", uint(marketID)).Order("created_at asc").Find(&comments).Error; err != nil {
		http.Error(w, "failed to fetch comments", http.StatusInternalServerError)
		return
	}

	responses := make([]CommentResponse, len(comments))
	for i, c := range comments {
		responses[i] = CommentResponse{
			ID:        c.ID,
			MarketID:  c.MarketID,
			Username:  c.Username,
			Content:   c.Content,
			CreatedAt: c.CreatedAt,
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(responses)
}

// CreateCommentHandler adds a new comment to a market. Requires authentication.
func CreateCommentHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	marketID, err := strconv.ParseUint(vars["marketId"], 10, 64)
	if err != nil {
		http.Error(w, "invalid market id", http.StatusBadRequest)
		return
	}

	db := util.GetDB()
	user, httpErr := middleware.ValidateUserAndEnforcePasswordChangeGetUser(r, db)
	if httpErr != nil {
		http.Error(w, httpErr.Message, httpErr.StatusCode)
		return
	}

	var req commentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	content := strings.TrimSpace(req.Content)
	if content == "" {
		http.Error(w, "comment content cannot be empty", http.StatusBadRequest)
		return
	}
	if len([]rune(content)) > maxCommentLength {
		http.Error(w, "comment content too long (max 2000 characters)", http.StatusBadRequest)
		return
	}

	// Verify the market exists
	var market models.Market
	if err := db.First(&market, uint(marketID)).Error; err != nil {
		http.Error(w, "market not found", http.StatusNotFound)
		return
	}

	comment := models.Comment{
		MarketID: uint(marketID),
		Username: user.Username,
		Content:  content,
	}
	if err := db.Create(&comment).Error; err != nil {
		http.Error(w, "failed to create comment", http.StatusInternalServerError)
		return
	}

	resp := CommentResponse{
		ID:        comment.ID,
		MarketID:  comment.MarketID,
		Username:  comment.Username,
		Content:   comment.Content,
		CreatedAt: comment.CreatedAt,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp)
}

// DeleteCommentHandler removes a comment. Only the comment author may delete it.
func DeleteCommentHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	commentID, err := strconv.ParseUint(vars["commentId"], 10, 64)
	if err != nil {
		http.Error(w, "invalid comment id", http.StatusBadRequest)
		return
	}

	db := util.GetDB()
	user, httpErr := middleware.ValidateUserAndEnforcePasswordChangeGetUser(r, db)
	if httpErr != nil {
		http.Error(w, httpErr.Message, httpErr.StatusCode)
		return
	}

	var comment models.Comment
	if err := db.First(&comment, uint(commentID)).Error; err != nil {
		http.Error(w, "comment not found", http.StatusNotFound)
		return
	}

	if comment.Username != user.Username {
		http.Error(w, "forbidden: you can only delete your own comments", http.StatusForbidden)
		return
	}

	if err := db.Delete(&comment).Error; err != nil {
		http.Error(w, "failed to delete comment", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
