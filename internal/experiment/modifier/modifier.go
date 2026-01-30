package modifier

import "math"

type Modifier interface {
	Modify(u float64) float64
}

type DeadzoneModifier struct {
	Threshold float64
}

func (m *DeadzoneModifier) Modify(u float64) float64 {
	absU := math.Abs(u)
	if absU < m.Threshold {
		return 0
	}
	if u > 0 {
		return absU - m.Threshold
	}
	return -(absU - m.Threshold)
}

type chain struct {
	modifiers []Modifier
}

func (c *chain) Modify(u float64) float64 {
	for _, mod := range c.modifiers {
		u = mod.Modify(u)
	}
	return u
}

func Chain(mods ...Modifier) Modifier {
	return &chain{modifiers: mods}
}
