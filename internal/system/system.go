package system

// System is the minimal closed-loop interface for simulation experiments.
//
// Observe returns the current measurement (e.g., velocity in RPM).
// Actuate applies the controller command (e.g., volts).
// Step advances the system by dt seconds.
//
// For real hardware, you'll typically use a realtime runner that measures dt from
// wall-clock time and calls Observe/Actuate at a fixed cadence.
type System interface {
	Observe() float64
	Actuate(u float64)
	Step(dt float64)
}
