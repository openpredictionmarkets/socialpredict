package marketshandlers

import (
	"context"
	"html/template"
	"net/http"
	"strconv"

	dmarkets "socialpredict/internal/domain/markets"

	"github.com/gorilla/mux"
)

type shareMetadataService interface {
	GetShareMetadata(ctx context.Context, marketID int64, config dmarkets.ShareMetadataConfig) (*dmarkets.ShareMetadata, error)
}

type ShareMetadataConfigProvider func(ctx context.Context, fallback dmarkets.ShareMetadataConfig) dmarkets.ShareMetadataConfig

type shareShellData struct {
	Title        string
	Description  string
	CanonicalURL string
	ImageURL     string
	ImageAlt     string
	SiteName     string
	PublicStatus string
	MarketID     int64
}

var marketShareShellTemplate = template.Must(template.New("market-share-shell").Parse(`<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8" />
  <meta name="viewport" content="width=device-width, initial-scale=1" />
  <meta name="theme-color" content="#000000" />
  <title>{{ .Title }}</title>
  <meta name="description" content="{{ .Description }}" />
  <link rel="canonical" href="{{ .CanonicalURL }}" />
  <meta property="og:title" content="{{ .Title }}" />
  <meta property="og:type" content="website" />
  <meta property="og:url" content="{{ .CanonicalURL }}" />
  <meta property="og:description" content="{{ .Description }}" />
  <meta property="og:image" content="{{ .ImageURL }}" />
  <meta property="og:image:alt" content="{{ .ImageAlt }}" />
  <meta property="og:site_name" content="{{ .SiteName }}" />
  <meta name="twitter:card" content="summary_large_image" />
  <meta name="twitter:title" content="{{ .Title }}" />
  <meta name="twitter:description" content="{{ .Description }}" />
  <meta name="twitter:image" content="{{ .ImageURL }}" />
  <meta name="twitter:image:alt" content="{{ .ImageAlt }}" />
  <meta name="socialpredict:market_id" content="{{ .MarketID }}" />
  <meta name="socialpredict:public_status" content="{{ .PublicStatus }}" />
  <script src="/env-config.js"></script>
  <script type="module" crossorigin src="/assets/index.js"></script>
  <link rel="stylesheet" crossorigin href="/assets/index.css" />
</head>
<body>
  <noscript>You need to enable JavaScript to run this app.</noscript>
  <div id="root"></div>
  <div id="modal-root"></div>
</body>
</html>
`))

// MarketShareShellHandler serves initial HTML metadata for public market URLs.
func MarketShareShellHandler(svc shareMetadataService, config dmarkets.ShareMetadataConfig, providers ...ShareMetadataConfigProvider) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		id, err := strconv.ParseInt(vars["id"], 10, 64)
		if err != nil || id <= 0 {
			writeInvalidRequest(w)
			return
		}

		resolvedConfig := config
		for _, provider := range providers {
			if provider != nil {
				resolvedConfig = provider(r.Context(), resolvedConfig)
			}
		}

		metadata, err := svc.GetShareMetadata(r.Context(), id, resolvedConfig)
		if err != nil {
			writeDetailsError(w, err)
			return
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Header().Set("Cache-Control", "public, max-age=300")
		w.WriteHeader(http.StatusOK)
		_ = marketShareShellTemplate.Execute(w, shareShellData{
			Title:        metadata.Title,
			Description:  metadata.Description,
			CanonicalURL: metadata.CanonicalURL,
			ImageURL:     metadata.ImageURL,
			ImageAlt:     metadata.ImageAlt,
			SiteName:     metadata.SiteName,
			PublicStatus: metadata.PublicStatus,
			MarketID:     metadata.MarketID,
		})
	}
}
