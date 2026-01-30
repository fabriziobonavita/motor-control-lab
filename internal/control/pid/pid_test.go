package pid

import (
	"math"
	"testing"
)

const eps = 1e-9

func TestOutputClamping(t *testing.T) {
	tests := []struct {
		name   string
		kp     float64
		target float64
		actual float64
		want   float64 // expected clamped output
	}{
		{
			name:   "saturates high",
			kp:     100.0, // large gain
			target: 1.0,
			actual: 0.0,  // error = 1.0
			want:   24.0, // OutMax
		},
		{
			name:   "saturates low",
			kp:     100.0,
			target: 0.0,
			actual: 1.0,   // error = -1.0
			want:   -24.0, // OutMin
		},
		{
			name:   "no saturation",
			kp:     0.1,
			target: 10.0,
			actual: 0.0, // error = 10.0
			want:   1.0, // 0.1 * 10
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := New(tt.kp, 0, 0)
			ctrl.OutMin = -24.0
			ctrl.OutMax = 24.0

			var tr Trace
			out := ctrl.Step(tt.target, tt.actual, 0.01, &tr)

			if math.Abs(out-tt.want) > eps {
				t.Errorf("Step() = %v, want %v", out, tt.want)
			}

			// Verify clamping
			if out > ctrl.OutMax || out < ctrl.OutMin {
				t.Errorf("output %v outside bounds [%v, %v]", out, ctrl.OutMin, ctrl.OutMax)
			}

			// Verify trace matches
			if math.Abs(tr.Out-out) > eps {
				t.Errorf("trace.Out = %v, want %v", tr.Out, out)
			}
		})
	}
}

func TestDerivativeNoKickOnFirstStep(t *testing.T) {
	ctrl := New(0.1, 0, 1.0) // Kd = 1.0
	ctrl.OutMin = -24.0
	ctrl.OutMax = 24.0

	target := 100.0
	actual := 0.0
	dt := 0.01

	var tr Trace
	out1 := ctrl.Step(target, actual, dt, &tr)

	// On first step, D term should be 0 (no previous error)
	if math.Abs(tr.D) > eps {
		t.Errorf("first step D term = %v, want 0", tr.D)
	}

	// Output should only reflect P term on first step
	expectedP := ctrl.Kp * (target - actual)
	if math.Abs(tr.P-expectedP) > eps {
		t.Errorf("P term = %v, want %v", tr.P, expectedP)
	}

	if math.Abs(out1-tr.P) > eps {
		t.Errorf("first step output = %v, should equal P term %v", out1, tr.P)
	}

	// Second step should have non-zero D term
	actual2 := 10.0
	var tr2 Trace
	ctrl.Step(target, actual2, dt, &tr2)

	if math.Abs(tr2.D) < eps {
		t.Errorf("second step D term = %v, should be non-zero", tr2.D)
	}
}

func TestAntiWindupFreezesIntegralWhenSaturatedSameDirection(t *testing.T) {
	// Use small output max and positive Ki to force saturation
	ctrl := New(0.1, 1.0, 0) // Ki = 1.0
	ctrl.OutMin = -24.0
	ctrl.OutMax = 2.0 // Small max to force saturation

	target := 100.0
	actual := 0.0
	dt := 0.01

	// Track integral over multiple steps
	initialIntegral := ctrl.integral
	var prevIntegral float64

	// Run many steps with large positive error (will saturate high)
	for i := 0; i < 100; i++ {
		var tr Trace
		out := ctrl.Step(target, actual, dt, &tr)

		// After first few steps, output should saturate
		if i > 5 {
			if out < ctrl.OutMax-eps {
				t.Errorf("step %d: output %v should be saturated at %v", i, out, ctrl.OutMax)
			}

			// When saturated high with positive error, integral should freeze
			// (integrated should be false)
			if tr.Error > 0 && out >= ctrl.OutMax-eps {
				if tr.Integrated {
					t.Errorf("step %d: integral updated when saturated high with positive error", i)
				}
			}

			// Integral should not grow unbounded
			if i > 10 {
				// Integral should remain constant or grow much slower
				// In this implementation, it freezes completely when saturated
				if tr.Integrated {
					// If it did integrate, check it's not growing unbounded
					if ctrl.integral > 1000 {
						t.Errorf("step %d: integral %v growing unbounded", i, ctrl.integral)
					}
				} else {
					// If frozen, integral should not change
					if i > 20 && math.Abs(ctrl.integral-prevIntegral) > eps {
						t.Errorf("step %d: integral changed from %v to %v when frozen", i, prevIntegral, ctrl.integral)
					}
				}
			}
		}

		prevIntegral = ctrl.integral
	}

	// Verify integral didn't grow from initial value when frozen
	if math.Abs(ctrl.integral-initialIntegral) < eps {
		t.Logf("integral remained at initial value (frozen), which is correct")
	}
}

func TestTraceFieldsAreConsistent(t *testing.T) {
	ctrl := New(0.02, 0.05, 0.01)
	ctrl.OutMin = -24.0
	ctrl.OutMax = 24.0

	target := 100.0
	actual := 50.0
	dt := 0.01

	var tr Trace
	ctrl.Step(target, actual, dt, &tr)

	// Verify outRaw == P + I + D (within tolerance)
	expectedOutRaw := tr.P + tr.I + tr.D
	if math.Abs(tr.OutRaw-expectedOutRaw) > eps {
		t.Errorf("OutRaw = %v, want P+I+D = %v (P=%v, I=%v, D=%v)",
			tr.OutRaw, expectedOutRaw, tr.P, tr.I, tr.D)
	}

	// Verify outClamped is within bounds
	if tr.Out > ctrl.OutMax || tr.Out < ctrl.OutMin {
		t.Errorf("Out = %v outside bounds [%v, %v]", tr.Out, ctrl.OutMin, ctrl.OutMax)
	}

	// Verify saturated == (outClamped != outRaw)
	expectedSaturated := math.Abs(tr.Out-tr.OutRaw) > eps
	if tr.Saturated != expectedSaturated {
		t.Errorf("Saturated = %v, want %v (Out=%v, OutRaw=%v)",
			tr.Saturated, expectedSaturated, tr.Out, tr.OutRaw)
	}

	// Verify error calculation
	expectedError := target - actual
	if math.Abs(tr.Error-expectedError) > eps {
		t.Errorf("Error = %v, want %v", tr.Error, expectedError)
	}

	// Verify target and actual are preserved
	if math.Abs(tr.Target-target) > eps {
		t.Errorf("Target = %v, want %v", tr.Target, target)
	}
	if math.Abs(tr.Actual-actual) > eps {
		t.Errorf("Actual = %v, want %v", tr.Actual, actual)
	}
}
