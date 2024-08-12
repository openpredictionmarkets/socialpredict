package marketshandlers

import (
	"strings"
	"testing"
)

// TestCheckQuestionTitleLength_invalid tests the question titles that should generate an error
func TestCheckQuestionTitleLength_invalid(t *testing.T) {
	tests := []struct {
		testname      string
		questionTitle string
	}{
		{
			testname:      "TitleExceedsLength",
			questionTitle: strings.Repeat("a", maxQuestionTitleLength+1),
		},
		{
			testname:      "EmptyTitle",
			questionTitle: "",
		},
	}
	for _, test := range tests {
		err := checkQuestionTitleLength(test.questionTitle)
		if err == nil {
			t.Errorf("Expected error in test %s", test.testname)
		}
	}
}

func TestCheckQuestionTitleLength_valid(t *testing.T) {
	tests := []struct {
		testname      string
		questionTitle string
	}{
		{
			testname:      "Single character title",
			questionTitle: "a",
		},
		{
			testname:      "Max length title",
			questionTitle: strings.Repeat("a", maxQuestionTitleLength),
		},
	}
	for _, test := range tests {
		err := checkQuestionTitleLength(test.questionTitle)
		if err != nil {
			t.Errorf("Unexpected error in test %s", test.testname)
		}
	}
}
