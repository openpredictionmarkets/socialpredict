package http

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"

	"socialpredict/handlers"
	"socialpredict/handlers/authhttp"
	"socialpredict/handlers/cms/marketdiscovery"
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
	IsPublished                bool   `json:"isPublished"`
	Version                    uint   `json:"version"`
}

type pageResponse struct {
	Slug                       string `json:"slug"`
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
	RecommendationLimit        int    `json:"recommendationLimit"`
	IsPublished                bool   `json:"isPublished"`
	Version                    uint   `json:"version"`
	UpdatedAt                  string `json:"updatedAt,omitempty"`
}

func (h *Handler) PublicGet(w http.ResponseWriter, r *http.Request) {
	page, err := h.svc.GetPage(mux.Vars(r)["slug"])
	if err != nil {
		_ = handlers.WriteFailure(w, http.StatusBadRequest, handlers.ReasonInvalidRequest)
		return
	}
	_ = handlers.WriteResult(w, http.StatusOK, responseFromPage(page))
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
		IsPublished:                in.IsPublished,
		Version:                    in.Version,
		UpdatedBy:                  admin.Username,
	})
	if err != nil {
		_ = handlers.WriteFailure(w, http.StatusBadRequest, handlers.ReasonValidationFailed)
		return
	}
	_ = handlers.WriteResult(w, http.StatusOK, responseFromPage(page))
}

func responseFromPage(page *models.MarketDiscoveryPage) pageResponse {
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
		IsPublished:                page.IsPublished,
		Version:                    page.Version,
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
