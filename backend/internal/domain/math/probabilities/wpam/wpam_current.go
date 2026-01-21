package wpam

func GetCurrentProbability(probChanges []ProbabilityChange) float64 {
	if len(probChanges) == 0 {
		return 0
	}
	return probChanges[len(probChanges)-1].Probability
}
