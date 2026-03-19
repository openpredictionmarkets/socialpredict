package pollshandlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"socialpredict/middleware"
	"socialpredict/models"
	"socialpredict/util"

	"github.com/gorilla/mux"
)

const maxPollQuestionLength = 200
const maxPollDescriptionLength = 2000

type PollResponse struct {
	ID              uint       `json:"id"`
	CreatorUsername string     `json:"creatorUsername"`
	Question        string     `json:"question"`
	Description     string     `json:"description"`
	IsClosed        bool       `json:"isClosed"`
	ClosedAt        *time.Time `json:"closedAt"`
	YesCount        int64      `json:"yesCount"`
	NoCount         int64      `json:"noCount"`
	CreatedAt       time.Time  `json:"createdAt"`
	UserVote        string     `json:"userVote"` // "YES", "NO", or "" if not voted / unauthenticated
}

// loadPollResponse enriches a Poll with vote counts and the calling user's vote (if any).
func loadPollResponse(poll models.Poll, username string) PollResponse {
	db := util.GetDB()
	var yesCount, noCount int64
	db.Model(&models.PollVote{}).Where("poll_id = ? AND vote = ?", poll.ID, "YES").Count(&yesCount)
	db.Model(&models.PollVote{}).Where("poll_id = ? AND vote = ?", poll.ID, "NO").Count(&noCount)

	userVote := ""
	if username != "" {
		var vote models.PollVote
		if err := db.Where("poll_id = ? AND username = ?", poll.ID, username).First(&vote).Error; err == nil {
			userVote = vote.Vote
		}
	}

	return PollResponse{
		ID:              poll.ID,
		CreatorUsername: poll.CreatorUsername,
		Question:        poll.Question,
		Description:     poll.Description,
		IsClosed:        poll.IsClosed,
		ClosedAt:        poll.ClosedAt,
		YesCount:        yesCount,
		NoCount:         noCount,
		CreatedAt:       poll.CreatedAt,
		UserVote:        userVote,
	}
}

// optionalUsername extracts username from JWT if present, without requiring auth.
func optionalUsername(r *http.Request) string {
	db := util.GetDB()
	user, err := middleware.ValidateTokenAndGetUser(r, db)
	if err != nil {
		return ""
	}
	return user.Username
}

// ListPollsHandler GET /v0/polls — public, returns all open polls newest-first.
func ListPollsHandler(w http.ResponseWriter, r *http.Request) {
	db := util.GetDB()
	var polls []models.Poll
	if err := db.Where("is_closed = ?", false).Order("created_at desc").Find(&polls).Error; err != nil {
		http.Error(w, "failed to fetch polls", http.StatusInternalServerError)
		return
	}

	username := optionalUsername(r)
	responses := make([]PollResponse, 0, len(polls))
	for _, p := range polls {
		responses = append(responses, loadPollResponse(p, username))
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(responses)
}

// GetPollHandler GET /v0/polls/{pollId} — public.
func GetPollHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseUint(vars["pollId"], 10, 64)
	if err != nil {
		http.Error(w, "invalid poll id", http.StatusBadRequest)
		return
	}

	db := util.GetDB()
	var poll models.Poll
	if err := db.First(&poll, uint(id)).Error; err != nil {
		http.Error(w, "poll not found", http.StatusNotFound)
		return
	}

	username := optionalUsername(r)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(loadPollResponse(poll, username))
}

// CreatePollHandler POST /v0/polls — auth required.
func CreatePollHandler(w http.ResponseWriter, r *http.Request) {
	db := util.GetDB()
	user, httperr := middleware.ValidateUserAndEnforcePasswordChangeGetUser(r, db)
	if httperr != nil {
		http.Error(w, httperr.Error(), httperr.StatusCode)
		return
	}

	var input struct {
		Question    string `json:"question"`
		Description string `json:"description"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	input.Question = strings.TrimSpace(input.Question)
	if utf8.RuneCountInString(input.Question) < 1 || utf8.RuneCountInString(input.Question) > maxPollQuestionLength {
		http.Error(w, "question must be 1–200 characters", http.StatusBadRequest)
		return
	}
	if utf8.RuneCountInString(input.Description) > maxPollDescriptionLength {
		http.Error(w, "description exceeds 2000 characters", http.StatusBadRequest)
		return
	}

	poll := models.Poll{
		CreatorUsername: user.Username,
		Question:        input.Question,
		Description:     strings.TrimSpace(input.Description),
	}
	if err := db.Create(&poll).Error; err != nil {
		http.Error(w, "failed to create poll", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(loadPollResponse(poll, user.Username))
}

// VotePollHandler POST /v0/polls/{pollId}/vote — auth required, one vote per user.
func VotePollHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseUint(vars["pollId"], 10, 64)
	if err != nil {
		http.Error(w, "invalid poll id", http.StatusBadRequest)
		return
	}

	db := util.GetDB()
	user, httperr := middleware.ValidateUserAndEnforcePasswordChangeGetUser(r, db)
	if httperr != nil {
		http.Error(w, httperr.Error(), httperr.StatusCode)
		return
	}

	var poll models.Poll
	if err := db.First(&poll, uint(id)).Error; err != nil {
		http.Error(w, "poll not found", http.StatusNotFound)
		return
	}
	if poll.IsClosed {
		http.Error(w, "poll is closed", http.StatusConflict)
		return
	}

	// Check for existing vote
	var existing models.PollVote
	if err := db.Where("poll_id = ? AND username = ?", poll.ID, user.Username).First(&existing).Error; err == nil {
		http.Error(w, "already voted on this poll", http.StatusConflict)
		return
	}

	var input struct {
		Vote string `json:"vote"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	input.Vote = strings.ToUpper(strings.TrimSpace(input.Vote))
	if input.Vote != "YES" && input.Vote != "NO" {
		http.Error(w, "vote must be YES or NO", http.StatusBadRequest)
		return
	}

	vote := models.PollVote{
		PollID:   poll.ID,
		Username: user.Username,
		Vote:     input.Vote,
	}
	if err := db.Create(&vote).Error; err != nil {
		http.Error(w, "failed to record vote", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(loadPollResponse(poll, user.Username))
}

// ClosePollHandler POST /v0/polls/{pollId}/close — creator or admin only.
func ClosePollHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseUint(vars["pollId"], 10, 64)
	if err != nil {
		http.Error(w, "invalid poll id", http.StatusBadRequest)
		return
	}

	db := util.GetDB()
	user, httperr := middleware.ValidateUserAndEnforcePasswordChangeGetUser(r, db)
	if httperr != nil {
		http.Error(w, httperr.Error(), httperr.StatusCode)
		return
	}

	var poll models.Poll
	if err := db.First(&poll, uint(id)).Error; err != nil {
		http.Error(w, "poll not found", http.StatusNotFound)
		return
	}

	if poll.CreatorUsername != user.Username && user.UserType != "ADMIN" {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}
	if poll.IsClosed {
		http.Error(w, "poll already closed", http.StatusConflict)
		return
	}

	now := time.Now().UTC()
	poll.IsClosed = true
	poll.ClosedAt = &now
	if err := db.Save(&poll).Error; err != nil {
		http.Error(w, "failed to close poll", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(loadPollResponse(poll, user.Username))
}
