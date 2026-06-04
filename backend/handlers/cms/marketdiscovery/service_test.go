package marketdiscovery

import (
	"errors"
	"testing"

	"socialpredict/models"

	"gorm.io/gorm"
)

type mockRepository struct {
	page    *models.MarketDiscoveryPage
	saveErr error
	getErr  error
}

func (m *mockRepository) GetPageBySlug(string) (*models.MarketDiscoveryPage, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	if m.page == nil {
		return nil, gorm.ErrRecordNotFound
	}
	return m.page, nil
}

func (m *mockRepository) SavePage(page *models.MarketDiscoveryPage) error {
	if m.saveErr != nil {
		return m.saveErr
	}
	m.page = page
	return nil
}

func TestGetPageReturnsDefaultWhenMissing(t *testing.T) {
	svc := NewService(&mockRepository{})

	page, err := svc.GetPage(PageSlugMarkets)
	if err != nil {
		t.Fatalf("GetPage returned error: %v", err)
	}
	if page.Slug != PageSlugMarkets || page.PageType != PageTypeTop || page.DefaultRecommendationLimit != 20 {
		t.Fatalf("unexpected default page: %+v", page)
	}
}

func TestUpdatePagePersistsLayout(t *testing.T) {
	repo := &mockRepository{}
	svc := NewService(repo)

	page, err := svc.UpdatePage(PageSlugMarkets, UpdateInput{
		Title:                      "Forecast Markets",
		Description:                "Curated markets.",
		PageType:                   PageTypeTop,
		SearchScope:                SearchScopeAll,
		FeaturedTopicsEnabled:      true,
		FeaturedMarketsEnabled:     true,
		DefaultRecommendationLimit: 30,
		CuratedRecommendationLimit: 7,
		IsPublished:                true,
		UpdatedBy:                  "admin",
	})
	if err != nil {
		t.Fatalf("UpdatePage returned error: %v", err)
	}
	if page.Title != "Forecast Markets" || !page.FeaturedTopicsEnabled || page.CuratedRecommendationLimit != 7 {
		t.Fatalf("unexpected saved page: %+v", page)
	}
	if repo.page == nil || repo.page.UpdatedBy != "admin" {
		t.Fatalf("expected saved page with UpdatedBy admin, got %+v", repo.page)
	}
}

func TestUpdatePageRejectsInvalidSearchScope(t *testing.T) {
	svc := NewService(&mockRepository{})

	_, err := svc.UpdatePage(PageSlugMarkets, UpdateInput{
		Title:       "Markets",
		PageType:    PageTypeTop,
		SearchScope: "everywhere",
	})
	if err == nil {
		t.Fatalf("expected validation error")
	}
}

func TestUpdatePagePropagatesRepositoryError(t *testing.T) {
	wantErr := errors.New("database unavailable")
	svc := NewService(&mockRepository{saveErr: wantErr})

	_, err := svc.UpdatePage(PageSlugMarkets, UpdateInput{
		Title:       "Markets",
		PageType:    PageTypeTop,
		SearchScope: SearchScopeAll,
	})
	if !errors.Is(err, wantErr) {
		t.Fatalf("UpdatePage error = %v, want %v", err, wantErr)
	}
}
