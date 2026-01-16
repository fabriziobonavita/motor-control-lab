package experiment

import (
	"time"

	"github.com/fabriziobonavita/motor-control-lab/internal/control/pid"
	"github.com/fabriziobonavita/motor-control-lab/internal/system"
)

// StepConfig defines a constant-setpoint step experiment.
type StepConfig struct {
	TargetRPM float64
	DT        float64
	Duration  float64
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
}

// RunStep executes the closed-loop experiment and returns the full time series.
// The returned wall time is useful for profiling (sim should be much faster than realtime).
func RunStep(sys system.System, ctrl *pid.Controller, cfg StepConfig) ([]Sample, time.Duration) {
	start := time.Now()

	if cfg.DT <= 0 || cfg.Duration <= 0 {
		return nil, time.Since(start)
	}

	steps := int(cfg.Duration / cfg.DT)
	out := make([]Sample, 0, steps)

	for i := 0; i < steps; i++ {
		t := float64(i) * cfg.DT

		actual := sys.Observe()
		var tr pid.Trace
		u := ctrl.Step(cfg.TargetRPM, actual, cfg.DT, &tr)

		sys.Actuate(u)
		sys.Step(cfg.DT)

		out = append(out, Sample{
			T:          t,
			DT:         cfg.DT,
			Target:     tr.Target,
			Actual:     tr.Actual,
			Error:      tr.Error,
			U:          tr.Out,
			P:          tr.P,
			I:          tr.I,
			D:          tr.D,
			OutRaw:     tr.OutRaw,
			Saturated:  tr.Saturated,
			Integrated: tr.Integrated,
		})
	}

	return out, time.Since(start)
}
