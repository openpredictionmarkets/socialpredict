package dto

import (
	"encoding/json"
	"reflect"
	"testing"
	"time"
)

func TestCreateMarketRequestJSONParsing(t *testing.T) {
	input := []byte(`{
		"questionTitle":"Will it rain?",
		"description":"Forecast for tomorrow",
		"outcomeType":"BINARY",
		"resolutionDateTime":"2025-01-01T00:00:00Z",
		"yesLabel":"Yes",
		"noLabel":"No"
	}`)

	var req CreateMarketRequest
	if err := json.Unmarshal(input, &req); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	wantTime := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	if req.QuestionTitle != "Will it rain?" ||
		req.Description != "Forecast for tomorrow" ||
		req.OutcomeType != "BINARY" ||
		!req.ResolutionDateTime.Equal(wantTime) ||
		req.YesLabel != "Yes" ||
		req.NoLabel != "No" {
		t.Fatalf("unexpected request contents: %+v", req)
	}
}

func TestMarketOverviewResponseJSONRoundTrip(t *testing.T) {
	now := time.Date(2025, 3, 4, 5, 6, 7, 0, time.UTC)
	market := &MarketResponse{
		ID:                 11,
		QuestionTitle:      "Sample market",
		Description:        "Desc",
		OutcomeType:        "BINARY",
		ResolutionDateTime: now,
		CreatorUsername:    "author",
		YesLabel:           "Up",
		NoLabel:            "Down",
		Status:             "active",
		IsResolved:         false,
		ResolutionResult:   "",
		CreatedAt:          now.Add(-time.Hour),
		UpdatedAt:          now,
	}
	creator := &CreatorResponse{
		Username:      "author",
		PersonalEmoji: "😀",
		DisplayName:   "Author",
	}

	resp := MarketOverviewResponse{
		Market:          market,
		Creator:         creator,
		LastProbability: 0.42,
		NumUsers:        10,
		TotalVolume:     1234,
		MarketDust:      2,
	}

	payload, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	var decoded MarketOverviewResponse
	if err := json.Unmarshal(payload, &decoded); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	if decoded.Market == nil || decoded.Creator == nil {
		t.Fatalf("expected nested objects, got nil: %+v", decoded)
	}

	decoded.Market.CreatedAt = market.CreatedAt
	if !reflect.DeepEqual(decoded, resp) {
		t.Fatalf("roundtrip mismatch: got %+v want %+v", decoded, resp)
	}
}

func TestSearchResponseJSON(t *testing.T) {
	resp := SearchResponse{
		PrimaryResults: []*MarketOverviewResponse{
			{
				Market: &MarketResponse{
					ID:            1,
					QuestionTitle: "A",
					Status:        "ACTIVE",
				},
				Creator:         &CreatorResponse{Username: "alice"},
				LastProbability: 0.4,
			},
		},
		FallbackResults: []*MarketOverviewResponse{
			{
				Market: &MarketResponse{
					ID:            2,
					QuestionTitle: "B",
					Status:        "CLOSED",
				},
				Creator:         &CreatorResponse{Username: "bob"},
				LastProbability: 0.6,
			},
		},
		Query:         "a",
		PrimaryStatus: "ACTIVE",
		PrimaryCount:  1,
		FallbackCount: 1,
		TotalCount:    2,
		FallbackUsed:  true,
	}

	payload, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	var decoded SearchResponse
	if err := json.Unmarshal(payload, &decoded); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	if decoded.TotalCount != 2 || !decoded.FallbackUsed {
		t.Fatalf("unexpected decoded content: %+v", decoded)
	}

	if len(decoded.PrimaryResults) != 1 || len(decoded.FallbackResults) != 1 {
		t.Fatalf("results mismatch: %+v", decoded)
	}
}
