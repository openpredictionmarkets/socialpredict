package errors

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestWriteJSONError_SetsContentTypeAndBody(t *testing.T) {
	rec := httptest.NewRecorder()
	WriteJSONError(rec, http.StatusBadRequest, "bad input")

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
	if !strings.Contains(rec.Header().Get("Content-Type"), "application/json") {
		t.Errorf("expected application/json Content-Type, got %q", rec.Header().Get("Content-Type"))
	}
	var resp HTTPErrorResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if resp.Error != "bad input" {
		t.Errorf("expected error 'bad input', got %q", resp.Error)
	}
}

func TestWriteInternalError_Returns500AndLogsError(t *testing.T) {
	rec := httptest.NewRecorder()
	WriteInternalError(rec, errors.New("db connection dropped"), "TestWriteInternalError")

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", rec.Code)
	}
	var resp HTTPErrorResponse
	json.NewDecoder(rec.Body).Decode(&resp)
	// Internal detail must NOT be in client response
	if strings.Contains(resp.Error, "db connection") {
		t.Error("internal error detail must not leak to client")
	}
	if resp.Error == "" {
		t.Error("expected non-empty error message for client")
	}
}

func TestWriteInternalError_NilError_Still500(t *testing.T) {
	rec := httptest.NewRecorder()
	WriteInternalError(rec, nil, "some context")
	if rec.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", rec.Code)
	}
}

func TestWriteNotFound_IncludesResourceName(t *testing.T) {
	rec := httptest.NewRecorder()
	WriteNotFound(rec, "market")

	if rec.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "market") {
		t.Errorf("expected resource name in body, got %q", rec.Body.String())
	}
}

func TestWriteNotFound_EmptyResource_DefaultMessage(t *testing.T) {
	rec := httptest.NewRecorder()
	WriteNotFound(rec, "")
	if rec.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rec.Code)
	}
}

func TestWriteBadRequest(t *testing.T) {
	rec := httptest.NewRecorder()
	WriteBadRequest(rec, "question title too long")
	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "question title too long") {
		t.Errorf("expected message in body")
	}
}

func TestWriteUnauthorized(t *testing.T) {
	rec := httptest.NewRecorder()
	WriteUnauthorized(rec)
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rec.Code)
	}
}

func TestWriteForbidden(t *testing.T) {
	rec := httptest.NewRecorder()
	WriteForbidden(rec)
	if rec.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", rec.Code)
	}
}

func TestHandleHTTPError_SetsContentType(t *testing.T) {
	rec := httptest.NewRecorder()
	HandleHTTPError(rec, errors.New("internal"), http.StatusInternalServerError, "server error")
	if !strings.Contains(rec.Header().Get("Content-Type"), "application/json") {
		t.Errorf("expected application/json Content-Type after fix, got %q", rec.Header().Get("Content-Type"))
	}
}
