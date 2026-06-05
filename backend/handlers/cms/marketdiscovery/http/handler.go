package http

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"

	"socialpredict/handlers"
	"socialpredict/handlers/authhttp"
	"socialpredict/handlers/cms/marketdiscovery"
	dusers "socialpredict/internal/domain/users"
	authsvc "socialpredict/internal/service/auth"
	"socialpredict/models"
)

type Handler struct {
	svc  *marketdiscovery.Service
	auth authsvc.Authenticator
}

func NewHandler(svc *marketdiscovery.Service, auth authsvc.Authenticator) *Handler {
	return &Handler{svc: svc, auth: auth}
}

type updateReq struct {
	Title                      string `json:"title"`
	Description                string `json:"description"`
	PageType                   string `json:"pageType"`
	PrimaryTagSlug             string `json:"primaryTagSlug"`
	SearchScope                string `json:"searchScope"`
	FeaturedTopicsEnabled      bool   `json:"featuredTopicsEnabled"`
	FeaturedMarketsEnabled     bool   `json:"featuredMarketsEnabled"`
	SectionsEnabled            bool   `json:"sectionsEnabled"`
	DefaultRecommendationLimit int    `json:"defaultRecommendationLimit"`
	CuratedRecommendationLimit int    `json:"curatedRecommendationLimit"`
	Version                    uint   `json:"version"`
}

type replaceSectionsReq struct {
	Sections []sectionReq `json:"sections"`
}

type sectionReq struct {
	Slug          string `json:"slug"`
	Title         string `json:"title"`
	Description   string `json:"description"`
	TagFilterSlug string `json:"tagFilterSlug"`
	SortOrder     int    `json:"sortOrder"`
	IsActive      bool   `json:"isActive"`
}

type replacePinsReq struct {
	Pins []pinReq `json:"pins"`
}

type pinReq struct {
	PinType        string `json:"pinType"`
	MarketID       int64  `json:"marketId"`
	TargetPageSlug string `json:"targetPageSlug"`
	Label          string `json:"label"`
	SortOrder      int    `json:"sortOrder"`
}

type pageResponse struct {
	Slug                       string            `json:"slug"`
	Title                      string            `json:"title"`
	Description                string            `json:"description"`
	PageType                   string            `json:"pageType"`
	PrimaryTagSlug             string            `json:"primaryTagSlug"`
	SearchScope                string            `json:"searchScope"`
	FeaturedTopicsEnabled      bool              `json:"featuredTopicsEnabled"`
	FeaturedMarketsEnabled     bool              `json:"featuredMarketsEnabled"`
	SectionsEnabled            bool              `json:"sectionsEnabled"`
	DefaultRecommendationLimit int               `json:"defaultRecommendationLimit"`
	CuratedRecommendationLimit int               `json:"curatedRecommendationLimit"`
	RecommendationLimit        int               `json:"recommendationLimit"`
	Version                    uint              `json:"version"`
	UpdatedAt                  string            `json:"updatedAt,omitempty"`
	Sections                   []sectionResponse `json:"sections"`
	Pins                       []pinResponse     `json:"pins"`
}

type sectionResponse struct {
	ID            uint   `json:"id,omitempty"`
	Slug          string `json:"slug"`
	Title         string `json:"title"`
	Description   string `json:"description,omitempty"`
	TagFilterSlug string `json:"tagFilterSlug,omitempty"`
	SortOrder     int    `json:"sortOrder"`
	IsActive      bool   `json:"isActive"`
}

type pinResponse struct {
	ID             uint   `json:"id,omitempty"`
	PinType        string `json:"pinType"`
	MarketID       int64  `json:"marketId,omitempty"`
	TargetPageSlug string `json:"targetPageSlug,omitempty"`
	Label          string `json:"label,omitempty"`
	SortOrder      int    `json:"sortOrder"`
}

func (h *Handler) PublicGet(w http.ResponseWriter, r *http.Request) {
	composition, err := h.svc.GetComposition(mux.Vars(r)["slug"])
	if err != nil {
		_ = handlers.WriteFailure(w, http.StatusBadRequest, handlers.ReasonInvalidRequest)
		return
	}
	_ = handlers.WriteResult(w, http.StatusOK, responseFromComposition(composition))
}

func (h *Handler) AdminUpdate(w http.ResponseWriter, r *http.Request) {
	if h.auth == nil {
		_ = handlers.WriteFailure(w, http.StatusInternalServerError, handlers.ReasonInternalError)
		return
	}
	admin, authErr := h.auth.RequireAdmin(r)
	if authErr != nil {
		_ = authhttp.WriteFailure(w, authErr)
		return
	}

	var in updateReq
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		_ = handlers.WriteFailure(w, http.StatusBadRequest, handlers.ReasonInvalidRequest)
		return
	}
	page, err := h.svc.UpdatePage(mux.Vars(r)["slug"], marketdiscovery.UpdateInput{
		Title:                      in.Title,
		Description:                in.Description,
		PageType:                   in.PageType,
		PrimaryTagSlug:             in.PrimaryTagSlug,
		SearchScope:                in.SearchScope,
		FeaturedTopicsEnabled:      in.FeaturedTopicsEnabled,
		FeaturedMarketsEnabled:     in.FeaturedMarketsEnabled,
		SectionsEnabled:            in.SectionsEnabled,
		DefaultRecommendationLimit: in.DefaultRecommendationLimit,
		CuratedRecommendationLimit: in.CuratedRecommendationLimit,
		Version:                    in.Version,
		UpdatedBy:                  admin.Username,
	})
	if err != nil {
		_ = handlers.WriteFailure(w, http.StatusBadRequest, handlers.ReasonValidationFailed)
		return
	}
	composition, err := h.svc.GetComposition(page.Slug)
	if err != nil {
		_ = handlers.WriteFailure(w, http.StatusInternalServerError, handlers.ReasonInternalError)
		return
	}
	_ = handlers.WriteResult(w, http.StatusOK, responseFromComposition(composition))
}

