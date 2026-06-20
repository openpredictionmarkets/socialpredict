package marketshandlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"socialpredict/handlers"
	"socialpredict/handlers/markets/dto"
	dmarkets "socialpredict/internal/domain/markets"
	"socialpredict/internal/domain/readmodels"
	"socialpredict/logger"

	"github.com/gorilla/mux"
)

const groupedActivityLiveTargetFreshness = 0

type marketGroupService interface {
	CreateMarketGroup(ctx context.Context, req dmarkets.MarketGroupCreateRequest, creatorUsername string) (*dmarkets.MarketGroup, error)
	GetMarketGroupOverview(ctx context.Context, groupID int64) (*dmarkets.MarketGroupOverview, error)
	ResolveMarketGroup(ctx context.Context, groupID int64, req dmarkets.MarketGroupResolveRequest, username string) (*dmarkets.MarketGroup, error)
}

type marketGroupActivityService interface {
	GetMarketGroupBetsPage(ctx context.Context, groupID int64, p dmarkets.Page) (*dmarkets.MarketGroupBetsPage, error)
	GetMarketGroupPositionsPage(ctx context.Context, groupID int64, p dmarkets.Page) (*dmarkets.MarketGroupPositionsPage, error)
	GetMarketGroupLeaderboardPage(ctx context.Context, groupID int64, p dmarkets.Page) (*dmarkets.MarketGroupLeaderboardPage, error)
}

type marketGroupAnswerAdditionService interface {
	ProposeMarketGroupAnswerAddition(ctx context.Context, groupID int64, actorUsername string, req dmarkets.MarketGroupAnswerAdditionRequest) (*dmarkets.MarketGroupAnswerAddition, error)
	ListMarketGroupAnswerAdditionsForReviewer(ctx context.Context, reviewerUsername string, filters dmarkets.MarketGroupAnswerAdditionFilters) ([]dmarkets.MarketGroupAnswerAddition, error)
	ApproveMarketGroupAnswerAdditionForReviewer(ctx context.Context, additionID int64, actorUsername string, confirmed bool) (*dmarkets.MarketGroupAnswerAddition, error)
	RejectMarketGroupAnswerAdditionForReviewer(ctx context.Context, additionID int64, actorUsername string, reason string) (*dmarkets.MarketGroupAnswerAddition, error)
	UpdateMarketGroupAnswerAdditionSettings(ctx context.Context, groupID int64, actorUsername string, enabled bool) (*dmarkets.MarketGroup, error)
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
		QuestionTitle:              sanitizedReq.QuestionTitle,
		Description:                sanitizedReq.Description,
		ResolutionDateTime:         sanitizedReq.ResolutionDateTime,
		AnswerLabels:               sanitizedReq.AnswerLabels,
		TagSlugs:                   sanitizedReq.TagSlugs,
		AutoApproveAnswerAdditions: sanitizedReq.AutoApproveAnswerAdditions,
	}, user.Username)
	if err != nil {
		writeCreateError(w, err)
		return
	}

	h.invalidateCreatedMarketGroup(r.Context(), user.Username, group)

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

	_ = writeJSON(w, http.StatusCreated, marketGroupOverviewToResponse(r.Context(), h.service, overview))
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

	_ = writeJSON(w, http.StatusOK, marketGroupOverviewToResponse(r.Context(), h.service, overview))
}

// ResolveMarketGroup handles POST /v0/market-groups/{id}/resolve.
func (h *Handler) ResolveMarketGroup(w http.ResponseWriter, r *http.Request) {
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

	groupID, err := parseMarketGroupIDFromRequest(r)
	if err != nil {
		writeInvalidRequest(w)
		return
	}

	user, authErr := h.auth.CurrentUser(r)
	if authErr != nil {
		writeAuthError(w, authErr)
		return
	}

	var req dto.ResolveMarketGroupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeInvalidRequest(w)
		return
	}

	group, err := svc.ResolveMarketGroup(r.Context(), groupID, dmarkets.MarketGroupResolveRequest{
		Mode:            req.Mode,
		WinningMarketID: req.WinningMarketID,
		Resolutions:     groupChildResolutionsFromRequest(req.Resolutions),
	}, user.Username)
	if err != nil {
		writeResolveErrorResponse(w, err)
		return
	}

	if h.invalidator != nil && group != nil {
		for _, member := range group.Members {
			if err := h.invalidator.InvalidateAfterMarketTransaction(r.Context(), user.Username, member.MarketID, "market_group_resolved"); err != nil {
				logger.LogError("ResolveMarketGroup", "InvalidateReadModels", err)
			}
		}
	}

	_ = writeJSON(w, http.StatusOK, marketGroupToResponse(group))
}

