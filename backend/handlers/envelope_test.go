package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestWriteResult(t *testing.T) {
	recorder := httptest.NewRecorder()

	err := WriteResult(recorder, http.StatusCreated, map[string]int{"value": 7})
	if err != nil {
		t.Fatalf("WriteResult returned error: %v", err)
	}

	if recorder.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d", recorder.Code)
	}

	if contentType := recorder.Header().Get("Content-Type"); contentType != "application/json" {
		t.Fatalf("expected application/json content type, got %q", contentType)
	}

	var response SuccessEnvelope[map[string]int]
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if !response.OK {
		t.Fatalf("expected ok=true, got false")
	}

	if response.Result["value"] != 7 {
		t.Fatalf("expected result value 7, got %d", response.Result["value"])
	}
}

func TestWriteBusinessFailure(t *testing.T) {
	recorder := httptest.NewRecorder()

	err := WriteBusinessFailure(recorder, ReasonValidationFailed)
	if err != nil {
		t.Fatalf("WriteBusinessFailure returned error: %v", err)
	}

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", recorder.Code)
	}

	var response FailureEnvelope
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if response.OK {
		t.Fatalf("expected ok=false, got true")
	}

	if response.Reason != string(ReasonValidationFailed) {
		t.Fatalf("expected reason %q, got %q", ReasonValidationFailed, response.Reason)
	}
}

func TestIsValidationMessage(t *testing.T) {
	if !IsValidationMessage("display name must be between 1 and 50 characters") {
		t.Fatalf("expected validation phrase to match")
	}
	if IsValidationMessage("database connection string leaked") {
		t.Fatalf("expected internal error phrase not to match")
	}
}
