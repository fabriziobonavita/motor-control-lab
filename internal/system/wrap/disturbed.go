package wrap

import (
	"github.com/fabriziobonavita/motor-control-lab/internal/system"
)

// StepDisturbanceConfig defines a step load disturbance injection configuration.
// This config owns all disturbance semantics: timing, shape, and magnitude.
type StepDisturbanceConfig struct {
	Enabled          bool
	StartS           float64
	DurationS        float64 // 0 means infinite
	MagnitudeRPMPerS float64
}

// DisturbedSystem wraps a system.System and applies time-varying load disturbances.
// It manages internal simulation time and applies disturbances based on a StepDisturbanceConfig.
//
// The wrapper implements system.System by delegating Observe() and Actuate() to the inner system.
// In Step(dt), it computes the current disturbance, applies it to the inner system if it implements
// system.DisturbanceReceiver, then steps the inner system and increments its internal time.
type DisturbedSystem struct {
	inner system.System
	cfg   StepDisturbanceConfig

	// Internal simulation time (seconds)
	t float64

	// Last applied disturbance value (for reporting)
	lastDisturbanceRPMPerS float64
}

// NewDisturbedSystem creates a new DisturbedSystem wrapper around the given inner system.
// The wrapper will apply disturbances according to cfg when Step() is called.
func NewDisturbedSystem(inner system.System, cfg StepDisturbanceConfig) *DisturbedSystem {
	return &DisturbedSystem{
		inner:                  inner,
		cfg:                    cfg,
		t:                      0.0,
		lastDisturbanceRPMPerS: 0.0,
	}
}

// Observe delegates to the inner system.
func (d *DisturbedSystem) Observe() float64 {
	return d.inner.Observe()
}

// Actuate delegates to the inner system.
func (d *DisturbedSystem) Actuate(u float64) {
	d.inner.Actuate(u)
}

// Step computes the current disturbance, applies it to the inner system if it supports
// disturbance injection, then steps the inner system and increments internal time.
func (d *DisturbedSystem) Step(dt float64) {
	// Compute disturbance at the end of this step interval (after the step)
	// This represents the disturbance active during the step
	dist := computeDisturbance(d.t+dt, d.cfg)
	d.lastDisturbanceRPMPerS = dist

	// Apply disturbance to inner system if it supports it
	if distReceiver, ok := d.inner.(system.DisturbanceReceiver); ok {
		distReceiver.SetDisturbanceRPMPerS(dist)
	}

	// Step the inner system
	d.inner.Step(dt)

	// Increment internal time after stepping
	d.t += dt
}

// Signals implements system.SignalReporter.
// Returns a map containing the current disturbance signal.
func (d *DisturbedSystem) Signals() map[string]float64 {
	return map[string]float64{
		"disturbance_rpm_per_s": d.lastDisturbanceRPMPerS,
	}
}

// CurrentDisturbanceRPMPerS returns the disturbance value that was applied in the last Step() call.
// Deprecated: Use Signals() instead for generic signal reporting.
func (d *DisturbedSystem) CurrentDisturbanceRPMPerS() float64 {
	return d.lastDisturbanceRPMPerS
}

// ResetTime resets the internal simulation time to zero.
// Useful for reusing the wrapper in multiple experiments.
func (d *DisturbedSystem) ResetTime() {
	d.t = 0.0
	d.lastDisturbanceRPMPerS = 0.0
}

// computeDisturbance returns the disturbance magnitude at time t based on cfg.
// Returns 0 if disturbance is disabled, before StartS, or at/after StartS+DurationS (if DurationS > 0).
func computeDisturbance(t float64, cfg StepDisturbanceConfig) float64 {
	if !cfg.Enabled || cfg.MagnitudeRPMPerS == 0 {
		return 0
	}
	if t < cfg.StartS {
		return 0
	}
	if cfg.DurationS > 0 && t >= cfg.StartS+cfg.DurationS {
		return 0
	}
	return cfg.MagnitudeRPMPerS
}

var _ system.SignalReporter = (*DisturbedSystem)(nil)
