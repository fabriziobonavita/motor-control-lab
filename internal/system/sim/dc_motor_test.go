package sim

import (
	"math"
	"testing"

	"github.com/fabriziobonavita/motor-control-lab/internal/system"
)

const eps = 1e-9

// TestDCMotor_DisturbanceMath tests the disturbance dynamics:
// dv = alpha*(target - v) - d*dt
// where alpha = dt/tau, target = K*V
func TestDCMotor_DisturbanceMath(t *testing.T) {
	m := NewDCMotor()
	m.GainRPMPerVolt = 100.0
	m.TauSeconds = 0.5
	m.VelocityRPM = 0.0

	dt := 0.001
	voltage := 10.0
	disturbance := 50.0 // RPM/s

	m.Actuate(voltage)
	m.SetDisturbanceRPMPerS(disturbance)
	m.Step(dt)

	// Expected: target = K*V = 100*10 = 1000 RPM
	// alpha = dt/tau = 0.001/0.5 = 0.002
	// dv = alpha*(target - v) - d*dt = 0.002*(1000 - 0) - 50*0.001 = 2.0 - 0.05 = 1.95
	expectedVelocity := 1.95

	if math.Abs(m.VelocityRPM-expectedVelocity) > eps {
		t.Errorf("after first step: VelocityRPM = %v, want %v", m.VelocityRPM, expectedVelocity)
	}

	// Second step: v = 1.95, target = 1000
	// dv = 0.002*(1000 - 1.95) - 50*0.001 = 0.002*998.05 - 0.05 = 1.9961 - 0.05 = 1.9461
	// v_new = 1.95 + 1.9461 = 3.8961
	m.Step(dt)
	expectedVelocity2 := 1.95 + (0.002*(1000.0-1.95) - 50.0*0.001)

	if math.Abs(m.VelocityRPM-expectedVelocity2) > eps {
		t.Errorf("after second step: VelocityRPM = %v, want %v", m.VelocityRPM, expectedVelocity2)
	}
}

// TestDCMotor_SteadyStateWithDisturbance tests that in steady state:
// v = K*V - d*tau
// This is derived from: dv/dt = 0 = (1/tau)*(K*V - v) - d
// Solving: 0 = (1/tau)*(K*V - v) - d
// => d = (1/tau)*(K*V - v)
// => d*tau = K*V - v
// => v = K*V - d*tau
func TestDCMotor_SteadyStateWithDisturbance(t *testing.T) {
	m := NewDCMotor()
	m.GainRPMPerVolt = 100.0
	m.TauSeconds = 0.5
	m.VelocityRPM = 0.0

	voltage := 10.0
	disturbance := 50.0 // RPM/s

	m.Actuate(voltage)
	m.SetDisturbanceRPMPerS(disturbance)

	// Run many steps to reach steady state
	dt := 0.001
	for i := 0; i < 10000; i++ {
		m.Step(dt)
	}

	// Expected steady state: v = K*V - d*tau = 100*10 - 50*0.5 = 1000 - 25 = 975 RPM
	expectedSteadyState := m.GainRPMPerVolt*voltage - disturbance*m.TauSeconds

	if math.Abs(m.VelocityRPM-expectedSteadyState) > 0.1 {
		t.Errorf("steady state VelocityRPM = %v, want %v (within 0.1 RPM)", m.VelocityRPM, expectedSteadyState)
	}
}

// TestDCMotor_DisturbanceZero verifies that zero disturbance doesn't change behavior
func TestDCMotor_DisturbanceZero(t *testing.T) {
	m1 := NewDCMotor()
	m2 := NewDCMotor()

	m1.GainRPMPerVolt = 100.0
	m1.TauSeconds = 0.5
	m1.VelocityRPM = 0.0

	m2.GainRPMPerVolt = 100.0
	m2.TauSeconds = 0.5
	m2.VelocityRPM = 0.0

	voltage := 10.0
	dt := 0.001

	m1.Actuate(voltage)
	m1.SetDisturbanceRPMPerS(0.0) // Explicitly set to zero

	m2.Actuate(voltage)
	// m2 doesn't set disturbance (defaults to zero)

	// Run both for many steps
	for i := 0; i < 1000; i++ {
		m1.Step(dt)
		m2.Step(dt)
	}

	// Both should have the same velocity
	if math.Abs(m1.VelocityRPM-m2.VelocityRPM) > eps {
		t.Errorf("with zero disturbance: m1.VelocityRPM = %v, m2.VelocityRPM = %v, should be equal",
			m1.VelocityRPM, m2.VelocityRPM)
	}
}

// TestDCMotor_DisturbanceReceiverInterface verifies DCMotor implements the interface
func TestDCMotor_DisturbanceReceiverInterface(t *testing.T) {
	m := NewDCMotor()
	// This will fail at compile time if DCMotor doesn't implement DisturbanceReceiver
	var _ system.DisturbanceReceiver = m
}
