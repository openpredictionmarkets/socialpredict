package errors

import (
	"bytes"
	"errors"
	"log"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHandleHTTPError(t *testing.T) {
	tests := []struct {
		testName       string
		w              *httptest.ResponseRecorder
		err            error
		statusCode     int
		userMessage    string
		output         string
		wrappedMessage string
		res            bool
	}{
		{
			testName:       "Nil",
			w:              httptest.NewRecorder(),
			err:            nil,
			statusCode:     200,
			userMessage:    "",
			output:         "",
			wrappedMessage: "",
			res:            false,
		},
		{
			testName:       "Server Error 500",
			w:              httptest.NewRecorder(),
			err:            errors.New("foo"),
			statusCode:     500,
			userMessage:    "bar",
			output:         "Error: foo",
			wrappedMessage: "{\"error\":\"bar\"}",
			res:            true,
		},
	}
	for _, test := range tests {
		currentWriter := log.Writer()
		var buf bytes.Buffer
		log.SetOutput(&buf)
		res := HandleHTTPError(test.w, test.err, test.statusCode, test.userMessage)
		log.SetOutput(currentWriter)
		if test.res != res {
			t.Errorf("%s: expected %t, got %t", test.testName, test.res, res)
		}
		if test.w.Code != test.statusCode {
			t.Errorf("%s: expected %d, got %d", test.testName, test.statusCode, test.w.Code)
		}
		if !strings.Contains(test.w.Body.String(), test.wrappedMessage) {
			t.Errorf("%s: expected %s, got %s", test.testName, strings.TrimRight(test.w.Body.String(), "\n"), test.wrappedMessage)
		}
		if !strings.Contains(buf.String(), test.output) {
			t.Errorf("%s: expected Error%s, got %s", test.testName, strings.SplitN(strings.TrimRight(buf.String(), "\n"), "Error", 1)[0], test.output)
		}
	}
}
