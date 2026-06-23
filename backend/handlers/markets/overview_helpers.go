package marketshandlers

import (
	"context"
	"fmt"
	"strings"

	"socialpredict/handlers/markets/dto"
	dmarkets "socialpredict/internal/domain/markets"
)

type marketOverviewProvider interface {
	GetMarketDetails(ctx context.Context, marketID int64) (*dmarkets.MarketOverview, error)
}

type marketSummaryProvider interface {
	GetMarketSummaryReadModel(ctx context.Context, marketID int64) (*dmarkets.MarketSummaryReadModel, error)
}

type marketGroupLookupProvider interface {
	GetMarketGroupForMarket(ctx context.Context, marketID int64) (*dmarkets.MarketGroup, error)
}

func buildMarketOverviewResponses(ctx context.Context, provider marketOverviewProvider, markets []*dmarkets.Market) ([]*dto.MarketOverviewResponse, error) {
	if len(markets) == 0 {
		return []*dto.MarketOverviewResponse{}, nil
	}

	overviews := make([]*dto.MarketOverviewResponse, 0, len(markets))
	for _, market := range markets {
		overview, err := buildMarketOverviewResponse(ctx, provider, market)
		if err != nil {
			return nil, err
		}
		overviews = append(overviews, overview)
	}
	return overviews, nil
}

func buildMarketDiscoveryOverviewResponses(ctx context.Context, provider marketOverviewProvider, rows []dmarkets.MarketDiscoveryRow) ([]*dto.MarketOverviewResponse, error) {
	if len(rows) == 0 {
		return []*dto.MarketOverviewResponse{}, nil
	}

	overviews := make([]*dto.MarketOverviewResponse, 0, len(rows))
	for _, row := range rows {
		if row.Group == nil || row.Group.ID <= 0 {
			overview, err := buildMarketOverviewResponse(ctx, provider, row.Market)
			if err != nil {
				return nil, err
			}
			overviews = append(overviews, overview)
			continue
		}

		overview, err := buildMarketGroupDiscoveryOverviewResponse(ctx, provider, row)
		if err != nil {
			return nil, err
		}
		overviews = append(overviews, overview)
	}
	return overviews, nil
}

func buildMarketOverviewResponse(ctx context.Context, provider marketOverviewProvider, market *dmarkets.Market) (*dto.MarketOverviewResponse, error) {
	if market == nil {
		return &dto.MarketOverviewResponse{}, nil
	}
	if summaryProvider, ok := provider.(marketSummaryProvider); ok {
		summary, err := summaryProvider.GetMarketSummaryReadModel(ctx, market.ID)
		if err != nil {
			return nil, fmt.Errorf("market_id=%d: %w", market.ID, err)
		}
		return marketSummaryToOverviewResponse(ctx, provider, summary), nil
	}
	details, err := provider.GetMarketDetails(ctx, market.ID)
	if err != nil {
		return nil, fmt.Errorf("market_id=%d: %w", market.ID, err)
	}
	return marketOverviewToResponse(ctx, provider, details), nil
}

