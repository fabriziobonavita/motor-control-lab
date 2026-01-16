package artifacts

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"time"
)

// RunDir represents a single experiment run directory under runs/.
// It owns the naming convention and provides helpers to write common artifacts.
type RunDir struct {
	Dir string

	out *os.File
}

// Metadata is written to metadata.json to make runs self-describing.
// Params are experiment parameters (gains, dt, duration, target, etc.).
type Metadata struct {
	RunID        string            `json:"run_id"`
	CreatedAtUTC string            `json:"created_at_utc"`
	Kind         string            `json:"kind"` // e.g. "sim" or "hw"
	Plant        string            `json:"plant"`
	Experiment   string            `json:"experiment"`
	Params       map[string]any    `json:"params"`
	Environment  map[string]string `json:"environment"`
}

const (
	// timestampFormat is used for run directory names and timestamps.
	// Uses dashes instead of colons for filesystem compatibility.
	timestampFormat = "2006-01-02T15-04-05Z"
)

func Create(baseDir, kind, plant, experiment string, params map[string]any) (RunDir, Metadata, error) {
	ts := time.Now().UTC().Format(timestampFormat)
	runID := fmt.Sprintf("%s_%s_%s_%s", ts, kind, plant, experiment)
	dir := filepath.Join(baseDir, runID)

	if err := os.MkdirAll(dir, 0o755); err != nil {
		return RunDir{}, Metadata{}, err
	}

	md := Metadata{
		RunID:        runID,
		CreatedAtUTC: ts,
		Kind:         kind,
		Plant:        plant,
		Experiment:   experiment,
		Params:       params,
		Environment: map[string]string{
			"go_version": runtime.Version(),
			"os":         runtime.GOOS,
			"arch":       runtime.GOARCH,
		},
	}

	if err := WriteJSON(filepath.Join(dir, "metadata.json"), md); err != nil {
		return RunDir{}, Metadata{}, err
	}

	outFile, err := os.Create(filepath.Join(dir, "out.log"))
	if err != nil {
		return RunDir{}, Metadata{}, err
	}

	return RunDir{Dir: dir, out: outFile}, md, nil
}

func (r *RunDir) Out() *os.File { return r.out }

func (r *RunDir) Close() error {
	if r.out != nil {
		return r.out.Close()
	}
	return nil
}
