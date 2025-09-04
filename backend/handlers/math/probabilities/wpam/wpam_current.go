package wpam

func GetCurrentProbability(probChanges []ProbabilityChange) float64 {

	return probChanges[len(probChanges)-1].Probability
}
