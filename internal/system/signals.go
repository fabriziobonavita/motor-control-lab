package system

// SignalReporter is an optional capability for systems that can expose
// additional numeric signals for logging/analysis.
// Keys must be stable snake_case identifiers suitable for CSV headers.
type SignalReporter interface {
	// Signals returns a map of signal names to their current values.
	// The map should be non-nil and may be empty if no signals are available.
	// Keys should be stable snake_case identifiers suitable for CSV column headers.
	// The returned map may be modified by the caller without affecting the system.
	Signals() map[string]float64
}
