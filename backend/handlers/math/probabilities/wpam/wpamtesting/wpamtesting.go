package wpamtesting

import (
	"socialpredict/handlers/math/probabilities/wpam"
	"time"
)

// GenerateProbability creates wpam.ProbabilityChange points with timestamps
func GenerateProbability(probabilities ...float64) []wpam.ProbabilityChange {
	var changes []wpam.ProbabilityChange
	for _, p := range probabilities {
		changes = append(changes, wpam.ProbabilityChange{
			Probability: p,
			Timestamp:   time.Now(),
		})
	}
	return changes
}
