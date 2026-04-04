package wpam

type probabilitySelector interface {
	Select([]ProbabilityChange) float64
}

type trailingProbabilitySelector struct{}

var defaultProbabilitySelector probabilitySelector = trailingProbabilitySelector{}

func GetCurrentProbability(probChanges []ProbabilityChange) float64 {
	return defaultProbabilitySelector.Select(probChanges)
}

func (trailingProbabilitySelector) Select(probChanges []ProbabilityChange) float64 {
	if len(probChanges) == 0 {
		return 0
	}
	return probChanges[len(probChanges)-1].Probability
}
