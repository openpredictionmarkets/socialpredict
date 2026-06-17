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
		if summaryProvider, ok := provider.(marketSummaryProvider); ok {
			summary, err := summaryProvider.GetMarketSummaryReadModel(ctx, market.ID)
			if err != nil {
				return nil, fmt.Errorf("market_id=%d: %w", market.ID, err)
			}
			overviews = append(overviews, marketSummaryToOverviewResponse(ctx, provider, summary))
			continue
		}
		details, err := provider.GetMarketDetails(ctx, market.ID)
		if err != nil {
			return nil, fmt.Errorf("market_id=%d: %w", market.ID, err)
		}
		overviews = append(overviews, marketOverviewToResponse(ctx, provider, details))
	}
	return overviews, nil
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