// ProposeMarketGroupAnswerAddition handles POST /v0/market-groups/{id}/answers.
func (h *Handler) ProposeMarketGroupAnswerAddition(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeMethodNotAllowed(w)
		return
	}
	if h.auth == nil {
		writeInternalError(w)
		return
	}
	svc, ok := h.service.(marketGroupAnswerAdditionService)
	if !ok {
		writeInternalError(w)
		return
	}
	groupID, err := parseMarketGroupIDFromRequest(r)
	if err != nil {
		writeInvalidRequest(w)
		return
	}
	user, authErr := h.auth.CurrentUser(r)
	if authErr != nil {
		writeAuthError(w, authErr)
		return
	}
	var req dto.MarketGroupAnswerAdditionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeInvalidRequest(w)
		return
	}
	sanitizedLabel := strings.TrimSpace(req.AnswerLabel)
	if h.securityService != nil && h.securityService.Sanitizer != nil {
		sanitized, err := h.securityService.Sanitizer.SanitizeMarketTitle(sanitizedLabel)
		if err != nil {
			writeInvalidRequest(w)
			return
		}
		sanitizedLabel = sanitized
	}
	addition, err := svc.ProposeMarketGroupAnswerAddition(r.Context(), groupID, user.Username, dmarkets.MarketGroupAnswerAdditionRequest{
		AnswerLabel: sanitizedLabel,
	})
	if err != nil {
		writeMarketGroupDetailsError(w, err)
		return
	}
	if h.invalidator != nil && addition != nil && addition.MarketID > 0 {
		if err := h.invalidator.InvalidateAfterMarketTransaction(r.Context(), user.Username, addition.MarketID, "market_group_answer_added"); err != nil {
			logger.LogError("ProposeMarketGroupAnswerAddition", "InvalidateReadModels", err)
		}
	}
	_ = writeJSON(w, http.StatusCreated, marketGroupAnswerAdditionToResponse(addition))
}

// ListMarketGroupAnswerAdditionsForReview handles GET /v0/market-groups/{id}/answer-additions
// and GET /v0/profile/market-group-answer-additions.
func (h *Handler) ListMarketGroupAnswerAdditionsForReview(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeMethodNotAllowed(w)
		return
	}
	if h.auth == nil {
		writeInternalError(w)
		return
	}
	svc, ok := h.service.(marketGroupAnswerAdditionService)
	if !ok {
		writeInternalError(w)
		return
	}
	user, authErr := h.auth.CurrentUser(r)
	if authErr != nil {
		writeAuthError(w, authErr)
		return
	}
	filters, err := parseMarketGroupAnswerAdditionFiltersFromRequest(r)
	if err != nil {
		writeInvalidRequest(w)
		return
	}
	items, err := svc.ListMarketGroupAnswerAdditionsForReviewer(r.Context(), user.Username, filters)
	if err != nil {
		writeMarketGroupDetailsError(w, err)
		return
	}
	response := dto.MarketGroupAnswerAdditionsResponse{
		Additions: make([]dto.MarketGroupAnswerAdditionResponse, 0, len(items)),
		Total:     len(items),
	}
	for _, item := range items {
		itemCopy := item
		response.Additions = append(response.Additions, marketGroupAnswerAdditionToResponse(&itemCopy))
	}
	_ = writeJSON(w, http.StatusOK, response)
}

