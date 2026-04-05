package wpam

type CurrentProbabilitySelector interface {
	Select([]ProbabilityChange) float64
}

type trailingProbabilitySelector struct{}

var defaultProbabilitySelector CurrentProbabilitySelector = trailingProbabilitySelector{}

func GetCurrentProbability(probChanges []ProbabilityChange) float64 {
	return GetCurrentProbabilityWithSelector(probChanges, defaultProbabilitySelector)
}

func GetCurrentProbabilityWithSelector(probChanges []ProbabilityChange, selector CurrentProbabilitySelector) float64 {
	if selector == nil {
		selector = defaultProbabilitySelector
	}
	return selector.Select(probChanges)
}

func (trailingProbabilitySelector) Select(probChanges []ProbabilityChange) float64 {
	if len(probChanges) == 0 {
		return 0
	}
	return probChanges[len(probChanges)-1].Probability
}
