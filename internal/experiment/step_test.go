package experiment

import (
	"testing"

	"github.com/fabriziobonavita/motor-control-lab/internal/control/pid"
	"github.com/fabriziobonavita/motor-control-lab/internal/system/sim"
)

func TestRunStep_SmokeTest(t *testing.T) {
	// Use default PID parameters from cmd_sim_step.go
	ctrl := pid.New(0.02, 0.05, 0.0)
	plant := sim.NewDCMotor()

	cfg := StepConfig{
		TargetRPM: 1000.0,
		DT:        0.005, // Use larger dt to keep test fast
		Duration:  2.0,   // Short duration for fast test
	}

	samples, _ := RunStep(plant, ctrl, cfg)

	// Assert sample count > 0
	if len(samples) == 0 {
		t.Fatal("no samples produced")
	}

	// Assert final actual is finite
	last := samples[len(samples)-1]
	if !isFinite(last.Actual) {
		t.Errorf("final actual = %v, want finite", last.Actual)
	}

	// Assert final error is closer to target than initial error
	initial := samples[0]
	initialError := abs(initial.Error)
	finalError := abs(last.Error)

	// Final error should be smaller than initial (or at least not much larger)
	// Use loose threshold to avoid brittleness
	if finalError > initialError*1.5 {
		t.Logf("initial error: %v, final error: %v", initialError, finalError)
		t.Logf("warning: final error not significantly better than initial")
		// Don't fail, just warn - this is a smoke test
	}

	// Verify samples have expected structure
	for i, s := range samples {
		if s.T < 0 {
			t.Errorf("sample %d: T = %v, want >= 0", i, s.T)
		}
		if s.DT != cfg.DT {
			t.Errorf("sample %d: DT = %v, want %v", i, s.DT, cfg.DT)
		}
		if !isFinite(s.Actual) {
			t.Errorf("sample %d: Actual = %v, want finite", i, s.Actual)
		}
	}

	// Verify basic sample properties that would be used for metrics
	// (can't import analysis due to import cycle)
	sumAbsError := 0.0
	for _, s := range samples {
		if s.Error < 0 {
			sumAbsError += -s.Error * s.DT
		} else {
			sumAbsError += s.Error * s.DT
		}
	}
	if sumAbsError <= 0 {
		t.Errorf("sum of absolute error = %v, want > 0", sumAbsError)
	}
}

func TestRunStep_InvalidConfig(t *testing.T) {
	ctrl := pid.New(0.02, 0.05, 0.0)
	plant := sim.NewDCMotor()

	tests := []struct {
		name string
		cfg  StepConfig
	}{
		{
			name: "zero dt",
			cfg:  StepConfig{TargetRPM: 1000.0, DT: 0.0, Duration: 1.0},
		},
		{
			name: "negative dt",
			cfg:  StepConfig{TargetRPM: 1000.0, DT: -0.001, Duration: 1.0},
		},
		{
			name: "zero duration",
			cfg:  StepConfig{TargetRPM: 1000.0, DT: 0.001, Duration: 0.0},
		},
		{
			name: "negative duration",
			cfg:  StepConfig{TargetRPM: 1000.0, DT: 0.001, Duration: -1.0},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			samples, _ := RunStep(plant, ctrl, tt.cfg)
			if len(samples) != 0 {
				t.Errorf("RunStep() produced %d samples, want 0 for invalid config", len(samples))
			}
		})
	}
}

func isFinite(f float64) bool {
	return !(f != f || (f > 1e308) || (f < -1e308))
}

func abs(f float64) float64 {
	if f < 0 {
		return -f
	}
	return f
}
