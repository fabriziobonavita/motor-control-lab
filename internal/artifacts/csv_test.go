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
			Signals:    map[string]float64{"disturbance_rpm_per_s": 0.0},
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
			Signals:    map[string]float64{"disturbance_rpm_per_s": 5.0},
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
	defer func() {
		_ = f.Close()
	}()

	reader := csv.NewReader(f)
	records, err := reader.ReadAll()
	if err != nil {
		t.Fatalf("failed to read CSV: %v", err)
	}

	// Check header
	if len(records) == 0 {
		t.Fatal("CSV has no records")
	}

	header := records[0]
	// Base fields should be present
	baseFields := []string{"t", "dt", "target", "actual", "error", "u", "p", "i", "d", "out_raw", "saturated", "integrated"}
	if len(header) < len(baseFields) {
		t.Errorf("header length = %d, want at least %d", len(header), len(baseFields))
	}
	for i, field := range baseFields {
		if i < len(header) && header[i] != field {
			t.Errorf("header[%d] = %q, want %q", i, header[i], field)
		}
	}
	// Signal fields should be present after base fields
	if len(header) <= len(baseFields) {
		t.Error("header should include signal columns")
	}
	// Verify disturbance_rpm_per_s is in the header
	hasDisturbance := false
	for _, h := range header {
		if h == "disturbance_rpm_per_s" {
			hasDisturbance = true
			break
		}
	}
	if !hasDisturbance {
		t.Error("header should include disturbance_rpm_per_s signal column")
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

		// Find disturbance_rpm_per_s column index
		disturbanceIdx := -1
		for i, h := range header {
			if h == "disturbance_rpm_per_s" {
				disturbanceIdx = i
				break
			}
		}
		if disturbanceIdx == -1 {
			t.Error("disturbance_rpm_per_s column not found in header")
		} else if disturbanceIdx < len(row) {
			disturbanceVal, err := strconv.ParseFloat(row[disturbanceIdx], 64)
			if err != nil {
				t.Errorf("failed to parse disturbance_rpm_per_s: %v", err)
			} else {
				expectedDist := 0.0
				if samples[0].Signals != nil {
					expectedDist = samples[0].Signals["disturbance_rpm_per_s"]
				}
				if disturbanceVal != expectedDist {
					t.Errorf("disturbance_rpm_per_s = %v, want %v", disturbanceVal, expectedDist)
				}
			}
		}
	}
}

func TestWriteSamplesCSV_WithSignals(t *testing.T) {
	dir := t.TempDir()
	runDir := RunDir{Dir: dir}

	// Create sample data with signals
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
			Signals:    map[string]float64{"disturbance_rpm_per_s": 0.0},
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
			Signals:    map[string]float64{"disturbance_rpm_per_s": 50.0},
		},
		{
			T:          0.002,
			DT:         0.001,
			Target:     1000.0,
			Actual:     100.0,
			Error:      900.0,
			U:          20.0,
			P:          18.0,
			I:          7.0,
			D:          2.0,
			OutRaw:     27.0,
			Saturated:  false,
			Integrated: true,
			Signals:    nil, // No signals for this sample
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
	defer func() {
		_ = f.Close()
	}()

	reader := csv.NewReader(f)
	records, err := reader.ReadAll()
	if err != nil {
		t.Fatalf("failed to read CSV: %v", err)
	}

	if len(records) < 2 {
		t.Fatal("CSV has too few records")
	}

	header := records[0]

	// Verify header includes base fields and signal
	baseFields := []string{"t", "dt", "target", "actual", "error", "u", "p", "i", "d", "out_raw", "saturated", "integrated"}
	for i, field := range baseFields {
		if i >= len(header) || header[i] != field {
			t.Errorf("header[%d] = %q, want %q", i, header[i], field)
		}
	}

	// Verify signal column exists
	disturbanceIdx := -1
	for i, h := range header {
		if h == "disturbance_rpm_per_s" {
			disturbanceIdx = i
			break
		}
	}
	if disturbanceIdx == -1 {
		t.Error("disturbance_rpm_per_s column not found in header")
	}

	// Verify signal values are written correctly
	if len(records) > 1 {
		// First row: disturbance = 0.0
		row1 := records[1]
		if disturbanceIdx < len(row1) {
			val1, err := strconv.ParseFloat(row1[disturbanceIdx], 64)
			if err != nil {
				t.Errorf("failed to parse disturbance_rpm_per_s in row 1: %v", err)
			} else if val1 != 0.0 {
				t.Errorf("row 1 disturbance_rpm_per_s = %v, want 0.0", val1)
			}
		}

		// Second row: disturbance = 50.0
		if len(records) > 2 {
			row2 := records[2]
			if disturbanceIdx < len(row2) {
				val2, err := strconv.ParseFloat(row2[disturbanceIdx], 64)
				if err != nil {
					t.Errorf("failed to parse disturbance_rpm_per_s in row 2: %v", err)
				} else if val2 != 50.0 {
					t.Errorf("row 2 disturbance_rpm_per_s = %v, want 50.0", val2)
				}
			}
		}

		// Third row: no signals, should be 0.0
		if len(records) > 3 {
			row3 := records[3]
			if disturbanceIdx < len(row3) {
				val3, err := strconv.ParseFloat(row3[disturbanceIdx], 64)
				if err != nil {
					t.Errorf("failed to parse disturbance_rpm_per_s in row 3: %v", err)
				} else if val3 != 0.0 {
					t.Errorf("row 3 disturbance_rpm_per_s = %v, want 0.0 (no signals)", val3)
				}
			}
		}
	}
}
