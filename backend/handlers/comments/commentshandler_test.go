package comments

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/gorilla/mux"
	"socialpredict/models"
	"socialpredict/models/modelstesting"
	"socialpredict/util"
)

func setup(t *testing.T) {
	t.Helper()
	db := modelstesting.NewFakeDB(t)
	origDB := util.DB
	util.DB = db
	t.Cleanup(func() { util.DB = origDB })
	t.Setenv("JWT_SIGNING_KEY", "test-secret-key-for-testing")
}

func getDB() interface{ Create(interface{}) interface{ Error error } } {
	return nil // not used directly, only via util.DB
}

// seedUserAndMarket creates a user and market in the current util.DB.
func seedUserAndMarket(t *testing.T, username string) (models.User, models.Market) {
	t.Helper()
	db := util.DB

	user := modelstesting.GenerateUser(username, 100)
	user.MustChangePassword = false
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("create user %s: %v", username, err)
	}

	market := modelstesting.GenerateMarket(1, username)
	if err := db.Create(&market).Error; err != nil {
		t.Fatalf("create market: %v", err)
	}

	return user, market
}

func TestListComments_EmptyMarket(t *testing.T) {
	setup(t)
	_, market := seedUserAndMarket(t, "alice")

	req := httptest.NewRequest("GET", "/v0/markets/"+strconv.Itoa(int(market.ID))+"/comments", nil)
	req = mux.SetURLVars(req, map[string]string{"marketId": strconv.Itoa(int(market.ID))})
	rec := httptest.NewRecorder()

	ListCommentsHandler(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var result []CommentResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &result); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(result) != 0 {
		t.Fatalf("expected empty slice, got %d comments", len(result))
	}
}

func TestCreateComment_Success(t *testing.T) {
	setup(t)
	user, market := seedUserAndMarket(t, "alice")
	_ = user

	body, _ := json.Marshal(commentRequest{Content: "Great prediction!"})
	req := httptest.NewRequest("POST", "/v0/markets/1/comments", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+modelstesting.GenerateValidJWT("alice"))
	req = mux.SetURLVars(req, map[string]string{"marketId": strconv.Itoa(int(market.ID))})
	rec := httptest.NewRecorder()

	CreateCommentHandler(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body.String())
	}
	var resp CommentResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp.Content != "Great prediction!" {
		t.Fatalf("expected content %q, got %q", "Great prediction!", resp.Content)
	}
	if resp.Username != "alice" {
		t.Fatalf("expected username alice, got %s", resp.Username)
	}
	if resp.MarketID != uint(market.ID) {
		t.Fatalf("expected marketId %d, got %d", market.ID, resp.MarketID)
	}
}

func TestCreateComment_EmptyContent(t *testing.T) {
	setup(t)
	_, market := seedUserAndMarket(t, "alice")

	body, _ := json.Marshal(commentRequest{Content: "   "})
	req := httptest.NewRequest("POST", "/v0/markets/1/comments", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+modelstesting.GenerateValidJWT("alice"))
	req = mux.SetURLVars(req, map[string]string{"marketId": strconv.Itoa(int(market.ID))})
	rec := httptest.NewRecorder()

	CreateCommentHandler(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for empty content, got %d", rec.Code)
	}
}

func TestCreateComment_Unauthenticated(t *testing.T) {
	setup(t)
	_, market := seedUserAndMarket(t, "alice")

	body, _ := json.Marshal(commentRequest{Content: "Hello"})
	req := httptest.NewRequest("POST", "/v0/markets/1/comments", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = mux.SetURLVars(req, map[string]string{"marketId": strconv.Itoa(int(market.ID))})
	rec := httptest.NewRecorder()

	CreateCommentHandler(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}

func TestDeleteComment_OwnComment(t *testing.T) {
	setup(t)
	_, market := seedUserAndMarket(t, "alice")

	comment := models.Comment{
		MarketID: uint(market.ID),
		Username: "alice",
		Content:  "To be deleted",
	}
	if err := util.DB.Create(&comment).Error; err != nil {
		t.Fatalf("seed comment: %v", err)
	}

	req := httptest.NewRequest("DELETE", "/v0/markets/1/comments/"+strconv.Itoa(int(comment.ID)), nil)
	req.Header.Set("Authorization", "Bearer "+modelstesting.GenerateValidJWT("alice"))
	req = mux.SetURLVars(req, map[string]string{
		"marketId":  strconv.Itoa(int(market.ID)),
		"commentId": strconv.Itoa(int(comment.ID)),
	})
	rec := httptest.NewRecorder()

	DeleteCommentHandler(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d: %s", rec.Code, rec.Body.String())
	}

	// Verify soft-deleted in DB
	var check models.Comment
	if err := util.DB.Unscoped().First(&check, comment.ID).Error; err != nil {
		t.Fatalf("fetch deleted comment: %v", err)
	}
	if check.DeletedAt.Time.IsZero() {
		t.Fatal("expected comment to be soft-deleted")
	}
}

func TestDeleteComment_OtherUserForbidden(t *testing.T) {
	setup(t)
	_, market := seedUserAndMarket(t, "alice")

	bob := modelstesting.GenerateUser("bob", 0)
	bob.MustChangePassword = false
	if err := util.DB.Create(&bob).Error; err != nil {
		t.Fatalf("create bob: %v", err)
	}

	comment := models.Comment{
		MarketID: uint(market.ID),
		Username: "alice",
		Content:  "Alice's comment",
	}
	if err := util.DB.Create(&comment).Error; err != nil {
		t.Fatalf("seed comment: %v", err)
	}

	req := httptest.NewRequest("DELETE", "/v0/markets/1/comments/"+strconv.Itoa(int(comment.ID)), nil)
	req.Header.Set("Authorization", "Bearer "+modelstesting.GenerateValidJWT("bob"))
	req = mux.SetURLVars(req, map[string]string{
		"marketId":  strconv.Itoa(int(market.ID)),
		"commentId": strconv.Itoa(int(comment.ID)),
	})
	rec := httptest.NewRecorder()

	DeleteCommentHandler(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", rec.Code)
	}
}

func TestListComments_ReturnsSavedComments(t *testing.T) {
	setup(t)
	_, market := seedUserAndMarket(t, "alice")

	comments := []models.Comment{
		{MarketID: uint(market.ID), Username: "alice", Content: "First"},
		{MarketID: uint(market.ID), Username: "alice", Content: "Second"},
	}
	for i := range comments {
		if err := util.DB.Create(&comments[i]).Error; err != nil {
			t.Fatalf("seed comment %d: %v", i, err)
		}
	}

	req := httptest.NewRequest("GET", "/v0/markets/1/comments", nil)
	req = mux.SetURLVars(req, map[string]string{"marketId": strconv.Itoa(int(market.ID))})
	rec := httptest.NewRecorder()

	ListCommentsHandler(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var result []CommentResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &result); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(result) != 2 {
		t.Fatalf("expected 2 comments, got %d", len(result))
	}
	if result[0].Content != "First" {
		t.Fatalf("expected first comment to be 'First', got %q", result[0].Content)
	}
}