func (h *Handler) AdminReplaceSections(w http.ResponseWriter, r *http.Request) {
	admin, ok := h.requireAdmin(w, r)
	if !ok {
		return
	}
	_ = admin
	var in replaceSectionsReq
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		_ = handlers.WriteFailure(w, http.StatusBadRequest, handlers.ReasonInvalidRequest)
		return
	}
	composition, err := h.svc.ReplaceSections(mux.Vars(r)["slug"], sectionInputsFromRequest(in.Sections))
	if err != nil {
		_ = handlers.WriteFailure(w, http.StatusBadRequest, handlers.ReasonValidationFailed)
		return
	}
	_ = handlers.WriteResult(w, http.StatusOK, responseFromComposition(composition))
}

func (h *Handler) AdminReplacePins(w http.ResponseWriter, r *http.Request) {
	admin, ok := h.requireAdmin(w, r)
	if !ok {
		return
	}
	var in replacePinsReq
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		_ = handlers.WriteFailure(w, http.StatusBadRequest, handlers.ReasonInvalidRequest)
		return
	}
	composition, err := h.svc.ReplacePins(mux.Vars(r)["slug"], pinInputsFromRequest(in.Pins), admin.Username)
	if err != nil {
		_ = handlers.WriteFailure(w, http.StatusBadRequest, handlers.ReasonValidationFailed)
		return
	}
	_ = handlers.WriteResult(w, http.StatusOK, responseFromComposition(composition))
}

func (h *Handler) requireAdmin(w http.ResponseWriter, r *http.Request) (*dusers.User, bool) {
	if h.auth == nil {
		_ = handlers.WriteFailure(w, http.StatusInternalServerError, handlers.ReasonInternalError)
		return nil, false
	}
	admin, authErr := h.auth.RequireAdmin(r)
	if authErr != nil {
		_ = authhttp.WriteFailure(w, authErr)
		return nil, false
	}
	return admin, true
}

func responseFromComposition(composition *marketdiscovery.PageComposition) pageResponse {
	return responseFromPage(composition.Page, composition.Sections, composition.Pins)
}

func responseFromPage(page *models.MarketDiscoveryPage, sections []models.MarketDiscoverySection, pins []models.MarketDiscoveryPin) pageResponse {
	response := pageResponse{
		Slug:                       page.Slug,
		Title:                      page.Title,
		Description:                page.Description,
		PageType:                   page.PageType,
		PrimaryTagSlug:             page.PrimaryTagSlug,
		SearchScope:                page.SearchScope,
		FeaturedTopicsEnabled:      page.FeaturedTopicsEnabled,
		FeaturedMarketsEnabled:     page.FeaturedMarketsEnabled,
		SectionsEnabled:            page.SectionsEnabled,
		DefaultRecommendationLimit: page.DefaultRecommendationLimit,
		CuratedRecommendationLimit: page.CuratedRecommendationLimit,
		Version:                    page.Version,
		Sections:                   sectionResponses(sections),
		Pins:                       pinResponses(pins),
	}
	if page.UpdatedAt.IsZero() {
		response.UpdatedAt = ""
	} else {
		response.UpdatedAt = page.UpdatedAt.Format("2006-01-02T15:04:05Z07:00")
	}
	response.RecommendationLimit = response.DefaultRecommendationLimit
	if page.FeaturedTopicsEnabled || page.FeaturedMarketsEnabled || page.SectionsEnabled {
		response.RecommendationLimit = response.CuratedRecommendationLimit
	}
	return response
}

func sectionInputsFromRequest(sections []sectionReq) []marketdiscovery.SectionInput {
	inputs := make([]marketdiscovery.SectionInput, 0, len(sections))
	for _, section := range sections {
		inputs = append(inputs, marketdiscovery.SectionInput{
			Slug:          section.Slug,
			Title:         section.Title,
			Description:   section.Description,
			TagFilterSlug: section.TagFilterSlug,
			SortOrder:     section.SortOrder,
			IsActive:      section.IsActive,
		})
	}
	return inputs
}

func pinInputsFromRequest(pins []pinReq) []marketdiscovery.PinInput {
	inputs := make([]marketdiscovery.PinInput, 0, len(pins))
	for _, pin := range pins {
		inputs = append(inputs, marketdiscovery.PinInput{
			PinType:        pin.PinType,
			MarketID:       pin.MarketID,
			TargetPageSlug: pin.TargetPageSlug,
			Label:          pin.Label,
			SortOrder:      pin.SortOrder,
		})
	}
	return inputs
}

func sectionResponses(sections []models.MarketDiscoverySection) []sectionResponse {
	responses := make([]sectionResponse, 0, len(sections))
	for _, section := range sections {
		responses = append(responses, sectionResponse{
			ID:            section.ID,
			Slug:          section.Slug,
			Title:         section.Title,
			Description:   section.Description,
			TagFilterSlug: section.TagFilterSlug,
			SortOrder:     section.SortOrder,
			IsActive:      section.IsActive,
		})
	}
	return responses
}

func pinResponses(pins []models.MarketDiscoveryPin) []pinResponse {
	responses := make([]pinResponse, 0, len(pins))
	for _, pin := range pins {
		response := pinResponse{
			ID:             pin.ID,
			PinType:        pin.PinType,
			TargetPageSlug: pin.TargetPageSlug,
			Label:          pin.Label,
			SortOrder:      pin.SortOrder,
		}
		if pin.MarketID != nil {
			response.MarketID = *pin.MarketID
		}
		responses = append(responses, response)
	}
	return responses
}