func buildMarketGroupDiscoveryOverviewResponse(ctx context.Context, provider marketOverviewProvider, row dmarkets.MarketDiscoveryRow) (*dto.MarketOverviewResponse, error) {
	children := row.Children
	if len(children) == 0 && row.Market != nil {
		children = []*dmarkets.Market{row.Market}
	}

	childOverviews := make([]*dto.MarketOverviewResponse, 0, len(children))
	for _, child := range children {
		overview, err := buildMarketOverviewResponse(ctx, provider, child)
		if err != nil {
			return nil, err
		}
		childOverviews = append(childOverviews, overview)
	}

	representative := row.Market
	if representative == nil && len(children) > 0 {
		representative = children[0]
	}
	response := &dto.MarketOverviewResponse{
		Market:  marketToResponse(representative),
		Creator: creatorResponseFromGroup(row.Group),
	}
	if len(childOverviews) > 0 && childOverviews[0].Creator != nil {
		response.Creator = childOverviews[0].Creator
	}
	if response.Market == nil {
		response.Market = &dto.MarketResponse{}
	}

	childIDs := make([]int64, 0, len(children))
	childResolutions := make([]dto.MarketGroupChildResolutionResponse, 0, len(children))
	allResolved := len(children) > 0
	tags := make([]dmarkets.MarketTag, 0)
	for _, childOverview := range childOverviews {
		if childOverview == nil {
			continue
		}
		response.TotalVolume += childOverview.TotalVolume
		response.MarketDust += childOverview.MarketDust
		if childOverview.NumUsers > response.NumUsers {
			response.NumUsers = childOverview.NumUsers
		}
		if childOverview.Market == nil {
			continue
		}
		childIDs = append(childIDs, childOverview.Market.ID)
		childResolutions = append(childResolutions, dto.MarketGroupChildResolutionResponse{
			MarketID:         childOverview.Market.ID,
			AnswerLabel:      answerLabelForMarket(row.Group, childOverview.Market.ID, childOverview.Market.QuestionTitle),
			IsResolved:       childOverview.Market.IsResolved,
			ResolutionResult: childOverview.Market.ResolutionResult,
		})
		allResolved = allResolved && childOverview.Market.IsResolved
		for _, tag := range childrenTagsByID(children, childOverview.Market.ID) {
			tags = append(tags, tag)
		}
	}

	response.Market.ID = representativeMarketID(representative, children)
	response.Market.QuestionTitle = row.Group.QuestionTitle
	response.Market.Description = row.Group.Description
	response.Market.CreatorUsername = row.Group.CreatorUsername
	response.Market.StewardUsername = row.Group.CurrentStewardUsername()
	response.Market.LifecycleStatus = row.Group.LifecycleStatus
	response.Market.Status = marketGroupLinkFromDomain(row.Group, response.Market.ID).Status
	response.Market.ApprovedBy = row.Group.ApprovedBy
	response.Market.ApprovedAt = row.Group.ApprovedAt
	response.Market.RejectedBy = row.Group.RejectedBy
	response.Market.RejectedAt = row.Group.RejectedAt
	response.Market.RejectionReason = row.Group.RejectionReason
	response.Market.ProposalCost = row.Group.ProposalCost
	response.Market.ResolutionDateTime = row.Group.ResolutionDateTime
	response.Market.CreatedAt = row.Group.CreatedAt
	response.Market.UpdatedAt = row.Group.UpdatedAt
	response.Market.Tags = uniqueMarketTagResponses(tags)
	response.Market.MarketGroup = marketGroupLinkFromDomain(row.Group, response.Market.ID)
	response.Market.IsMarketGroupAggregate = true
	response.Market.GroupChildMarketIDs = childIDs
	response.Market.GroupChildResolutions = childResolutions
	response.Market.IsResolved = allResolved
	return response, nil
}

func marketSummaryToOverviewResponse(ctx context.Context, provider any, summary *dmarkets.MarketSummaryReadModel) *dto.MarketOverviewResponse {
	if summary == nil {
		return &dto.MarketOverviewResponse{}
	}

	return &dto.MarketOverviewResponse{
		Market:          marketToResponseWithGroup(ctx, provider, summary.Market),
		Creator:         creatorResponseFromSummary(summary.Creator),
		LastProbability: summary.Accounting.LastProbability,
		NumUsers:        summary.Accounting.UserCount,
		TotalVolume:     summary.Accounting.VolumeWithDust,
		MarketDust:      summary.Accounting.MarketDust,
	}
}

func creatorResponseFromGroup(group *dmarkets.MarketGroup) *dto.CreatorResponse {
	if group == nil {
		return nil
	}
	return &dto.CreatorResponse{Username: group.CreatorUsername}
}

func representativeMarketID(representative *dmarkets.Market, children []*dmarkets.Market) int64 {
	if representative != nil && representative.ID > 0 {
		return representative.ID
	}
	for _, child := range children {
		if child != nil && child.ID > 0 {
			return child.ID
		}
	}
	return 0
}

func answerLabelForMarket(group *dmarkets.MarketGroup, marketID int64, fallback string) string {
	if group != nil {
		for _, member := range group.Members {
			if member.MarketID == marketID && strings.TrimSpace(member.AnswerLabel) != "" {
				return member.AnswerLabel
			}
		}
	}
	return fallback
}

func childrenTagsByID(children []*dmarkets.Market, marketID int64) []dmarkets.MarketTag {
	for _, child := range children {
		if child != nil && child.ID == marketID {
			return child.Tags
		}
	}
	return nil
}

func uniqueMarketTagResponses(tags []dmarkets.MarketTag) []dto.MarketTagResponse {
	if len(tags) == 0 {
		return []dto.MarketTagResponse{}
	}
	seen := map[string]bool{}
	unique := make([]dmarkets.MarketTag, 0, len(tags))
	for _, tag := range tags {
		key := tag.Slug
		if key == "" {
			key = fmt.Sprintf("%d", tag.ID)
		}
		if key == "" || seen[key] {
			continue
		}
		seen[key] = true
		unique = append(unique, tag)
	}
	return marketTagResponsesFromDomain(unique)
}

