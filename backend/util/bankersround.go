package util

import "math"

// BankersRound rounds x to the nearest integer using banker's rounding
// (round-half-to-even), also called "unbiased rounding".
//
// Unlike math.Round (which rounds 0.5 away from zero), BankersRound rounds
// ties to the nearest even integer, eliminating systematic bias in financial
// calculations where many values fall on the 0.5 boundary.
//
// Examples:
//
//	BankersRound(0.5)  → 0   (0 is even)
//	BankersRound(1.5)  → 2   (2 is even)
//	BankersRound(2.5)  → 2   (2 is even)
//	BankersRound(3.5)  → 4   (4 is even)
//	BankersRound(-0.5) → 0   (0 is even)
//	BankersRound(-1.5) → -2  (-2 is even)
func BankersRound(x float64) float64 {
	if math.IsNaN(x) || math.IsInf(x, 0) {
		return x
	}

	floor := math.Floor(x)
	frac := x - floor

	const half = 0.5
	const eps = 1e-10 // guard against floating-point representation error

	if frac < half-eps {
		return floor
	}
	if frac > half+eps {
		return floor + 1
	}
	// Exactly 0.5 (within epsilon): round to nearest even
	if math.Mod(floor, 2) == 0 {
		return floor
	}
	return floor + 1
}
