package modifier

import (
	"math"
	"testing"
)

const eps = 1e-9

func TestDeadzone(t *testing.T) {
	tests := []struct {
		name      string
		threshold float64
		input     float64
		want      float64
	}{
		{
			name:      "below threshold positive",
			threshold: 1.0,
			input:     0.5,
			want:      0.0,
		},
		{
			name:      "below threshold negative",
			threshold: 1.0,
			input:     -0.5,
			want:      0.0,
		},
		{
			name:      "at threshold positive",
			threshold: 1.0,
			input:     1.0,
			want:      0.0,
		},
		{
			name:      "at threshold negative",
			threshold: 1.0,
			input:     -1.0,
			want:      0.0,
		},
		{
			name:      "above threshold positive",
			threshold: 1.0,
			input:     3.0,
			want:      2.0, // 3 - 1
		},
		{
			name:      "above threshold negative",
			threshold: 1.0,
			input:     -3.0,
			want:      -2.0, // -(3 - 1)
		},
		{
			name:      "zero threshold",
			threshold: 0.0,
			input:     5.0,
			want:      5.0,
		},
		{
			name:      "zero threshold negative",
			threshold: 0.0,
			input:     -5.0,
			want:      -5.0,
		},
		{
			name:      "large threshold small positive input",
			threshold: 10000.0,
			input:     5.0,
			want:      0.0,
		},
		{
			name:      "large threshold small negative input",
			threshold: 10000.0,
			input:     -5.0,
			want:      0.0,
		},
		{
			name:      "large threshold large positive input",
			threshold: 10000.0,
			input:     15000.0,
			want:      5000.0, // 15000 - 10000
		},
		{
			name:      "large threshold large negative input",
			threshold: 10000.0,
			input:     -15000.0,
			want:      -5000.0, // -(15000 - 10000)
		},
		{
			name:      "exactly at negative threshold",
			threshold: 2.5,
			input:     -2.5,
			want:      0.0,
		},
		{
			name:      "just above negative threshold",
			threshold: 2.5,
			input:     -2.6,
			want:      -0.1, // -2.6 + 2.5 = -0.1 (but should be -(2.6 - 2.5) = -0.1)
		},
		{
			name:      "very small threshold",
			threshold: 0.001,
			input:     0.0005,
			want:      0.0,
		},
		{
			name:      "very small threshold negative",
			threshold: 0.001,
			input:     -0.0005,
			want:      0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dz := &DeadzoneModifier{Threshold: tt.threshold}
			got := dz.Modify(tt.input)
			if math.Abs(got-tt.want) > eps {
				t.Errorf("Modify(%v) with threshold %v = %v, want %v", tt.input, tt.threshold, got, tt.want)
			}
		})
	}
}

// TestDeadzoneSymmetry verifies that deadzone is symmetric around zero
func TestDeadzoneSymmetry(t *testing.T) {
	threshold := 1.5
	dz := &DeadzoneModifier{Threshold: threshold}

	// For any positive value, the negative should produce the negative result
	testCases := []float64{0.5, 1.0, 1.5, 2.0, 5.0, 10.0}

	for _, posVal := range testCases {
		posResult := dz.Modify(posVal)
		negResult := dz.Modify(-posVal)

		// Results should be symmetric (opposite signs, same magnitude)
		expectedNegResult := -posResult
		if math.Abs(negResult-expectedNegResult) > eps {
			t.Errorf("Asymmetry detected: Modify(%v) = %v, Modify(%v) = %v (expected %v)",
				posVal, posResult, -posVal, negResult, expectedNegResult)
		}
	}
}

