package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHomeHandlerReturnsSharedEnvelope(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/v0/home", nil)
	rec := httptest.NewRecorder()

	HomeHandler(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var response SuccessEnvelope[map[string]string]
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if !response.OK {
		t.Fatalf("expected ok=true, got false")
	}
	if response.Result["message"] != "Data From the Backend!" {
		t.Fatalf("expected backend message, got %q", response.Result["message"])
	}
}

func TestHomeHandlerRejectsUnsupportedMethods(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/v0/home", nil)
	rec := httptest.NewRecorder()

	HomeHandler(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected status 405, got %d: %s", rec.Code, rec.Body.String())
	}

	var response FailureEnvelope
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if response.OK {
		t.Fatalf("expected ok=false, got true")
	}
	if response.Reason != string(ReasonMethodNotAllowed) {
		t.Fatalf("expected reason %q, got %q", ReasonMethodNotAllowed, response.Reason)
	}
}
