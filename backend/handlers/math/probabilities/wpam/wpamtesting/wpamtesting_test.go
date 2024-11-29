package wpamtesting

import (
	"testing"
)

// TestGenerateProbability tests the GenerateProbability helper function
func TestGenerateProbability(t *testing.T) {
	// Test data
	inputProbabilities := []float64{0.1, 0.5, 0.9}
	expectedLength := 3

	// Call GenerateProbability with test data
	result := GenerateProbability(inputProbabilities...)

	// Validate the length of the result
	if len(result) != expectedLength {
		t.Fatalf("Expected %d probabilities, got %d", expectedLength, len(result))
	}

	// Validate the probabilities and timestamps in the result
	for i, probabilityChange := range result {
		if probabilityChange.Probability != inputProbabilities[i] {
			t.Errorf("Expected probability %f at index %d, got %f", inputProbabilities[i], i, probabilityChange.Probability)
		}
		if probabilityChange.Timestamp.IsZero() {
			t.Errorf("Timestamp should not be zero at index %d", i)
		}
	}
}
