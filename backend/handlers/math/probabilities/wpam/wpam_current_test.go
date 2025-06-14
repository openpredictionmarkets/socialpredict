package wpam

import (
	"testing"
	"time"
)

func TestGetCurrentProbability(t *testing.T) {
	now := time.Now()

	t.Run("returns last probability", func(t *testing.T) {
		probChanges := []ProbabilityChange{
			{Probability: 0.5, Timestamp: now},
			{Probability: 0.6, Timestamp: now.Add(time.Minute)},
			{Probability: 0.7, Timestamp: now.Add(2 * time.Minute)},
		}

		prob = GetCurrentProbability(probChanges)
		if prob != 0.7 {
			t.Errorf("expected 0.7, got %f", prob)
		}
	})

	t.Run("returns only probability in single-entry slice", func(t *testing.T) {
		probChanges := []ProbabilityChange{
			{Probability: 0.42, Timestamp: now},
		}

		prob, err := GetCurrentProbability(probChanges)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if prob != 0.42 {
			t.Errorf("expected 0.42, got %f", prob)
		}
	})

	t.Run("returns error for empty slice", func(t *testing.T) {
		_, err := GetCurrentProbability([]ProbabilityChange{})
		if err == nil {
			t.Fatalf("expected error but got nil")
		}
	})
}
