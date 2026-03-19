package pollshandlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"

	"socialpredict/models"
	"socialpredict/models/modelstesting"
	"socialpredict/util"
)

func setupPollTest(t *testing.T) string {
	t.Helper()
	db := modelstesting.NewFakeDB(t)
	origDB := util.DB
	util.DB = db
	t.Cleanup(func() { util.DB = origDB })
	t.Setenv("JWT_SIGNING_KEY", "test-secret-key-for-testing")

	user := modelstesting.GenerateUser("alice", 100)
	user.MustChangePassword = false
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}
	return modelstesting.GenerateValidJWT("alice")
}

func TestCreatePoll_Success(t *testing.T) {
	jwt := setupPollTest(t)

	body, _ := json.Marshal(map[string]string{
		"question":    "Will it rain tomorrow?",
		"description": "Simple weather poll",
	})
	req := httptest.NewRequest("POST", "/v0/polls", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+jwt)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	CreatePollHandler(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body.String())
	}
	var resp PollResponse
	json.NewDecoder(rec.Body).Decode(&resp)
	if resp.Question != "Will it rain tomorrow?" {
		t.Errorf("expected question in response, got %q", resp.Question)
	}
	if resp.CreatorUsername != "alice" {
		t.Errorf("expected creator alice, got %q", resp.CreatorUsername)
	}
}

func TestCreatePoll_EmptyQuestion(t *testing.T) {
	jwt := setupPollTest(t)

	body, _ := json.Marshal(map[string]string{"question": ""})
	req := httptest.NewRequest("POST", "/v0/polls", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+jwt)
	rec := httptest.NewRecorder()

	CreatePollHandler(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for empty question, got %d", rec.Code)
	}
}

func TestCreatePoll_Unauthenticated(t *testing.T) {
	setupPollTest(t)

	body, _ := json.Marshal(map[string]string{"question": "Hello?"})
	req := httptest.NewRequest("POST", "/v0/polls", bytes.NewReader(body))
	rec := httptest.NewRecorder()

	CreatePollHandler(rec, req)

	if rec.Code != http.StatusForbidden && rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401/403 for unauthenticated, got %d", rec.Code)
	}
}

func TestListPolls_ReturnsOpenPolls(t *testing.T) {
	jwt := setupPollTest(t)

	// Create two polls
	poll1 := models.Poll{CreatorUsername: "alice", Question: "Poll 1?"}
	poll2 := models.Poll{CreatorUsername: "alice", Question: "Poll 2?", IsClosed: true}
	util.DB.Create(&poll1)
	util.DB.Create(&poll2)

	req := httptest.NewRequest("GET", "/v0/polls", nil)
	req.Header.Set("Authorization", "Bearer "+jwt)
	rec := httptest.NewRecorder()

	ListPollsHandler(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var resp []PollResponse
	json.NewDecoder(rec.Body).Decode(&resp)
	if len(resp) != 1 {
		t.Errorf("expected 1 open poll, got %d", len(resp))
	}
	if resp[0].Question != "Poll 1?" {
		t.Errorf("expected Poll 1?, got %q", resp[0].Question)
	}
}

func TestVotePoll_Success(t *testing.T) {
	jwt := setupPollTest(t)

	poll := models.Poll{CreatorUsername: "alice", Question: "Yes or no?"}
	util.DB.Create(&poll)

	body, _ := json.Marshal(map[string]string{"vote": "YES"})
	req := httptest.NewRequest("POST", "/v0/polls/1/vote", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+jwt)
	req = mux.SetURLVars(req, map[string]string{"pollId": "1"})
	rec := httptest.NewRecorder()

	VotePollHandler(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var resp PollResponse
	json.NewDecoder(rec.Body).Decode(&resp)
	if resp.YesCount != 1 {
		t.Errorf("expected YesCount=1, got %d", resp.YesCount)
	}
	if resp.UserVote != "YES" {
		t.Errorf("expected UserVote=YES, got %q", resp.UserVote)
	}
}

func TestVotePoll_DoubleVoteForbidden(t *testing.T) {
	jwt := setupPollTest(t)

	poll := models.Poll{CreatorUsername: "alice", Question: "Double?"}
	util.DB.Create(&poll)
	util.DB.Create(&models.PollVote{PollID: poll.ID, Username: "alice", Vote: "YES"})

	body, _ := json.Marshal(map[string]string{"vote": "NO"})
	req := httptest.NewRequest("POST", "/v0/polls/1/vote", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+jwt)
	req = mux.SetURLVars(req, map[string]string{"pollId": "1"})
	rec := httptest.NewRecorder()

	VotePollHandler(rec, req)

	if rec.Code != http.StatusConflict {
		t.Fatalf("expected 409 for double vote, got %d", rec.Code)
	}
}

func TestVotePoll_InvalidVoteValue(t *testing.T) {
	jwt := setupPollTest(t)

	poll := models.Poll{CreatorUsername: "alice", Question: "Valid?"}
	util.DB.Create(&poll)

	body, _ := json.Marshal(map[string]string{"vote": "MAYBE"})
	req := httptest.NewRequest("POST", "/v0/polls/1/vote", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+jwt)
	req = mux.SetURLVars(req, map[string]string{"pollId": "1"})
	rec := httptest.NewRecorder()

	VotePollHandler(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for invalid vote, got %d", rec.Code)
	}
}

func TestVotePoll_ClosedPoll(t *testing.T) {
	jwt := setupPollTest(t)

	poll := models.Poll{CreatorUsername: "alice", Question: "Closed?", IsClosed: true}
	util.DB.Create(&poll)

	body, _ := json.Marshal(map[string]string{"vote": "YES"})
	req := httptest.NewRequest("POST", "/v0/polls/1/vote", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+jwt)
	req = mux.SetURLVars(req, map[string]string{"pollId": "1"})
	rec := httptest.NewRecorder()

	VotePollHandler(rec, req)

	if rec.Code != http.StatusConflict {
		t.Fatalf("expected 409 for closed poll, got %d", rec.Code)
	}
}

func TestClosePoll_ByCreator(t *testing.T) {
	jwt := setupPollTest(t)

	poll := models.Poll{CreatorUsername: "alice", Question: "Close me?"}
	util.DB.Create(&poll)

	req := httptest.NewRequest("POST", "/v0/polls/1/close", nil)
	req.Header.Set("Authorization", "Bearer "+jwt)
	req = mux.SetURLVars(req, map[string]string{"pollId": "1"})
	rec := httptest.NewRecorder()

	ClosePollHandler(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var resp PollResponse
	json.NewDecoder(rec.Body).Decode(&resp)
	if !resp.IsClosed {
		t.Error("expected IsClosed=true")
	}
}

func TestClosePoll_NotCreatorForbidden(t *testing.T) {
	setupPollTest(t)

	bob := modelstesting.GenerateUser("bob", 0)
	bob.MustChangePassword = false
	util.DB.Create(&bob)

	poll := models.Poll{CreatorUsername: "alice", Question: "Bob can't close me?"}
	util.DB.Create(&poll)

	req := httptest.NewRequest("POST", "/v0/polls/1/close", nil)
	req.Header.Set("Authorization", "Bearer "+modelstesting.GenerateValidJWT("bob"))
	req = mux.SetURLVars(req, map[string]string{"pollId": "1"})
	rec := httptest.NewRecorder()

	ClosePollHandler(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", rec.Code)
	}
}
