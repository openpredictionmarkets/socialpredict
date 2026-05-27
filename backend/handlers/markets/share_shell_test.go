package marketshandlers

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"socialpredict/handlers"
	dmarkets "socialpredict/internal/domain/markets"

	"github.com/gorilla/mux"
)

type shareMetadataServiceStub struct {
	metadata *dmarkets.ShareMetadata
	err      error
}

func (s shareMetadataServiceStub) GetShareMetadata(context.Context, int64, dmarkets.ShareMetadataConfig) (*dmarkets.ShareMetadata, error) {
	return s.metadata, s.err
}

func TestMarketShareShellHandlerEmitsOpenGraphHTML(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/markets/42", nil)
	req = mux.SetURLVars(req, map[string]string{"id": "42"})
	rec := httptest.NewRecorder()

	MarketShareShellHandler(shareMetadataServiceStub{
		metadata: &dmarkets.ShareMetadata{
			MarketID:     42,
			Title:        `Will "quoted" markets share? | SocialPredict`,
			Description:  "A safe public description.",
			CanonicalURL: "https://kconfs.com/markets/42",
			ImageURL:     "https://kconfs.com/logo512.png",
			PublicStatus: dmarkets.MarketStatusActive,
			SiteName:     "SocialPredict",
		},
	}, dmarkets.ShareMetadataConfig{}).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
	if contentType := rec.Header().Get("Content-Type"); !strings.HasPrefix(contentType, "text/html") {
		t.Fatalf("Content-Type = %q", contentType)
	}
	body := rec.Body.String()
	for _, want := range []string{
		`<meta property="og:title" content="Will &#34;quoted&#34; markets share? | SocialPredict" />`,
		`<meta property="og:url" content="https://kconfs.com/markets/42" />`,
		`<meta property="og:image" content="https://kconfs.com/logo512.png" />`,
		`<script type="module" crossorigin src="/assets/index.js"></script>`,
		`<link rel="stylesheet" crossorigin href="/assets/index.css" />`,
	} {
		if !strings.Contains(body, want) {
			t.Fatalf("share shell missing %q in body:\n%s", want, body)
		}
	}
}

func TestMarketShareShellHandlerRejectsNonPublicMarket(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/markets/7", nil)
	req = mux.SetURLVars(req, map[string]string{"id": "7"})
	rec := httptest.NewRecorder()

	MarketShareShellHandler(shareMetadataServiceStub{err: dmarkets.ErrMarketNotFound}, dmarkets.ShareMetadataConfig{}).ServeHTTP(rec, req)

	assertFailureEnvelope(t, rec, http.StatusNotFound, handlers.ReasonMarketNotFound)
}

func TestMarketShareShellHandlerRejectsBadID(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/markets/nope", nil)
	req = mux.SetURLVars(req, map[string]string{"id": "nope"})
	rec := httptest.NewRecorder()

	MarketShareShellHandler(shareMetadataServiceStub{err: errors.New("should not be called")}, dmarkets.ShareMetadataConfig{}).ServeHTTP(rec, req)

	assertFailureEnvelope(t, rec, http.StatusBadRequest, handlers.ReasonInvalidRequest)
}
