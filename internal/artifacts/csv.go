package artifacts

import (
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"

	"github.com/fabriziobonavita/motor-control-lab/internal/experiment"
)

// WriteSamplesCSV writes the time series to samples.csv inside the run directory.
func (r *RunDir) WriteSamplesCSV(samples []experiment.Sample) error {
	f, err := os.Create(filepath.Join(r.Dir, "samples.csv"))
	if err != nil {
		return err
	}
	defer f.Close()

	w := csv.NewWriter(f)
	defer w.Flush()

	header := []string{"t", "dt", "target", "actual", "error", "u", "p", "i", "d", "out_raw", "saturated", "integrated"}
	if err := w.Write(header); err != nil {
		return err
	}

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
		if err := w.Write(rec); err != nil {
			return err
		}
	}

	return w.Error()
}