// ReviewMarketGroupAnswerAddition handles PATCH /v0/market-group-answer-additions/{additionId}.
func (h *Handler) ReviewMarketGroupAnswerAddition(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPatch {
		writeMethodNotAllowed(w)
		return
	}
	if h.auth == nil {
		writeInternalError(w)
		return
	}
	svc, ok := h.service.(marketGroupAnswerAdditionService)
	if !ok {
		writeInternalError(w)
		return
	}
	user, authErr := h.auth.CurrentUser(r)
	if authErr != nil {
		writeAuthError(w, authErr)
		return
	}
	additionID, err := parseMarketGroupAnswerAdditionIDFromRequest(r)
	if err != nil {
		writeInvalidRequest(w)
		return
	}
	var req dto.MarketGroupAnswerAdditionReviewRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeInvalidRequest(w)
		return
	}
	status := dmarkets.NormalizeMarketGroupAnswerAdditionStatus(req.Status)
	var addition *dmarkets.MarketGroupAnswerAddition
	switch status {
	case dmarkets.MarketGroupAnswerAdditionStatusApproved:
		addition, err = svc.ApproveMarketGroupAnswerAdditionForReviewer(r.Context(), additionID, user.Username, req.Confirm)
	case dmarkets.MarketGroupAnswerAdditionStatusRejected:
		addition, err = svc.RejectMarketGroupAnswerAdditionForReviewer(r.Context(), additionID, user.Username, req.Reason)
	default:
		writeInvalidRequest(w)
		return
	}
	if err != nil {
		writeMarketGroupDetailsError(w, err)
		return
	}
	if h.invalidator != nil && addition != nil && addition.MarketID > 0 {
		if err := h.invalidator.InvalidateAfterMarketTransaction(r.Context(), user.Username, addition.MarketID, "market_group_answer_reviewed"); err != nil {
			logger.LogError("ReviewMarketGroupAnswerAddition", "InvalidateReadModels", err)
		}
	}
	_ = writeJSON(w, http.StatusOK, marketGroupAnswerAdditionToResponse(addition))
}

// UpdateMarketGroupAnswerAdditionSettings handles PATCH /v0/market-groups/{id}/answer-addition-settings.
func (h *Handler) UpdateMarketGroupAnswerAdditionSettings(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPatch {
		writeMethodNotAllowed(w)
		return
	}
	if h.auth == nil {
		writeInternalError(w)
		return
	}
	svc, ok := h.service.(marketGroupAnswerAdditionService)
	if !ok {
		writeInternalError(w)
		return
	}
	groupID, err := parseMarketGroupIDFromRequest(r)
	if err != nil {
		writeInvalidRequest(w)
		return
	}
	user, authErr := h.auth.CurrentUser(r)
	if authErr != nil {
		writeAuthError(w, authErr)
		return
	}
	var req dto.MarketGroupAnswerAdditionSettingsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeInvalidRequest(w)
		return
	}
	group, err := svc.UpdateMarketGroupAnswerAdditionSettings(r.Context(), groupID, user.Username, req.AutoApproveAnswerAdditions)
	if err != nil {
		writeMarketGroupDetailsError(w, err)
		return
	}
	_ = writeJSON(w, http.StatusOK, marketGroupToResponse(group))
}

// MarketGroupBets handles GET /v0/market-groups/{id}/bets.
func (h *Handler) MarketGroupBets(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeMethodNotAllowed(w)
		return
	}
	svc, ok := h.service.(marketGroupActivityService)
	if !ok {
		writeInternalError(w)
		return
	}
	groupID, err := parseMarketGroupIDFromRequest(r)
	if err != nil {
		writeInvalidRequest(w)
		return
	}
	page := parsePagination(r, 20)
	result, err := svc.GetMarketGroupBetsPage(r.Context(), groupID, page)
	if err != nil {
		writeMarketGroupDetailsError(w, err)
		return
	}
	if err := handlers.WriteResult(w, http.StatusOK, marketGroupBetsPageToResponse(result)); err != nil {
		logger.LogError("MarketGroupBets", "WriteResponse", err)
	}
}

// MarketGroupPositions handles GET /v0/market-groups/{id}/positions.
func (h *Handler) MarketGroupPositions(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeMethodNotAllowed(w)
		return
	}
	svc, ok := h.service.(marketGroupActivityService)
	if !ok {
		writeInternalError(w)
		return
	}
	groupID, err := parseMarketGroupIDFromRequest(r)
	if err != nil {
		writeInvalidRequest(w)
		return
	}
	page := parsePagination(r, 20)
	result, err := svc.GetMarketGroupPositionsPage(r.Context(), groupID, page)
	if err != nil {
		writeMarketGroupDetailsError(w, err)
		return
	}
	if err := handlers.WriteResult(w, http.StatusOK, marketGroupPositionsPageToResponse(result)); err != nil {
		logger.LogError("MarketGroupPositions", "WriteResponse", err)
	}
}

// MarketGroupLeaderboard handles GET /v0/market-groups/{id}/leaderboard.
func (h *Handler) MarketGroupLeaderboard(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeMethodNotAllowed(w)
		return
	}
	svc, ok := h.service.(marketGroupActivityService)
	if !ok {
		writeInternalError(w)
		return
	}
	groupID, err := parseMarketGroupIDFromRequest(r)
	if err != nil {
		writeInvalidRequest(w)
		return
	}
	page := parsePagination(r, 20)
	result, err := svc.GetMarketGroupLeaderboardPage(r.Context(), groupID, page)
	if err != nil {
		writeMarketGroupDetailsError(w, err)
		return
	}
	if err := handlers.WriteResult(w, http.StatusOK, marketGroupLeaderboardPageToResponse(result)); err != nil {
		logger.LogError("MarketGroupLeaderboard", "WriteResponse", err)
	}
}

