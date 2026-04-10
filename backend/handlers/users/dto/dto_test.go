package dto

import (
	"encoding/json"
	"reflect"
	"testing"
	"time"
)

func TestChangeProfileRequestsJSON(t *testing.T) {
	req := ChangePersonalLinksRequest{
		PersonalLink1: "https://one",
		PersonalLink2: "https://two",
		PersonalLink3: "https://three",
		PersonalLink4: "https://four",
	}

	payload, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	var decoded ChangePersonalLinksRequest
	if err := json.Unmarshal(payload, &decoded); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	if !reflect.DeepEqual(decoded, req) {
		t.Fatalf("roundtrip mismatch: got %+v want %+v", decoded, req)
	}
}

func TestPrivateUserResponseJSONRoundTrip(t *testing.T) {
	resp := PrivateUserResponse{
		ID:                    9,
		Username:              "tester",
		DisplayName:           "Tester",
		UserType:              "REGULAR",
		InitialAccountBalance: 1000,
		AccountBalance:        900,
		PersonalEmoji:         "😀",
		Description:           "New user",
		PersonalLink1:         "link1",
		PersonalLink2:         "link2",
		PersonalLink3:         "link3",
		PersonalLink4:         "link4",
		Email:                 "user@example.com",
		APIKey:                "api-key",
		MustChangePassword:    true,
	}

	payload, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	var decoded PrivateUserResponse
	if err := json.Unmarshal(payload, &decoded); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	if !reflect.DeepEqual(decoded, resp) {
		t.Fatalf("roundtrip mismatch: got %+v want %+v", decoded, resp)
	}
}

func TestPortfolioResponseJSONRoundTrip(t *testing.T) {
	ts := time.Date(2025, 5, 6, 7, 8, 9, 0, time.UTC)
	resp := PortfolioResponse{
		PortfolioItems: []PortfolioItemResponse{
			{
				MarketID:       1,
				QuestionTitle:  "Market",
				YesSharesOwned: 10,
				NoSharesOwned:  5,
				LastBetPlaced:  ts,
			},
		},
		TotalSharesOwned: 15,
	}

	payload, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	var decoded PortfolioResponse
	if err := json.Unmarshal(payload, &decoded); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	if len(decoded.PortfolioItems) != 1 || decoded.TotalSharesOwned != 15 {
		t.Fatalf("unexpected portfolio payload: %+v", decoded)
	}

	if !decoded.PortfolioItems[0].LastBetPlaced.Equal(ts) {
		t.Fatalf("expected timestamp %s, got %s", ts, decoded.PortfolioItems[0].LastBetPlaced)
	}
}
