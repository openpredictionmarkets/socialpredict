package errors

import (
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
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
			output:         "",
			wrappedMessage: "{\"error\":\"bar\"}",
			res:            true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			oldStdout := os.Stdout
			readPipe, writePipe, err := os.Pipe()
			if err != nil {
				t.Fatalf("pipe stdout: %v", err)
			}
			os.Stdout = writePipe
			res := HandleHTTPError(test.w, test.err, test.statusCode, test.userMessage)
			_ = writePipe.Close()
			os.Stdout = oldStdout

			var buf bytes.Buffer
			if _, err := buf.ReadFrom(readPipe); err != nil {
				t.Fatalf("read stdout: %v", err)
			}
			_ = readPipe.Close()

			if test.res != res {
				t.Errorf("got %t, want %t", test.res, res)
			}
			if test.w.Code != test.statusCode {
				t.Errorf("got %d, want %d", test.statusCode, test.w.Code)
			}
			if !strings.Contains(test.w.Body.String(), test.wrappedMessage) {
				t.Errorf("got %s, want %s", strings.TrimRight(test.w.Body.String(), "\n"), test.wrappedMessage)
			}
			if test.output != "" && !strings.Contains(buf.String(), test.output) {
				t.Errorf("got Error%s, want %s", strings.SplitN(strings.TrimRight(buf.String(), "\n"), "Error", 1)[0], test.output)
			}
		})

	}
}