func marketOverviewToResponse(ctx context.Context, provider any, overview *dmarkets.MarketOverview) *dto.MarketOverviewResponse {
	if overview == nil {
		return &dto.MarketOverviewResponse{}
	}

	return &dto.MarketOverviewResponse{
		Market:          marketToResponseWithGroup(ctx, provider, overview.Market),
		Creator:         creatorResponseFromSummary(overview.Creator),
		LastProbability: overview.LastProbability,
		NumUsers:        overview.NumUsers,
		TotalVolume:     overview.TotalVolume,
		MarketDust:      overview.MarketDust,
	}
}

func marketToResponseWithGroup(ctx context.Context, provider any, market *dmarkets.Market) *dto.MarketResponse {
	response := marketToResponse(market)
	if response == nil || market == nil {
		return response
	}
	response.MarketGroup = marketGroupLinkForMarket(ctx, provider, market.ID)
	return response
}

func marketToResponse(market *dmarkets.Market) *dto.MarketResponse {
	if market == nil {
		return &dto.MarketResponse{}
	}

	return &dto.MarketResponse{
		ID:                 market.ID,
		QuestionTitle:      market.QuestionTitle,
		Description:        market.Description,
		OutcomeType:        market.OutcomeType,
		ResolutionDateTime: market.ResolutionDateTime,
		CreatorUsername:    market.CreatorUsername,
		StewardUsername:    market.CurrentStewardUsername(),
		YesLabel:           market.YesLabel,
		NoLabel:            market.NoLabel,
		Status:             market.Status,
		LifecycleStatus:    market.LifecycleStatus,
		ApprovedBy:         market.ApprovedBy,
		ApprovedAt:         market.ApprovedAt,
		RejectedBy:         market.RejectedBy,
		RejectedAt:         market.RejectedAt,
		RejectionReason:    market.RejectionReason,
		ProposalCost:       market.ProposalCost,
		IsResolved:         strings.EqualFold(market.Status, "resolved"),
		ResolutionResult:   market.ResolutionResult,
		CreatedAt:          market.CreatedAt,
		UpdatedAt:          market.UpdatedAt,
		Tags:               marketTagResponsesFromDomain(market.Tags),
	}
}

func creatorResponseFromSummary(summary *dmarkets.CreatorSummary) *dto.CreatorResponse {
	if summary == nil {
		return nil
	}
	return &dto.CreatorResponse{
		Username:      summary.Username,
		PersonalEmoji: summary.PersonalEmoji,
		DisplayName:   summary.DisplayName,
	}
}

func publicMarketResponseFromDomain(market *dmarkets.Market) dto.PublicMarketResponse {
	if market == nil {
		return dto.PublicMarketResponse{}
	}
	return dto.PublicMarketResponse{
		ID:                      market.ID,
		QuestionTitle:           market.QuestionTitle,
		Description:             market.Description,
		OutcomeType:             market.OutcomeType,
		ResolutionDateTime:      market.ResolutionDateTime,
		FinalResolutionDateTime: market.FinalResolutionDateTime,
		UTCOffset:               market.UTCOffset,
		IsResolved:              strings.EqualFold(market.Status, "resolved"),
		ResolutionResult:        market.ResolutionResult,
		InitialProbability:      market.InitialProbability,
		CreatorUsername:         market.CreatorUsername,
		StewardUsername:         market.CurrentStewardUsername(),
		CreatedAt:               market.CreatedAt,
		YesLabel:                market.YesLabel,
		NoLabel:                 market.NoLabel,
		Tags:                    marketTagResponsesFromDomain(market.Tags),
	}
}

func publicMarketResponseFromDomainWithGroup(ctx context.Context, provider any, market *dmarkets.Market) dto.PublicMarketResponse {
	response := publicMarketResponseFromDomain(market)
	if market != nil {
		response.MarketGroup = marketGroupLinkForMarket(ctx, provider, market.ID)
	}
	return response
}

func marketSummaryToDetailsResponse(ctx context.Context, provider any, summary *dmarkets.MarketSummaryReadModel) dto.MarketDetailsResponse {
	if summary == nil {
		return dto.MarketDetailsResponse{}
	}
	return dto.MarketDetailsResponse{
		Market:             publicMarketResponseFromDomainWithGroup(ctx, provider, summary.Market),
		Creator:            creatorResponseFromSummary(summary.Creator),
		ProbabilityChanges: probabilityChangesToResponse(summary.Accounting.ProbabilityChanges),
		NumUsers:           summary.Accounting.UserCount,
		TotalVolume:        summary.Accounting.VolumeWithDust,
		MarketDust:         summary.Accounting.MarketDust,
	}
}

