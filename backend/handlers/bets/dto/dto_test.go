package dto

import (
	"encoding/json"
	"reflect"
	"testing"
	"time"
)

func TestPlaceBetDTOJSONRoundTrip(t *testing.T) {
	req := PlaceBetRequest{
		MarketID: 12,
		Amount:   345,
		Outcome:  "YES",
	}

	payload, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	var decoded PlaceBetRequest
	if err := json.Unmarshal(payload, &decoded); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	if !reflect.DeepEqual(decoded, req) {
		t.Fatalf("roundtrip mismatch: got %+v want %+v", decoded, req)
	}
}

func TestSellBetDTOJSONRoundTrip(t *testing.T) {
	ts := time.Date(2025, 2, 3, 4, 5, 6, 0, time.UTC)

	resp := SellBetResponse{
		Username:      "tester",
		MarketID:      77,
		SharesSold:    5,
		SaleValue:     125,
		Dust:          1,
		Outcome:       "NO",
		TransactionAt: ts,
	}

	payload, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	var decoded SellBetResponse
	if err := json.Unmarshal(payload, &decoded); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	if !decoded.TransactionAt.Equal(ts) {
		t.Fatalf("expected timestamp %s, got %s", ts, decoded.TransactionAt)
	}

	decoded.TransactionAt = ts
	if !reflect.DeepEqual(decoded, resp) {
		t.Fatalf("roundtrip mismatch: got %+v want %+v", decoded, resp)
	}
}