// TestDeadzoneAbsoluteValue verifies deadzone uses absolute value correctly
func TestDeadzoneAbsoluteValue(t *testing.T) {
	threshold := 2.0
	dz := &DeadzoneModifier{Threshold: threshold}

	// Inputs with same absolute value should produce same absolute output
	tests := []struct {
		input1, input2 float64
		description    string
	}{
		{1.0, -1.0, "both below threshold"},
		{3.0, -3.0, "both above threshold"},
		{2.0, -2.0, "both at threshold"},
	}

	for _, tt := range tests {
		result1 := dz.Modify(tt.input1)
		result2 := dz.Modify(tt.input2)

		absResult1 := math.Abs(result1)
		absResult2 := math.Abs(result2)

		if math.Abs(absResult1-absResult2) > eps {
			t.Errorf("%s: Modify(%v) = %v, Modify(%v) = %v (absolute values should match)",
				tt.description, tt.input1, result1, tt.input2, result2)
		}

		// Results should be opposite signs (unless both are zero)
		if result1 != 0 && result2 != 0 {
			if (result1 > 0 && result2 > 0) || (result1 < 0 && result2 < 0) {
				t.Errorf("%s: Modify(%v) = %v, Modify(%v) = %v (should have opposite signs)",
					tt.description, tt.input1, result1, tt.input2, result2)
			}
		}
	}
}

func TestChain(t *testing.T) {
	// Test chaining multiple deadzones (shouldn't happen in practice, but tests composition)
	dz1 := &DeadzoneModifier{Threshold: 1.0}
	dz2 := &DeadzoneModifier{Threshold: 0.5}

	chain := Chain(dz1, dz2)

	// Input 3.0 -> after dz1 (1.0) = 2.0 -> after dz2 (0.5) = 1.5
	got := chain.Modify(3.0)
	want := 1.5
	if math.Abs(got-want) > eps {
		t.Errorf("Chain.Modify(3.0) = %v, want %v", got, want)
	}

	// Test negative input in chain
	gotNeg := chain.Modify(-3.0)
	wantNeg := -1.5
	if math.Abs(gotNeg-wantNeg) > eps {
		t.Errorf("Chain.Modify(-3.0) = %v, want %v", gotNeg, wantNeg)
	}

	// Test empty chain
	emptyChain := Chain()
	if emptyChain == nil {
		t.Error("Chain() with no args should not return nil (should return no-op chain)")
	}
	result := emptyChain.Modify(3.0)
	if result != 3.0 {
		t.Errorf("Empty chain should pass through value, got %v, want 3.0", result)
	}

	// Test single modifier
	single := Chain(dz1)
	got2 := single.Modify(3.0)
	want2 := 2.0
	if math.Abs(got2-want2) > eps {
		t.Errorf("Chain(single).Modify(3.0) = %v, want %v", got2, want2)
	}

	// Test chain order matters
	dz3 := &DeadzoneModifier{Threshold: 2.0}
	dz4 := &DeadzoneModifier{Threshold: 1.0}
	chain1 := Chain(dz3, dz4) // Apply 2.0 first, then 1.0
	chain2 := Chain(dz4, dz3) // Apply 1.0 first, then 2.0

	input := 5.0
	// chain1: 5.0 -> (5.0-2.0=3.0) -> (3.0-1.0=2.0)
	// chain2: 5.0 -> (5.0-1.0=4.0) -> (4.0-2.0=2.0)
	// Actually both should give same result in this case, but order matters in general
	result1 := chain1.Modify(input)
	result2 := chain2.Modify(input)
	if math.Abs(result1-result2) > eps {
		t.Logf("Chain order test: chain(dz3,dz4).Modify(%v) = %v, chain(dz4,dz3).Modify(%v) = %v",
			input, result1, input, result2)
	}
}

// TestChainWithNegativeValues tests chain behavior with negative inputs
func TestChainWithNegativeValues(t *testing.T) {
	dz1 := &DeadzoneModifier{Threshold: 1.0}
	dz2 := &DeadzoneModifier{Threshold: 0.5}
	chain := Chain(dz1, dz2)

	// Negative input: -3.0 -> after dz1 -> after dz2
	got := chain.Modify(-3.0)
	// Expected: -3.0 -> (if |u| < 1.0 then 0, else sign(u)*(|u|-1.0)) = -2.0
	// Then -2.0 -> (if |u| < 0.5 then 0, else sign(u)*(|u|-0.5)) = -1.5
	want := -1.5
	if math.Abs(got-want) > eps {
		t.Errorf("Chain.Modify(-3.0) = %v, want %v", got, want)
	}
}
