package wpam

import (
	"testing"
	"time"
)

func TestGetCurrentProbability(t *testing.T) {
	now := time.Now()

	t.Run("returns last probability from multiple entries", func(t *testing.T) {
		probChanges := []ProbabilityChange{
			{Probability: 0.5, Timestamp: now},
			{Probability: 0.6, Timestamp: now.Add(time.Minute)},
			{Probability: 0.7, Timestamp: now.Add(2 * time.Minute)},
		}

		prob := GetCurrentProbability(probChanges)
		if prob != 0.7 {
			t.Errorf("expected 0.7, got %f", prob)
		}
	})

	t.Run("returns only probability in single entry", func(t *testing.T) {
		probChanges := []ProbabilityChange{
			{Probability: 0.42, Timestamp: now},
		}

		prob := GetCurrentProbability(probChanges)
		if prob != 0.42 {
			t.Errorf("expected 0.42, got %f", prob)
		}
	})

	// 🚫 No test for empty input — we are okay with panics in this case
}
