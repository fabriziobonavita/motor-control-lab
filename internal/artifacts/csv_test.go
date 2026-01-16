package artifacts

import (
	"encoding/csv"
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/fabriziobonavita/motor-control-lab/internal/experiment"
)

func TestWriteSamplesCSV(t *testing.T) {
	dir := t.TempDir()
	runDir := RunDir{Dir: dir}

	// Create sample data
	samples := []experiment.Sample{
		{
			T:          0.0,
			DT:         0.001,
			Target:     1000.0,
			Actual:     0.0,
			Error:      1000.0,
			U:          10.0,
			P:          20.0,
			I:          5.0,
			D:          0.0,
			OutRaw:     25.0,
			Saturated:  false,
			Integrated: true,
		},
		{
			T:          0.001,
			DT:         0.001,
			Target:     1000.0,
			Actual:     50.0,
			Error:      950.0,
			U:          15.0,
			P:          19.0,
			I:          6.0,
			D:          1.0,
			OutRaw:     26.0,
			Saturated:  false,
			Integrated: true,
		},
	}

	// Write CSV
	if err := runDir.WriteSamplesCSV(samples); err != nil {
		t.Fatalf("WriteSamplesCSV() error = %v", err)
	}

	// Read back and verify
	csvPath := filepath.Join(dir, "samples.csv")
	f, err := os.Open(csvPath)
	if err != nil {
		t.Fatalf("failed to open CSV: %v", err)
	}
	defer f.Close()

	reader := csv.NewReader(f)
	records, err := reader.ReadAll()
	if err != nil {
		t.Fatalf("failed to read CSV: %v", err)
	}

	// Check header
	expectedHeader := []string{"t", "dt", "target", "actual", "error", "u", "p", "i", "d", "out_raw", "saturated", "integrated"}
	if len(records) == 0 {
		t.Fatal("CSV has no records")
	}

	header := records[0]
	if len(header) != len(expectedHeader) {
		t.Errorf("header length = %d, want %d", len(header), len(expectedHeader))
	}
	for i, h := range expectedHeader {
		if i < len(header) && header[i] != h {
			t.Errorf("header[%d] = %q, want %q", i, header[i], h)
		}
	}

	// Check number of data rows
	expectedRows := len(samples)
	actualRows := len(records) - 1 // minus header
	if actualRows != expectedRows {
		t.Errorf("number of data rows = %d, want %d", actualRows, expectedRows)
	}

	// Verify first data row matches first sample
	if len(records) > 1 {
		row := records[1]
		if len(row) < 4 {
			t.Fatal("first data row has too few columns")
		}

		// Parse and check a few key fields
		tVal, err := strconv.ParseFloat(row[0], 64)
		if err != nil {
			t.Errorf("failed to parse t: %v", err)
		} else if tVal != samples[0].T {
			t.Errorf("t = %v, want %v", tVal, samples[0].T)
		}

		actualVal, err := strconv.ParseFloat(row[3], 64)
		if err != nil {
			t.Errorf("failed to parse actual: %v", err)
		} else if actualVal != samples[0].Actual {
			t.Errorf("actual = %v, want %v", actualVal, samples[0].Actual)
		}

		saturatedVal, err := strconv.ParseBool(row[10])
		if err != nil {
			t.Errorf("failed to parse saturated: %v", err)
		} else if saturatedVal != samples[0].Saturated {
			t.Errorf("saturated = %v, want %v", saturatedVal, samples[0].Saturated)
		}
	}
}
