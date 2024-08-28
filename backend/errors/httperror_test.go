package errors

import (
	"bytes"
	"errors"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHandleHTTPError(t *testing.T) {
	tests := []struct {
		name           string
		w              *httptest.ResponseRecorder
		err            error
		statusCode     int
		userMessage    string
		output         string
		wrappedMessage string
		res            bool
	}{
		{
			name:           "Nil",
			w:              httptest.NewRecorder(),
			err:            nil,
			statusCode:     http.StatusOK,
			userMessage:    "",
			output:         "",
			wrappedMessage: "",
			res:            false,
		},
		{
			name:           "Server Error 500",
			w:              httptest.NewRecorder(),
			err:            errors.New("foo"),
			statusCode:     http.StatusInternalServerError,
			userMessage:    "bar",
			output:         "Error: foo",
			wrappedMessage: "{\"error\":\"bar\"}",
			res:            true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			currentWriter := log.Writer()
			var buf bytes.Buffer
			log.SetOutput(&buf)
			res := HandleHTTPError(test.w, test.err, test.statusCode, test.userMessage)
			log.SetOutput(currentWriter)
			if test.res != res {
				t.Errorf("got %t, want %t", test.res, res)
			}
			if test.w.Code != test.statusCode {
				t.Errorf("got %d, want %d", test.statusCode, test.w.Code)
			}
			if !strings.Contains(test.w.Body.String(), test.wrappedMessage) {
				t.Errorf("got %s, want %s", strings.TrimRight(test.w.Body.String(), "\n"), test.wrappedMessage)
			}
			if !strings.Contains(buf.String(), test.output) {
				t.Errorf("got Error%s, want %s", strings.SplitN(strings.TrimRight(buf.String(), "\n"), "Error", 1)[0], test.output)
			}
		})

	}
}
