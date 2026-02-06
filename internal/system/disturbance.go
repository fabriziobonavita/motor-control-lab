package system

// DisturbanceReceiver is an optional extension for systems whose dynamics
// can accept an external disturbance modeled as RPM/s deceleration.
//
// The disturbance d(t) is applied in the plant dynamics:
//
//	dv/dt = (1/tau) * (K*V - v) - d(t)
//
// where:
//
//	v = velocity (RPM)
//	V = applied voltage (clamped)
//	K = steady-state gain (RPM/V)
//	tau = mechanical time constant (s)
//	d(t) = external load disturbance (RPM/s)
//
// Systems that implement this interface can accept time-varying disturbances
// for testing disturbance rejection in control loops.
type DisturbanceReceiver interface {
	// SetDisturbanceRPMPerS sets the current disturbance value in RPM/s.
	// The disturbance is applied during the next Step() call.
	SetDisturbanceRPMPerS(d float64)
}

// DisturbanceReporter is an optional interface for systems that can report
// the current or last-applied disturbance value. This is useful for logging
// and analysis without requiring the experiment harness to know about disturbance
// scheduling logic.
type DisturbanceReporter interface {
	// CurrentDisturbanceRPMPerS returns the disturbance value that was applied
	// in the last Step() call, or 0 if no disturbance was applied.
	CurrentDisturbanceRPMPerS() float64
}
