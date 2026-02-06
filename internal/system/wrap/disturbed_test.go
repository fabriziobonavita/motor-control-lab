package wrap

import (
	"math"
	"testing"

	"github.com/fabriziobonavita/motor-control-lab/internal/system"
	"github.com/fabriziobonavita/motor-control-lab/internal/system/sim"
)

const eps = 1e-9

// mockSystem is a simple system for testing the wrapper
type mockSystem struct {
	observed float64
	actuated float64
	stepped  bool
}

func (m *mockSystem) Observe() float64  { return m.observed }
func (m *mockSystem) Actuate(u float64) { m.actuated = u }
func (m *mockSystem) Step(dt float64)   { m.stepped = true }

// mockDisturbanceReceiver extends mockSystem to support disturbance
type mockDisturbanceReceiver struct {
	mockSystem
	disturbance float64
}

func (m *mockDisturbanceReceiver) SetDisturbanceRPMPerS(d float64) {
	m.disturbance = d
}

func TestDisturbedSystem_Delegation(t *testing.T) {
	mock := &mockSystem{observed: 100.0}
	cfg := StepDisturbanceConfig{Enabled: false}
	wrapper := NewDisturbedSystem(mock, cfg)

	// Test Observe delegation
	if got := wrapper.Observe(); got != 100.0 {
		t.Errorf("Observe() = %v, want 100.0", got)
	}

	// Test Actuate delegation
	wrapper.Actuate(5.0)
	if mock.actuated != 5.0 {
		t.Errorf("Actuate() did not delegate, got %v", mock.actuated)
	}

	// Test Step delegation
	mock.stepped = false
	wrapper.Step(0.001)
	if !mock.stepped {
		t.Error("Step() did not delegate to inner system")
	}
}

func TestDisturbedSystem_DisturbanceInjection(t *testing.T) {
	mock := &mockDisturbanceReceiver{}
	cfg := StepDisturbanceConfig{
		Enabled:          true,
		StartS:           0.5,
		DurationS:        2.0,
		MagnitudeRPMPerS: 10.0,
	}
	wrapper := NewDisturbedSystem(mock, cfg)

	// Before disturbance starts
	wrapper.Step(0.1) // t = 0.1
	if math.Abs(mock.disturbance) > eps {
		t.Errorf("disturbance before start = %v, want 0", mock.disturbance)
	}

	// At disturbance start
	wrapper.Step(0.4) // t = 0.5
	if math.Abs(mock.disturbance-10.0) > eps {
		t.Errorf("disturbance at start = %v, want 10.0", mock.disturbance)
	}

	// During disturbance
	wrapper.Step(0.5) // t = 1.0
	if math.Abs(mock.disturbance-10.0) > eps {
		t.Errorf("disturbance during = %v, want 10.0", mock.disturbance)
	}

	// After disturbance ends
	wrapper.Step(1.6) // t = 2.6
	if math.Abs(mock.disturbance) > eps {
		t.Errorf("disturbance after end = %v, want 0", mock.disturbance)
	}
}

func TestDisturbedSystem_SignalReporter(t *testing.T) {
	mock := &mockDisturbanceReceiver{}
	cfg := StepDisturbanceConfig{
		Enabled:          true,
		StartS:           0.5,
		MagnitudeRPMPerS: 10.0,
	}
	wrapper := NewDisturbedSystem(mock, cfg)

	// Before disturbance
	wrapper.Step(0.1)
	sigs := wrapper.Signals()
	if got, ok := sigs["disturbance_rpm_per_s"]; !ok || math.Abs(got) > eps {
		t.Errorf("Signals()[\"disturbance_rpm_per_s\"] before start = %v, want 0", got)
	}

	// During disturbance
	wrapper.Step(0.4) // t = 0.5
	sigs = wrapper.Signals()
	if got, ok := sigs["disturbance_rpm_per_s"]; !ok || math.Abs(got-10.0) > eps {
		t.Errorf("Signals()[\"disturbance_rpm_per_s\"] during = %v, want 10.0", got)
	}

	// Verify it implements the interface
	var _ system.SignalReporter = wrapper
}

