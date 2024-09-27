package helpers_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"socialpredict/handlers/helpers"
	"testing"

	"gorm.io/gorm"
)

func TestHandleError_UserNotFound(t *testing.T) {
	// Create a ResponseRecorder to record the response
	rr := httptest.NewRecorder()

	// Create an error that is gorm.ErrRecordNotFound
	err := gorm.ErrRecordNotFound
	message := "Error fetching user"

	// Call handleError with the ResponseRecorder, error, and message
	helpers.HandleError(rr, err, message)

	// Check the status code is 404 Not Found
	if status := rr.Code; status != http.StatusNotFound {
		t.Errorf("Expected status code %d, got %d", http.StatusNotFound, status)
	}

	// Check the response body is "User not found\n"
	expectedBody := "User not found\n"
	if rr.Body.String() != expectedBody {
		t.Errorf("Expected response body '%s', got '%s'", expectedBody, rr.Body.String())
	}
}

func TestHandleError_InternalServerError(t *testing.T) {
	// Create a ResponseRecorder to record the response
	rr := httptest.NewRecorder()

	// Create a generic error
	err := errors.New("some internal error")
	message := "Internal server error"

	// Call handleError with the ResponseRecorder, error, and message
	helpers.HandleError(rr, err, message)

	// Check the status code is 500 Internal Server Error
	if status := rr.Code; status != http.StatusInternalServerError {
		t.Errorf("Expected status code %d, got %d", http.StatusInternalServerError, status)
	}

	// Check the response body is the provided message followed by a newline
	expectedBody := message + "\n"
	if rr.Body.String() != expectedBody {
		t.Errorf("Expected response body '%s', got '%s'", expectedBody, rr.Body.String())
	}
}
