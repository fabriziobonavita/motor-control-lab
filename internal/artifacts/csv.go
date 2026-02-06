package artifacts

import (
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/fabriziobonavita/motor-control-lab/internal/experiment"
)

// WriteSamplesCSV writes the time series to samples.csv inside the run directory.
// Signal columns are included in deterministic lexicographic order.
func (r *RunDir) WriteSamplesCSV(samples []experiment.Sample) error {
	f, err := os.Create(filepath.Join(r.Dir, "samples.csv"))
	if err != nil {
		return err
	}
	defer func() {
		_ = f.Close() // Error on close is non-fatal for CSV writing - file is already written
	}()

	w := csv.NewWriter(f)
	defer w.Flush()

	// Gather all signal keys from all samples
	signalKeysSet := make(map[string]bool)
	for _, s := range samples {
		if s.Signals != nil {
			for k := range s.Signals {
				signalKeysSet[k] = true
			}
		}
	}

	// Convert to sorted slice for deterministic ordering
	var signalKeys []string
	if len(signalKeysSet) > 0 {
		signalKeys = make([]string, 0, len(signalKeysSet))
		for k := range signalKeysSet {
			signalKeys = append(signalKeys, k)
		}
		sort.Strings(signalKeys)
	}

	// Build header: base fields first, then signal keys
	header := []string{"t", "dt", "target", "actual", "error", "u", "p", "i", "d", "out_raw", "saturated", "integrated"}
	header = append(header, signalKeys...)
	if err := w.Write(header); err != nil {
		return err
	}

	// Write data rows
	for _, s := range samples {
		rec := []string{
			fmt.Sprintf("%.6f", s.T),
			fmt.Sprintf("%.6f", s.DT),
			fmt.Sprintf("%.6f", s.Target),
			fmt.Sprintf("%.6f", s.Actual),
			fmt.Sprintf("%.6f", s.Error),
			fmt.Sprintf("%.6f", s.U),
			fmt.Sprintf("%.6f", s.P),
			fmt.Sprintf("%.6f", s.I),
			fmt.Sprintf("%.6f", s.D),
			fmt.Sprintf("%.6f", s.OutRaw),
			fmt.Sprintf("%t", s.Saturated),
			fmt.Sprintf("%t", s.Integrated),
		}

		// Append signal values in sorted key order
		for _, key := range signalKeys {
			val := 0.0
			if s.Signals != nil {
				if v, ok := s.Signals[key]; ok {
					val = v
				}
			}
			rec = append(rec, fmt.Sprintf("%.6f", val))
		}

		if err := w.Write(rec); err != nil {
			return err
		}
	}

	return w.Error()
}
