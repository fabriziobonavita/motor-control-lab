package experiment

import (
	"time"

	"github.com/fabriziobonavita/motor-control-lab/internal/control/pid"
	"github.com/fabriziobonavita/motor-control-lab/internal/experiment/modifier"
	"github.com/fabriziobonavita/motor-control-lab/internal/system"
)

// StepConfig defines a constant-setpoint step experiment.
type StepConfig struct {
	TargetRPM float64
	DT        float64
	Duration  float64
	Modifier  modifier.Modifier
}

// Sample is a single time step of recorded run data.
type Sample struct {
	T  float64
	DT float64

	Target float64
	Actual float64
	Error  float64

	U float64

	P float64
	I float64
	D float64

	OutRaw     float64
	Saturated  bool
	Integrated bool

	// Signals contains additional numeric signals exposed by the system (e.g., disturbance_rpm_per_s).
	// The map is nil or empty when no signals are available.
	// Keys are stable snake_case identifiers suitable for CSV headers.
	Signals map[string]float64
}

// RunStep executes the closed-loop experiment and returns the full time series.
// The returned wall time is useful for profiling (sim should be much faster than realtime).
//
// RunStep is a clean generic harness: Observe -> ctrl.Step -> Modifier -> Actuate -> Step -> record sample.
// It optionally queries system capabilities for logging purposes but does not apply or schedule any physics.
func RunStep(sys system.System, ctrl *pid.Controller, cfg StepConfig) ([]Sample, time.Duration) {
	start := time.Now()

	if cfg.DT <= 0 || cfg.Duration <= 0 {
		return nil, time.Since(start)
	}

	steps := int(cfg.Duration / cfg.DT)
	out := make([]Sample, 0, steps)

	// Optionally query system capabilities for logging (generic, no semantic knowledge)
	var signalReporter system.SignalReporter
	if sr, ok := sys.(system.SignalReporter); ok {
		signalReporter = sr
	}

	for i := 0; i < steps; i++ {
		t := float64(i) * cfg.DT

		actual := sys.Observe()
		var tr pid.Trace
		u := ctrl.Step(cfg.TargetRPM, actual, cfg.DT, &tr)

		if cfg.Modifier != nil {
			u = cfg.Modifier.Modify(u)
		}

		sys.Actuate(u)
		sys.Step(cfg.DT)

		// Query signals if system exposes them (for logging only)
		var sigs map[string]float64
		if signalReporter != nil {
			raw := signalReporter.Signals()
			if len(raw) > 0 {
				// Copy the map to avoid mutation affecting stored samples
				sigs = make(map[string]float64, len(raw))
				for k, v := range raw {
					sigs[k] = v
				}
			}
		}

		out = append(out, Sample{
			T:          t,
			DT:         cfg.DT,
			Target:     tr.Target,
			Actual:     tr.Actual,
			Error:      tr.Error,
			U:          u,
			P:          tr.P,
			I:          tr.I,
			D:          tr.D,
			OutRaw:     tr.OutRaw,
			Saturated:  tr.Saturated,
			Integrated: tr.Integrated,
			Signals:    sigs,
		})
	}

	return out, time.Since(start)
}