func TestDisturbedSystem_ResetTime(t *testing.T) {
	mock := &mockDisturbanceReceiver{}
	cfg := StepDisturbanceConfig{
		Enabled:          true,
		StartS:           1.0,
		MagnitudeRPMPerS: 10.0,
	}
	wrapper := NewDisturbedSystem(mock, cfg)

	// Step past disturbance start
	wrapper.Step(1.5) // t = 1.5
	if math.Abs(mock.disturbance-10.0) > eps {
		t.Error("disturbance should be active")
	}

	// Reset time
	wrapper.ResetTime()

	// Step again - should be before disturbance start now
	wrapper.Step(0.5) // t = 0.5
	if math.Abs(mock.disturbance) > eps {
		t.Errorf("after ResetTime, disturbance = %v, want 0", mock.disturbance)
	}
}

func TestDisturbedSystem_WithDCMotor(t *testing.T) {
	motor := sim.NewDCMotor()
	motor.VelocityRPM = 0.0

	cfg := StepDisturbanceConfig{
		Enabled:          true,
		StartS:           0.0,
		MagnitudeRPMPerS: 50.0,
	}
	wrapper := NewDisturbedSystem(motor, cfg)

	// Step with disturbance
	wrapper.Step(0.001)

	// Motor should have received disturbance
	// Check that velocity is affected (disturbance should reduce it)
	// Without disturbance, motor at rest with no input should stay at 0
	// With disturbance, it should go negative
	if motor.VelocityRPM >= 0 {
		t.Errorf("motor velocity with disturbance = %v, want negative", motor.VelocityRPM)
	}

	// Verify signal reporter works
	sigs := wrapper.Signals()
	if got, ok := sigs["disturbance_rpm_per_s"]; !ok || math.Abs(got-50.0) > eps {
		t.Errorf("Signals()[\"disturbance_rpm_per_s\"] = %v, want 50.0", got)
	}
}

// Test computeDisturbance function directly (it's not exported, but we can test via wrapper)
func TestComputeDisturbance_Behavior(t *testing.T) {
	tests := []struct {
		name     string
		steps    []float64 // sequence of dt values to step
		cfg      StepDisturbanceConfig
		expected []float64 // expected disturbance after each step
	}{
		{
			name:     "disabled",
			steps:    []float64{0.1, 0.5, 1.0},
			cfg:      StepDisturbanceConfig{Enabled: false, StartS: 0.5, MagnitudeRPMPerS: 10.0},
			expected: []float64{0.0, 0.0, 0.0},
		},
		{
			name:     "before start",
			steps:    []float64{0.1, 0.2},
			cfg:      StepDisturbanceConfig{Enabled: true, StartS: 0.5, MagnitudeRPMPerS: 10.0},
			expected: []float64{0.0, 0.0},
		},
		{
			name:     "at start",
			steps:    []float64{0.4, 0.1}, // t goes from 0.4 to 0.5
			cfg:      StepDisturbanceConfig{Enabled: true, StartS: 0.5, MagnitudeRPMPerS: 10.0},
			expected: []float64{0.0, 10.0},
		},
		{
			name:     "infinite duration",
			steps:    []float64{0.5, 0.5, 0.5}, // t: 0.5, 1.0, 1.5
			cfg:      StepDisturbanceConfig{Enabled: true, StartS: 0.5, DurationS: 0.0, MagnitudeRPMPerS: 10.0},
			expected: []float64{10.0, 10.0, 10.0},
		},
		{
			name:     "finite duration",
			steps:    []float64{0.5, 0.5, 0.5, 0.5, 0.5}, // compute at: 0.5, 1.0, 1.5, 2.0, 2.5
			cfg:      StepDisturbanceConfig{Enabled: true, StartS: 0.5, DurationS: 2.0, MagnitudeRPMPerS: 10.0},
			expected: []float64{10.0, 10.0, 10.0, 10.0, 0.0}, // last step computes at t=2.5 which is at/after end (0.5+2.0=2.5)
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockDisturbanceReceiver{}
			wrapper := NewDisturbedSystem(mock, tt.cfg)
			for i, dt := range tt.steps {
				wrapper.Step(dt)
				expected := tt.expected[i]
				if math.Abs(mock.disturbance-expected) > eps {
					t.Errorf("step %d: disturbance = %v, want %v", i, mock.disturbance, expected)
				}
			}
		})
	}
}