func (h *Handler) invalidateCreatedMarketGroup(ctx context.Context, username string, group *dmarkets.MarketGroup) {
	if h.invalidator == nil || group == nil {
		return
	}
	for _, member := range group.Members {
		if member.MarketID <= 0 {
			continue
		}
		if err := h.invalidator.InvalidateAfterMarketTransaction(ctx, username, member.MarketID, "market_group_created"); err != nil {
			logger.LogError("CreateMarketGroup", "InvalidateReadModels", err)
		}
	}
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

func parseMarketGroupAnswerAdditionIDFromRequest(r *http.Request) (int64, error) {
	idStr := mux.Vars(r)["additionId"]
	if idStr == "" {
		idStr = mux.Vars(r)["id"]
	}
	if idStr == "" {
		return 0, errors.New("missing market group answer addition id")
	}
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		return 0, errors.New("invalid market group answer addition id")
	}
	return id, nil
}

func parseMarketGroupAnswerAdditionFiltersFromRequest(r *http.Request) (dmarkets.MarketGroupAnswerAdditionFilters, error) {
	query := r.URL.Query()
	status := dmarkets.NormalizeMarketGroupAnswerAdditionStatus(query.Get("status"))
	if status != dmarkets.MarketGroupAnswerAdditionStatusPending &&
		status != dmarkets.MarketGroupAnswerAdditionStatusApproved &&
		status != dmarkets.MarketGroupAnswerAdditionStatusRejected {
		return dmarkets.MarketGroupAnswerAdditionFilters{}, errors.New("invalid answer addition status")
	}
	groupID := int64(0)
	if rawID := mux.Vars(r)["id"]; rawID != "" {
		parsed, err := strconv.ParseInt(rawID, 10, 64)
		if err != nil || parsed <= 0 {
			return dmarkets.MarketGroupAnswerAdditionFilters{}, errors.New("invalid market group id")
		}
		groupID = parsed
	}
	if rawGroupID := query.Get("groupId"); rawGroupID != "" {
		parsed, err := strconv.ParseInt(rawGroupID, 10, 64)
		if err != nil || parsed <= 0 {
			return dmarkets.MarketGroupAnswerAdditionFilters{}, errors.New("invalid market group id")
		}
		groupID = parsed
	}
	return dmarkets.MarketGroupAnswerAdditionFilters{
		GroupID: groupID,
		Status:  status,
		Limit:   boundedQueryInt(query.Get("limit"), 50, 1, 200),
		Offset:  boundedQueryInt(query.Get("offset"), 0, 0, 100000),
	}, nil
}

func marketGroupOverviewToResponse(ctx context.Context, provider any, overview *dmarkets.MarketGroupOverview) dto.MarketGroupDetailsResponse {
	if overview == nil {
		return dto.MarketGroupDetailsResponse{
			Answers: []dto.MarketGroupAnswerResponse{},
		}
	}

	answers := make([]dto.MarketGroupAnswerResponse, 0, len(overview.Answers))
	for _, answer := range overview.Answers {
		var probabilityChanges []dto.ProbabilityChangeResponse
		var descriptionAmendments []dto.MarketDescriptionAmendmentResponse
		if answer.Overview != nil {
			probabilityChanges = probabilityChangesToResponse(answer.Overview.ProbabilityChanges)
			descriptionAmendments = descriptionAmendmentsToResponse(answer.Overview.DescriptionAmendments)
		}
		answers = append(answers, dto.MarketGroupAnswerResponse{
			ID:                    answer.Member.ID,
			GroupID:               answer.Member.GroupID,
			MarketID:              answer.Member.MarketID,
			AnswerLabel:           answer.Member.AnswerLabel,
			DisplayOrder:          answer.Member.DisplayOrder,
			Market:                marketOverviewToResponse(ctx, provider, answer.Overview),
			ProbabilityChanges:    probabilityChanges,
			DescriptionAmendments: descriptionAmendments,
		})
	}

	return dto.MarketGroupDetailsResponse{
		Group:   marketGroupToResponse(overview.Group),
		Creator: creatorResponseFromSummary(overview.Creator),
		Answers: answers,
	}
}

