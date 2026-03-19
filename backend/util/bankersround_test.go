package util

import (
	"math"
	"testing"
)

func TestBankersRound(t *testing.T) {
	tests := []struct {
		input    float64
		expected float64
	}{
		// Non-half values — same as math.Round
		{0.0, 0},
		{0.1, 0},
		{0.9, 1},
		{1.0, 1},
		{1.1, 1},
		{1.9, 2},
		{-0.1, 0},
		{-0.9, -1},
		{-1.1, -1},
		// Half values — round to even
		{0.5, 0},  // 0 is even → stay at 0
		{1.5, 2},  // 2 is even → go to 2
		{2.5, 2},  // 2 is even → stay at 2
		{3.5, 4},  // 4 is even → go to 4
		{4.5, 4},  // 4 is even → stay at 4
		{5.5, 6},  // 6 is even → go to 6
		{-0.5, 0}, // 0 is even → go to 0
		{-1.5, -2}, // -2 is even → stay at -2
		{-2.5, -2}, // -2 is even → go to -2 (i.e. floor(-2.5)+1 = -1+1 = -2... wait
		// Let me recheck: floor(-2.5) = -3, frac = -2.5 - (-3) = 0.5, floor=-3 is odd → return -3+1=-2. Correct.
		{-3.5, -4}, // floor(-3.5)=-4, frac=0.5, -4 is even → return -4. Correct.
		// Special values
		{math.Inf(1), math.Inf(1)},
		{math.Inf(-1), math.Inf(-1)},
	}

	for _, tt := range tests {
		result := BankersRound(tt.input)
		if result != tt.expected {
			t.Errorf("BankersRound(%v) = %v, want %v", tt.input, result, tt.expected)
		}
	}
}

func TestBankersRound_NaN(t *testing.T) {
	result := BankersRound(math.NaN())
	if !math.IsNaN(result) {
		t.Errorf("BankersRound(NaN) should return NaN, got %v", result)
	}
}

// Verify the key difference from math.Round: 0.5 rounds to 0 (even), not 1
func TestBankersRound_DifferentFromMathRound(t *testing.T) {
	if math.Round(0.5) == 0 {
		t.Skip("math.Round already uses banker's rounding on this platform")
	}
	// 0.5 → math.Round gives 1, BankersRound gives 0
	if BankersRound(0.5) != 0 {
		t.Errorf("BankersRound(0.5) should be 0 (nearest even), got %v", BankersRound(0.5))
	}
	// 2.5 → math.Round gives 3, BankersRound gives 2
	if BankersRound(2.5) != 2 {
		t.Errorf("BankersRound(2.5) should be 2 (nearest even), got %v", BankersRound(2.5))
	}
}
