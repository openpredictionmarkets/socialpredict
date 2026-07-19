package mcpserver

import (
	"testing"

	dmarkets "socialpredict/internal/domain/markets"
)

func TestNormalizeStatus(t *testing.T) {
	tests := []struct {
		name      string
		raw       string
		canonical string
		filter    string
		wantErr   bool
	}{
		{name: "blank is all", raw: "", canonical: "all", filter: ""},
		{name: "all clears backend filter", raw: " all ", canonical: "all", filter: ""},
		{name: "open aliases active", raw: " OPEN ", canonical: "active", filter: "active"},
		{name: "active stays active", raw: "active", canonical: "active", filter: "active"},
		{name: "closed stays closed", raw: "closed", canonical: "closed", filter: "closed"},
		{name: "resolved stays resolved", raw: "resolved", canonical: "resolved", filter: "resolved"},
		{name: "bad status fails", raw: "paused", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NormalizeStatus(tt.raw)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("NormalizeStatus returned nil error")
				}
				return
			}
			if err != nil {
				t.Fatalf("NormalizeStatus returned error: %v", err)
			}
			if got.Canonical != tt.canonical || got.Filter != tt.filter {
				t.Fatalf("NormalizeStatus = %#v, want canonical=%q filter=%q", got, tt.canonical, tt.filter)
			}
		})
	}
}

func TestNormalizeTagSlug(t *testing.T) {
	got, err := NormalizeTagSlug("  --Macro-News-- ")
	if err != nil {
		t.Fatalf("NormalizeTagSlug returned error: %v", err)
	}
	if got != "macro-news" {
		t.Fatalf("NormalizeTagSlug = %q, want macro-news", got)
	}

	if _, err := NormalizeTagSlug("bad slug with spaces"); err == nil {
		t.Fatalf("NormalizeTagSlug accepted invalid slug")
	}
}

func TestNormalizePage(t *testing.T) {
	got := NormalizePage(0, -10)
	if got != (dmarkets.Page{Limit: 20, Offset: 0}) {
		t.Fatalf("NormalizePage default = %#v", got)
	}
	got = NormalizePage(500, 9)
	if got != (dmarkets.Page{Limit: 100, Offset: 9}) {
		t.Fatalf("NormalizePage capped = %#v", got)
	}
}

func TestNormalizeOutcome(t *testing.T) {
	got, err := NormalizeOutcome(" yes ")
	if err != nil || got != "YES" {
		t.Fatalf("NormalizeOutcome yes = %q, %v", got, err)
	}
	got, err = NormalizeOutcome("NO")
	if err != nil || got != "NO" {
		t.Fatalf("NormalizeOutcome no = %q, %v", got, err)
	}
	if _, err := NormalizeOutcome("MAYBE"); err == nil {
		t.Fatalf("NormalizeOutcome accepted invalid outcome")
	}
}