func groupChildResolutionsFromRequest(items []dto.ResolveMarketGroupChildRequest) []dmarkets.MarketGroupChildResolution {
	resolutions := make([]dmarkets.MarketGroupChildResolution, 0, len(items))
	for _, item := range items {
		resolutions = append(resolutions, dmarkets.MarketGroupChildResolution{
			MarketID:   item.MarketID,
			Resolution: item.Resolution,
		})
	}
	return resolutions
}

func marketGroupBetsPageToResponse(page *dmarkets.MarketGroupBetsPage) dto.MarketGroupBetsResponse {
	if page == nil {
		return dto.MarketGroupBetsResponse{Bets: []dto.MarketGroupBetResponse{}}
	}
	rows := make([]dto.MarketGroupBetResponse, 0, len(page.Bets))
	for _, bet := range page.Bets {
		if bet == nil {
			continue
		}
		rows = append(rows, dto.MarketGroupBetResponse{
			AnswerMarketID: bet.AnswerMarketID,
			AnswerLabel:    bet.AnswerLabel,
			DisplayOrder:   bet.DisplayOrder,
			Username:       bet.Username,
			Outcome:        bet.Outcome,
			Amount:         bet.Amount,
			Probability:    bet.Probability,
			PlacedAt:       bet.PlacedAt,
		})
	}
	return dto.MarketGroupBetsResponse{
		GroupID:   page.GroupID,
		Bets:      rows,
		Total:     page.Total,
		Freshness: groupedActivityLiveFreshnessResponse(),
	}
}

func marketGroupPositionsPageToResponse(page *dmarkets.MarketGroupPositionsPage) dto.MarketGroupPositionsResponse {
	if page == nil {
		return dto.MarketGroupPositionsResponse{Positions: []dto.MarketGroupPositionResponse{}}
	}
	rows := make([]dto.MarketGroupPositionResponse, 0, len(page.Positions))
	for _, position := range page.Positions {
		if position == nil {
			continue
		}
		answers := make([]dto.MarketGroupPositionAnswerResponse, 0, len(position.Answers))
		for _, answer := range position.Answers {
			if answer == nil {
				continue
			}
			answers = append(answers, dto.MarketGroupPositionAnswerResponse{
				AnswerMarketID:   answer.AnswerMarketID,
				AnswerLabel:      answer.AnswerLabel,
				DisplayOrder:     answer.DisplayOrder,
				MarketID:         answer.MarketID,
				YesSharesOwned:   answer.YesSharesOwned,
				NoSharesOwned:    answer.NoSharesOwned,
				Value:            answer.Value,
				TotalSpent:       answer.TotalSpent,
				TotalSpentInPlay: answer.TotalSpentInPlay,
				IsResolved:       answer.IsResolved,
				ResolutionResult: answer.ResolutionResult,
			})
		}
		rows = append(rows, dto.MarketGroupPositionResponse{
			Username:         position.Username,
			YesSharesOwned:   position.YesSharesOwned,
			NoSharesOwned:    position.NoSharesOwned,
			Value:            position.Value,
			TotalSpent:       position.TotalSpent,
			TotalSpentInPlay: position.TotalSpentInPlay,
			Answers:          answers,
		})
	}
	return dto.MarketGroupPositionsResponse{
		GroupID:   page.GroupID,
		Positions: rows,
		Total:     page.Total,
		Freshness: groupedActivityLiveFreshnessResponse(),
	}
}

func marketGroupLeaderboardPageToResponse(page *dmarkets.MarketGroupLeaderboardPage) dto.MarketGroupLeaderboardResponse {
	if page == nil {
		return dto.MarketGroupLeaderboardResponse{Leaderboard: []dto.MarketGroupLeaderboardRowResponse{}}
	}
	rows := make([]dto.MarketGroupLeaderboardRowResponse, 0, len(page.Leaderboard))
	for _, entry := range page.Leaderboard {
		if entry == nil {
			continue
		}
		answers := make([]dto.MarketGroupLeaderboardAnswerResponse, 0, len(entry.Answers))
		for _, answer := range entry.Answers {
			if answer == nil {
				continue
			}
			answers = append(answers, dto.MarketGroupLeaderboardAnswerResponse{
				AnswerMarketID: answer.AnswerMarketID,
				AnswerLabel:    answer.AnswerLabel,
				DisplayOrder:   answer.DisplayOrder,
				Profit:         answer.Profit,
				CurrentValue:   answer.CurrentValue,
				TotalSpent:     answer.TotalSpent,
				Position:       answer.Position,
				YesSharesOwned: answer.YesSharesOwned,
				NoSharesOwned:  answer.NoSharesOwned,
			})
		}
		rows = append(rows, dto.MarketGroupLeaderboardRowResponse{
			Username:       entry.Username,
			Profit:         entry.Profit,
			CurrentValue:   entry.CurrentValue,
			TotalSpent:     entry.TotalSpent,
			Position:       entry.Position,
			YesSharesOwned: entry.YesSharesOwned,
			NoSharesOwned:  entry.NoSharesOwned,
			Rank:           entry.Rank,
			Answers:        answers,
		})
	}
	return dto.MarketGroupLeaderboardResponse{
		GroupID:     page.GroupID,
		Leaderboard: rows,
		Total:       page.Total,
		Freshness:   groupedActivityLiveFreshnessResponse(),
	}
}

