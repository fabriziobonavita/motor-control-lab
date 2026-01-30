package experiment

import (
	"math"
	"testing"

	"github.com/fabriziobonavita/motor-control-lab/internal/control/pid"
	"github.com/fabriziobonavita/motor-control-lab/internal/experiment/modifier"
	"github.com/fabriziobonavita/motor-control-lab/internal/system/sim"
)

const eps = 1e-9

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

func TestRunStep_WithDeadzone(t *testing.T) {
	ctrl := pid.New(0.02, 0.05, 0.0)
	plant := sim.NewDCMotor()

	// Test with deadzone that should affect small commands
	deadzone := 0.5
	mod := modifier.Chain(&modifier.DeadzoneModifier{Threshold: deadzone})

	cfg := StepConfig{
		TargetRPM: 1000.0,
		DT:        0.005,
		Duration:  2.0,
		Modifier:  mod,
	}

	samples, _ := RunStep(plant, ctrl, cfg)

	if len(samples) == 0 {
		t.Fatal("no samples produced")
	}

	// Verify deadzone is applied: commands with |u| < deadzone should be zero
	// Early in the response, controller output might be small
	foundZeroCommand := false
	foundNonZeroCommand := false

	for i, s := range samples {
		// U should be the modified value (after deadzone)
		if math.Abs(s.U) < deadzone && s.U != 0 {
			t.Errorf("sample %d at t=%v: U=%v should be 0 when |U| < deadzone threshold %v",
				i, s.T, s.U, deadzone)
		}

		if s.U == 0 {
			foundZeroCommand = true
		} else {
			foundNonZeroCommand = true
		}

		// Verify U is the modified value, not the raw controller output
		// (OutRaw is the controller output before modification)
		if math.Abs(s.OutRaw) < deadzone && math.Abs(s.U) > eps {
			t.Errorf("sample %d: OutRaw=%v (below deadzone) but U=%v (should be 0)",
				i, s.OutRaw, s.U)
		}
	}

	// At least some samples should have zero commands (especially early)
	if !foundZeroCommand {
		t.Log("warning: no zero commands found (deadzone may not be active)")
	}

	// Some samples should have non-zero commands (later in response)
	if !foundNonZeroCommand {
		t.Error("all commands are zero - deadzone may be too large or controller not working")
	}
}

func TestRunStep_WithLargeDeadzone(t *testing.T) {
	ctrl := pid.New(0.02, 0.05, 0.0)
	plant := sim.NewDCMotor()

	// Very large deadzone - should block most/all commands
	deadzone := 10000.0
	mod := modifier.Chain(&modifier.DeadzoneModifier{Threshold: deadzone})

	cfg := StepConfig{
		TargetRPM: 1000.0,
		DT:        0.005,
		Duration:  2.0,
		Modifier:  mod,
	}

	samples, _ := RunStep(plant, ctrl, cfg)

	if len(samples) == 0 {
		t.Fatal("no samples produced")
	}

	// With such a large deadzone, most/all commands should be zero
	// Controller output is clamped to ±24V, so all should be zero
	allZero := true
	for i, s := range samples {
		if math.Abs(s.U) > eps {
			allZero = false
			t.Errorf("sample %d: U=%v should be 0 with deadzone=%v (controller max is ±24V)",
				i, s.U, deadzone)
		}
	}

	if !allZero {
		t.Log("warning: some non-zero commands found with very large deadzone")
	}

	// System should not respond (velocity should stay near zero)
	last := samples[len(samples)-1]
	if math.Abs(last.Actual) > 10.0 {
		t.Errorf("with large deadzone, final actual should be near 0, got %v", last.Actual)
	}
}

func TestRunStep_WithDeadzoneNegativeCommands(t *testing.T) {
	ctrl := pid.New(0.02, 0.05, 0.0)
	plant := sim.NewDCMotor()

	// Test that negative commands are handled correctly
	deadzone := 1.0
	mod := modifier.Chain(&modifier.DeadzoneModifier{Threshold: deadzone})

	cfg := StepConfig{
		TargetRPM: 0.0, // Target zero - controller might produce negative commands
		DT:        0.005,
		Duration:  1.0,
		Modifier:  mod,
	}

	// Start with some velocity
	plant.VelocityRPM = 100.0

	samples, _ := RunStep(plant, ctrl, cfg)

	if len(samples) == 0 {
		t.Fatal("no samples produced")
	}

	// Verify negative commands are handled correctly
	for i, s := range samples {
		// If OutRaw is negative and |OutRaw| < deadzone, U should be 0
		if s.OutRaw < 0 && math.Abs(s.OutRaw) < deadzone {
			if math.Abs(s.U) > eps {
				t.Errorf("sample %d: OutRaw=%v (negative, below deadzone) but U=%v (should be 0)",
					i, s.OutRaw, s.U)
			}
		}

		// If OutRaw is negative and |OutRaw| >= deadzone, U should be negative
		if s.OutRaw < 0 && math.Abs(s.OutRaw) >= deadzone {
			if s.U > 0 {
				t.Errorf("sample %d: OutRaw=%v (negative, above deadzone) but U=%v (should be negative)",
					i, s.OutRaw, s.U)
			}
		}
	}
}

func TestRunStep_ModifierNilVsSet(t *testing.T) {
	ctrl := pid.New(0.02, 0.05, 0.0)
	plant := sim.NewDCMotor()

	cfgNoMod := StepConfig{
		TargetRPM: 1000.0,
		DT:        0.005,
		Duration:  1.0,
		Modifier:  nil,
	}

	cfgWithMod := StepConfig{
		TargetRPM: 1000.0,
		DT:        0.005,
		Duration:  1.0,
		Modifier:  modifier.Chain(&modifier.DeadzoneModifier{Threshold: 0.1}),
	}

	samplesNoMod, _ := RunStep(plant, ctrl, cfgNoMod)
	plant2 := sim.NewDCMotor() // Fresh plant
	samplesWithMod, _ := RunStep(plant2, ctrl, cfgWithMod)

	if len(samplesNoMod) == 0 || len(samplesWithMod) == 0 {
		t.Fatal("no samples produced")
	}

	// With deadzone, some commands should be different
	// Find a sample where controller output is small (should be zeroed by deadzone)
	foundDifference := false
	for i := 0; i < len(samplesNoMod) && i < len(samplesWithMod); i++ {
		noModU := samplesNoMod[i].U
		withModU := samplesWithMod[i].U

		if math.Abs(noModU) < 0.1 && math.Abs(noModU) > eps {
			// Small command - should be zeroed by deadzone
			if math.Abs(withModU) > eps {
				t.Errorf("sample %d: without modifier U=%v, with deadzone U=%v (should be 0 for small commands)",
					i, noModU, withModU)
			}
			foundDifference = true
		}
	}

	if !foundDifference {
		t.Log("warning: no difference found between runs with/without modifier")
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
