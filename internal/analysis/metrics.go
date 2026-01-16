package analysis

import (
	"math"

	"github.com/fabriziobonavita/motor-control-lab/internal/experiment"
)

// Metrics summarizes a run in engineering-friendly terms.
type Metrics struct {
	Target float64 `json:"target"`

	MaxActual float64 `json:"max_actual"`
	MinActual float64 `json:"min_actual"`

	OvershootPercent    float64 `json:"overshoot_percent"`
	SteadyStateError    float64 `json:"steady_state_error"`
	IAE                 float64 `json:"iae"`
	SettlingTimeSeconds float64 `json:"settling_time_seconds"`
	SaturationFraction  float64 `json:"saturation_fraction"`
}

// Compute calculates common step-response metrics.
// settleBandFrac is typically 0.02 for a 2% band.
func Compute(samples []experiment.Sample, settleBandFrac float64) Metrics {
	if len(samples) == 0 {
		return Metrics{SettlingTimeSeconds: math.NaN()}
	}

	target := samples[len(samples)-1].Target

	maxA := samples[0].Actual
	minA := samples[0].Actual

	var iae float64
	var sat int
	for _, s := range samples {
		if s.Actual > maxA {
			maxA = s.Actual
		}
		if s.Actual < minA {
			minA = s.Actual
		}
		iae += math.Abs(s.Error) * s.DT
		if s.Saturated {
			sat++
		}
	}

	overshoot := 0.0
	if target != 0 {
		o := (maxA - target) / math.Abs(target) * 100.0
		if o > 0 {
			overshoot = o
		}
	}

	steadyErr := samples[len(samples)-1].Error

	band := math.Abs(target) * settleBandFrac
	if band == 0 {
		band = 1e-9
	}

	settle := math.NaN()
	for i := range samples {
		if math.Abs(samples[i].Error) > band {
			continue
		}
		ok := true
		for j := i; j < len(samples); j++ {
			if math.Abs(samples[j].Error) > band {
				ok = false
				break
			}
		}
		if ok {
			settle = samples[i].T
			break
		}
	}

	return Metrics{
		Target:              target,
		MaxActual:           maxA,
		MinActual:           minA,
		OvershootPercent:    overshoot,
		SteadyStateError:    steadyErr,
		IAE:                 iae,
		SettlingTimeSeconds: settle,
		SaturationFraction:  float64(sat) / float64(len(samples)),
	}
}