func groupedActivityLiveFreshnessResponse() *dto.Freshness {
	freshness := readModelFreshnessToResponse(readmodels.NewFreshness(
		time.Now().UTC(),
		"live",
		groupedActivityLiveTargetFreshness,
		false,
	))
	return &freshness
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
		ID:                         group.ID,
		QuestionTitle:              group.QuestionTitle,
		Description:                group.Description,
		GroupType:                  group.GroupType,
		ProbabilityPolicy:          group.ProbabilityPolicy,
		ResolutionPolicy:           group.ResolutionPolicy,
		LifecycleStatus:            group.LifecycleStatus,
		Status:                     status,
		ProposalCost:               group.ProposalCost,
		CreatorUsername:            group.CreatorUsername,
		StewardUsername:            group.StewardUsername,
		ApprovedBy:                 group.ApprovedBy,
		ApprovedAt:                 group.ApprovedAt,
		RejectedBy:                 group.RejectedBy,
		RejectedAt:                 group.RejectedAt,
		RejectionReason:            group.RejectionReason,
		ResolutionDateTime:         group.ResolutionDateTime,
		AutoApproveAnswerAdditions: group.AutoApproveAnswerAdditions,
		CreatedAt:                  group.CreatedAt,
		UpdatedAt:                  group.UpdatedAt,
		AnswerCount:                len(group.Members),
	}
}

func marketGroupAnswerAdditionToResponse(addition *dmarkets.MarketGroupAnswerAddition) dto.MarketGroupAnswerAdditionResponse {
	if addition == nil {
		return dto.MarketGroupAnswerAdditionResponse{}
	}
	return dto.MarketGroupAnswerAdditionResponse{
		ID:              addition.ID,
		GroupID:         addition.GroupID,
		MarketID:        addition.MarketID,
		GroupTitle:      addition.GroupTitle,
		AnswerLabel:     addition.AnswerLabel,
		Status:          addition.Status,
		ProposedBy:      addition.ProposedBy,
		ReviewedBy:      addition.ReviewedBy,
		ReviewedAt:      addition.ReviewedAt,
		RejectionReason: addition.RejectionReason,
		AdditionCost:    addition.AdditionCost,
		CreatedAt:       addition.CreatedAt,
		UpdatedAt:       addition.UpdatedAt,
		MarketGroup:     marketGroupToResponse(addition.MarketGroup),
	}
}

func writeMarketGroupDetailsError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, dmarkets.ErrMarketGroupNotFound):
		_ = handlers.WriteFailure(w, http.StatusNotFound, handlers.ReasonMarketNotFound)
	case errors.Is(err, dmarkets.ErrUserNotFound):
		_ = handlers.WriteFailure(w, http.StatusNotFound, handlers.ReasonUserNotFound)
	case errors.Is(err, dmarkets.ErrUnauthorized):
		_ = handlers.WriteFailure(w, http.StatusForbidden, handlers.ReasonAuthorizationDenied)
	case errors.Is(err, dmarkets.ErrInsufficientBalance):
		_ = handlers.WriteFailure(w, http.StatusUnprocessableEntity, handlers.ReasonInsufficientBalance)
	case errors.Is(err, dmarkets.ErrInvalidState):
		_ = handlers.WriteFailure(w, http.StatusConflict, handlers.ReasonInvalidState)
	case errors.Is(err, dmarkets.ErrInvalidInput):
		_ = handlers.WriteFailure(w, http.StatusBadRequest, handlers.ReasonValidationFailed)
	default:
		writeDetailsError(w, err)
	}
}
