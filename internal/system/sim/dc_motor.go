package sim

import (
	"math"

	"github.com/fabriziobonavita/motor-control-lab/internal/system"
)

// DCMotor is a deliberately simple first-order speed plant:
//
//	dv/dt = (1/tau) * (K*V - v) - d(t)
//
// where:
//
//	v  = velocity (RPM)
//	V  = applied voltage (clamped)
//	K  = steady-state gain (RPM/V)
//	tau = mechanical time constant (s)
//	d(t) = external load disturbance (RPM/s)
//
// This model is good for building a control/experiment harness and for
// parameter sweeps. It is not a full electromechanical motor model.
// (Roadmap items like deadzone, Coulomb friction, load torque, encoder
// quantization can be added on top.)
type DCMotor struct {
	VelocityRPM float64

	GainRPMPerVolt float64
	TauSeconds     float64
	MaxVoltage     float64

	appliedVoltage     float64
	disturbanceRPMPerS float64
}

func NewDCMotor() *DCMotor {
	return &DCMotor{
		GainRPMPerVolt: 100.0,
		TauSeconds:     0.5,  // equivalent to Friction=2.0 in the old code
		MaxVoltage:     24.0, // headroom
	}
}

func (m *DCMotor) Observe() float64 {
	return m.VelocityRPM
}

func (m *DCMotor) Actuate(u float64) {
	m.appliedVoltage = clamp(u, -m.MaxVoltage, m.MaxVoltage)
}

// SetDisturbanceRPMPerS implements system.DisturbanceReceiver.
// Sets the external load disturbance in RPM/s deceleration.
func (m *DCMotor) SetDisturbanceRPMPerS(d float64) {
	m.disturbanceRPMPerS = d
}

// CurrentDisturbanceRPMPerS implements system.DisturbanceReporter.
// Returns the currently set disturbance value in RPM/s.
func (m *DCMotor) CurrentDisturbanceRPMPerS() float64 {
	return m.disturbanceRPMPerS
}

func (m *DCMotor) Step(dt float64) {
	if dt <= 0 {
		return
	}

	// first-order approach to target speed
	target := m.GainRPMPerVolt * m.appliedVoltage
	alpha := dt / m.TauSeconds
	// Apply disturbance: dv = alpha*(target - v) - d*dt
	m.VelocityRPM += alpha*(target-m.VelocityRPM) - m.disturbanceRPMPerS*dt
}

var (
	_ system.DisturbanceReceiver = (*DCMotor)(nil)
	_ system.DisturbanceReporter = (*DCMotor)(nil)
)

func clamp(x, lo, hi float64) float64 {
	return math.Min(math.Max(x, lo), hi)
}
