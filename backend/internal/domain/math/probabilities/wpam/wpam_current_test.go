package wpam

import (
	"testing"
	"time"
)

type fixedSelector struct{ value float64 }

func (s fixedSelector) Select([]ProbabilityChange) float64 { return s.value }

var currentProbabilityBaseTime = time.Date(2025, 1, 1, 13, 0, 0, 0, time.UTC)

func makeProbabilityChanges(values ...float64) []ProbabilityChange {
	changes := make([]ProbabilityChange, 0, len(values))
	for i, value := range values {
		changes = append(changes, ProbabilityChange{
			Probability: value,
			Timestamp:   currentProbabilityBaseTime.Add(time.Duration(i) * time.Minute),
		})
	}
	return changes
}

func TestGetCurrentProbability(t *testing.T) {
	tests := []struct {
		name    string
		changes []ProbabilityChange
		want    float64
	}{
		{
			name:    "returns last probability from multiple entries",
			changes: makeProbabilityChanges(0.5, 0.6, 0.7),
			want:    0.7,
		},
		{
			name:    "returns only probability in single entry",
			changes: makeProbabilityChanges(0.42),
			want:    0.42,
		},
		{
			name:    "returns 0 for empty input",
			changes: nil,
			want:    0,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := GetCurrentProbability(test.changes); got != test.want {
				t.Fatalf("expected %f, got %f", test.want, got)
			}
		})
	}
}

func TestGetCurrentProbabilityWithSelector(t *testing.T) {
	if got := GetCurrentProbabilityWithSelector(nil, fixedSelector{value: 0.91}); got != 0.91 {
		t.Fatalf("expected injected probability 0.91, got %f", got)
	}
}
