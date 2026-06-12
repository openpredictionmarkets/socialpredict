package marketshandlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"socialpredict/handlers"
	"socialpredict/handlers/markets/dto"
	dmarkets "socialpredict/internal/domain/markets"

	"github.com/gorilla/mux"
)

type marketGroupService interface {
	CreateMarketGroup(ctx context.Context, req dmarkets.MarketGroupCreateRequest, creatorUsername string) (*dmarkets.MarketGroup, error)
	GetMarketGroupOverview(ctx context.Context, groupID int64) (*dmarkets.MarketGroupOverview, error)
}

// CreateMarketGroup handles POST /v0/market-groups.
func (h *Handler) CreateMarketGroup(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeMethodNotAllowed(w)
		return
	}
	if h.auth == nil {
		writeInternalError(w)
		return
	}

	svc, ok := h.service.(marketGroupService)
	if !ok {
		writeInternalError(w)
		return
	}

	user, authErr := h.auth.CurrentUser(r)
	if authErr != nil {
		writeAuthError(w, authErr)
		return
	}

	var req dto.CreateMarketGroupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeInvalidRequest(w)
		return
	}

	sanitizedReq, err := h.sanitizeMarketGroupRequest(req)
	if err != nil {
		writeInvalidRequest(w)
		return
	}

	group, err := svc.CreateMarketGroup(r.Context(), dmarkets.MarketGroupCreateRequest{
		QuestionTitle:      sanitizedReq.QuestionTitle,
		Description:        sanitizedReq.Description,
		ResolutionDateTime: sanitizedReq.ResolutionDateTime,
		AnswerLabels:       sanitizedReq.AnswerLabels,
		TagSlugs:           sanitizedReq.TagSlugs,
	}, user.Username)
	if err != nil {
		writeCreateError(w, err)
		return
	}

	overview, err := svc.GetMarketGroupOverview(r.Context(), group.ID)
	if err != nil {
		response := dto.MarketGroupDetailsResponse{
			Group:   marketGroupToResponse(group),
			Creator: &dto.CreatorResponse{Username: user.Username},
			Answers: []dto.MarketGroupAnswerResponse{},
		}
		_ = writeJSON(w, http.StatusCreated, response)
		return
	}

	_ = writeJSON(w, http.StatusCreated, marketGroupOverviewToResponse(overview))
}

// GetMarketGroup handles GET /v0/market-groups/{id}.
func (h *Handler) GetMarketGroup(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeMethodNotAllowed(w)
		return
	}

	svc, ok := h.service.(marketGroupService)
	if !ok {
		writeInternalError(w)
		return
	}

	id, err := parseMarketGroupIDFromRequest(r)
	if err != nil {
		writeInvalidRequest(w)
		return
	}

	overview, err := svc.GetMarketGroupOverview(r.Context(), id)
	if err != nil {
		writeMarketGroupDetailsError(w, err)
		return
	}

	_ = writeJSON(w, http.StatusOK, marketGroupOverviewToResponse(overview))
}

func (h *Handler) sanitizeMarketGroupRequest(req dto.CreateMarketGroupRequest) (dto.CreateMarketGroupRequest, error) {
	if h.securityService == nil || h.securityService.Sanitizer == nil {
		return dto.CreateMarketGroupRequest{}, errors.New("security service unavailable")
	}

	baseReq := dto.CreateMarketRequest{
		QuestionTitle:      req.QuestionTitle,
		Description:        req.Description,
		ResolutionDateTime: req.ResolutionDateTime,
	}
	sanitizedBase, err := sanitizeMarketRequest(h.securityService, baseReq)
	if err != nil {
		return dto.CreateMarketGroupRequest{}, err
	}

	answerLabels := make([]string, 0, len(req.AnswerLabels))
	for _, rawLabel := range req.AnswerLabels {
		label := strings.TrimSpace(rawLabel)
		sanitizedLabel, err := h.securityService.Sanitizer.SanitizeMarketTitle(label)
		if err != nil {
			return dto.CreateMarketGroupRequest{}, err
		}
		answerLabels = append(answerLabels, sanitizedLabel)
	}

	req.QuestionTitle = sanitizedBase.QuestionTitle
	req.Description = sanitizedBase.Description
	req.AnswerLabels = answerLabels
	return req, nil
}

func parseMarketGroupIDFromRequest(r *http.Request) (int64, error) {
	idStr := mux.Vars(r)["id"]
	if idStr == "" {
		return 0, errors.New("missing market group id")
	}
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		return 0, errors.New("invalid market group id")
	}
	return id, nil
}

func marketGroupOverviewToResponse(overview *dmarkets.MarketGroupOverview) dto.MarketGroupDetailsResponse {
	if overview == nil {
		return dto.MarketGroupDetailsResponse{
			Answers: []dto.MarketGroupAnswerResponse{},
		}
	}

	answers := make([]dto.MarketGroupAnswerResponse, 0, len(overview.Answers))
	for _, answer := range overview.Answers {
		answers = append(answers, dto.MarketGroupAnswerResponse{
			ID:           answer.Member.ID,
			GroupID:      answer.Member.GroupID,
			MarketID:     answer.Member.MarketID,
			AnswerLabel:  answer.Member.AnswerLabel,
			DisplayOrder: answer.Member.DisplayOrder,
			Market:       marketOverviewToResponse(answer.Overview),
		})
	}

	return dto.MarketGroupDetailsResponse{
		Group:   marketGroupToResponse(overview.Group),
		Creator: creatorResponseFromSummary(overview.Creator),
		Answers: answers,
	}
}

func marketGroupToResponse(group *dmarkets.MarketGroup) *dto.MarketGroupResponse {
	if group == nil {
		return &dto.MarketGroupResponse{}
	}
	status := group.LifecycleStatus
	if strings.EqualFold(status, dmarkets.MarketLifecyclePublished) {
		status = dmarkets.MarketStatusActive
	}
	return &dto.MarketGroupResponse{
		ID:                 group.ID,
		QuestionTitle:      group.QuestionTitle,
		Description:        group.Description,
		GroupType:          group.GroupType,
		ProbabilityPolicy:  group.ProbabilityPolicy,
		ResolutionPolicy:   group.ResolutionPolicy,
		LifecycleStatus:    group.LifecycleStatus,
		Status:             status,
		ProposalCost:       group.ProposalCost,
		CreatorUsername:    group.CreatorUsername,
		StewardUsername:    group.StewardUsername,
		ApprovedBy:         group.ApprovedBy,
		ApprovedAt:         group.ApprovedAt,
		RejectedBy:         group.RejectedBy,
		RejectedAt:         group.RejectedAt,
		RejectionReason:    group.RejectionReason,
		ResolutionDateTime: group.ResolutionDateTime,
		CreatedAt:          group.CreatedAt,
		UpdatedAt:          group.UpdatedAt,
		AnswerCount:        len(group.Members),
	}
}

func writeMarketGroupDetailsError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, dmarkets.ErrMarketGroupNotFound):
		_ = handlers.WriteFailure(w, http.StatusNotFound, handlers.ReasonMarketNotFound)
	case errors.Is(err, dmarkets.ErrInvalidInput):
		_ = handlers.WriteFailure(w, http.StatusBadRequest, handlers.ReasonValidationFailed)
	default:
		writeDetailsError(w, err)
	}
}