func marketDetailsToResponse(ctx context.Context, provider any, details *dmarkets.MarketOverview) dto.MarketDetailsResponse {
	if details == nil {
		return dto.MarketDetailsResponse{}
	}
	return dto.MarketDetailsResponse{
		Market:                publicMarketResponseFromDomainWithGroup(ctx, provider, details.Market),
		Creator:               creatorResponseFromSummary(details.Creator),
		ProbabilityChanges:    probabilityChangesToResponse(details.ProbabilityChanges),
		NumUsers:              details.NumUsers,
		TotalVolume:           details.TotalVolume,
		MarketDust:            details.MarketDust,
		DescriptionAmendments: descriptionAmendmentsToResponse(details.DescriptionAmendments),
	}
}

func marketGroupLinkForMarket(ctx context.Context, provider any, marketID int64) *dto.MarketGroupLink {
	if marketID <= 0 {
		return nil
	}
	lookup, ok := provider.(marketGroupLookupProvider)
	if !ok {
		return nil
	}
	group, err := lookup.GetMarketGroupForMarket(ctx, marketID)
	if err != nil || group == nil || group.ID <= 0 {
		return nil
	}
	status := group.LifecycleStatus
	if strings.EqualFold(status, dmarkets.MarketLifecyclePublished) {
		status = dmarkets.MarketStatusActive
	}
	link := &dto.MarketGroupLink{
		ID:                 group.ID,
		QuestionTitle:      group.QuestionTitle,
		Description:        group.Description,
		GroupType:          group.GroupType,
		LifecycleStatus:    group.LifecycleStatus,
		Status:             status,
		AnswerCount:        len(group.Members),
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
	}
	for _, member := range group.Members {
		if member.MarketID == marketID {
			link.AnswerLabel = member.AnswerLabel
			link.DisplayOrder = member.DisplayOrder
			break
		}
	}
	return link
}

func marketGroupLinkFromDomain(group *dmarkets.MarketGroup, marketID int64) *dto.MarketGroupLink {
	if group == nil || group.ID <= 0 {
		return nil
	}
	status := group.LifecycleStatus
	if strings.EqualFold(status, dmarkets.MarketLifecyclePublished) {
		status = dmarkets.MarketStatusActive
	}
	link := &dto.MarketGroupLink{
		ID:                 group.ID,
		QuestionTitle:      group.QuestionTitle,
		Description:        group.Description,
		GroupType:          group.GroupType,
		LifecycleStatus:    group.LifecycleStatus,
		Status:             status,
		AnswerCount:        len(group.Members),
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
	}
	for _, member := range group.Members {
		if member.MarketID == marketID {
			link.AnswerLabel = member.AnswerLabel
			link.DisplayOrder = member.DisplayOrder
			break
		}
	}
	return link
}

func probabilityChangesToResponse(points []dmarkets.ProbabilityPoint) []dto.ProbabilityChangeResponse {
	if len(points) == 0 {
		return []dto.ProbabilityChangeResponse{}
	}
	result := make([]dto.ProbabilityChangeResponse, len(points))
	for i, point := range points {
		result[i] = dto.ProbabilityChangeResponse{
			Probability: point.Probability,
			Timestamp:   point.Timestamp,
		}
	}
	return result
}

func descriptionAmendmentsToResponse(amendments []dmarkets.MarketDescriptionAmendment) []dto.MarketDescriptionAmendmentResponse {
	if len(amendments) == 0 {
		return []dto.MarketDescriptionAmendmentResponse{}
	}
	result := make([]dto.MarketDescriptionAmendmentResponse, 0, len(amendments))
	for _, amendment := range amendments {
		result = append(result, dto.MarketDescriptionAmendmentResponse{
			ID:                         amendment.ID,
			MarketID:                   amendment.MarketID,
			MarketTitle:                amendment.MarketTitle,
			MarketDescription:          amendment.MarketDescription,
			MarketGroup:                marketGroupLinkFromDomain(amendment.MarketGroup, amendment.MarketID),
			Version:                    amendment.Version,
			Body:                       amendment.Body,
			BodyFormat:                 amendment.BodyFormat,
			Status:                     amendment.Status,
			CreatedBy:                  amendment.CreatedBy,
			CreatedAt:                  amendment.CreatedAt,
			UpdatedAt:                  amendment.UpdatedAt,
			ApprovedBy:                 amendment.ApprovedBy,
			ApprovedAt:                 amendment.ApprovedAt,
			RejectedBy:                 amendment.RejectedBy,
			RejectedAt:                 amendment.RejectedAt,
			RejectionReason:            amendment.RejectionReason,
			SubmitReason:               amendment.SubmitReason,
			PreviousApprovedAmendments: descriptionAmendmentsToResponse(amendment.PreviousApprovedAmendments),
		})
	}
	return result
}
