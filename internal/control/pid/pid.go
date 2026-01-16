package pid

import "math"

// Trace captures the internal terms of the PID controller for logging and
// debugging. If you don't need tracing, pass nil to Controller.Step().
//
// The terms are expressed in the same units as the output (e.g., volts).
// OutRaw is the sum before clamping; Out is the clamped output.
type Trace struct {
	Target float64
	Actual float64
	Error  float64

	P float64
	I float64
	D float64

	OutRaw     float64
	Out        float64
	Saturated  bool
	Integrated bool // whether the integrator was updated this step
}

// Controller is a classic PID controller with output clamping and basic anti-windup.
//
// Anti-windup strategy: freeze the integrator when the predicted output is saturated
// in the same direction as the error.
//
// This preserves the behavior of the original implementation, but uses clearer
// names and an optional trace output.
type Controller struct {
	Kp, Ki, Kd float64

	OutMin float64
	OutMax float64

	integral  float64
	prevError float64
	hasPrev   bool
}

func New(kp, ki, kd float64) *Controller {
	return &Controller{
		Kp:     kp,
		Ki:     ki,
		Kd:     kd,
		OutMin: -24.0,
		OutMax: 24.0,
	}
}

// Step computes the control output for the given target and measurement.
//
// If tr != nil, it is populated with the term breakdown and clamping info.
func (c *Controller) Step(target, actual, dt float64, tr *Trace) float64 {
	err := target - actual
	if dt <= 0 {
		if tr != nil {
			*tr = Trace{Target: target, Actual: actual, Error: err}
		}
		return 0
	}

	pTerm := c.Kp * err

	dTerm := 0.0
	if c.hasPrev {
		dTerm = c.Kd * (err - c.prevError) / dt
	}

	// Predict saturation using the current integrator state.
	outNoI := pTerm + dTerm
	outPred := outNoI + c.Ki*c.integral

	satHigh := outPred >= c.OutMax
	satLow := outPred <= c.OutMin

	integrated := true
	if (satHigh && err > 0) || (satLow && err < 0) {
		// Would wind up further into saturation.
		integrated = false
	} else {
		c.integral += err * dt
	}

	iTerm := c.Ki * c.integral

	outRaw := pTerm + iTerm + dTerm
	out := clamp(outRaw, c.OutMin, c.OutMax)

	if tr != nil {
		*tr = Trace{
			Target:     target,
			Actual:     actual,
			Error:      err,
			P:          pTerm,
			I:          iTerm,
			D:          dTerm,
			OutRaw:     outRaw,
			Out:        out,
			Saturated:  out != outRaw,
			Integrated: integrated,
		}
	}

	c.prevError = err
	c.hasPrev = true
	return out
}

func clamp(x, lo, hi float64) float64 {
	return math.Min(math.Max(x, lo), hi)
}
