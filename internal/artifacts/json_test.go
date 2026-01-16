package artifacts

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestWriteJSON(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.json")

	// Test data structure
	type TestData struct {
		Name  string  `json:"name"`
		Value float64 `json:"value"`
		Count int     `json:"count"`
	}

	data := TestData{
		Name:  "test",
		Value: 123.456,
		Count: 42,
	}

	// Write JSON
	if err := WriteJSON(path, data); err != nil {
		t.Fatalf("WriteJSON() error = %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Fatal("JSON file was not created")
	}

	// Read back and verify
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read JSON file: %v", err)
	}

	// Verify it's valid JSON
	var decoded TestData
	if err := json.Unmarshal(content, &decoded); err != nil {
		t.Fatalf("failed to parse JSON: %v", err)
	}

	// Verify content matches
	if decoded.Name != data.Name {
		t.Errorf("decoded.Name = %q, want %q", decoded.Name, data.Name)
	}
	if decoded.Value != data.Value {
		t.Errorf("decoded.Value = %v, want %v", decoded.Value, data.Value)
	}
	if decoded.Count != data.Count {
		t.Errorf("decoded.Count = %v, want %v", decoded.Count, data.Count)
	}

	// Verify it's pretty-printed (contains newlines/indentation)
	if len(content) < 50 {
		t.Error("JSON appears to not be pretty-printed (too short)")
	}
	// Pretty-printed JSON should have newlines
	hasNewline := false
	for _, b := range content {
		if b == '\n' {
			hasNewline = true
			break
		}
	}
	if !hasNewline {
		t.Error("JSON does not appear to be pretty-printed (no newlines)")
	}
}

func TestWriteJSON_WithMetrics(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "metrics.json")

	// Use a structure similar to Metrics
	type Metrics struct {
		Target              float64 `json:"target"`
		OvershootPercent    float64 `json:"overshoot_percent"`
		IAE                 float64 `json:"iae"`
		SettlingTimeSeconds float64 `json:"settling_time_seconds"`
	}

	metrics := Metrics{
		Target:              1000.0,
		OvershootPercent:    5.2,
		IAE:                 123.45,
		SettlingTimeSeconds: 2.5,
	}

	if err := WriteJSON(path, metrics); err != nil {
		t.Fatalf("WriteJSON() error = %v", err)
	}

	// Verify it parses
	var decoded Metrics
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read file: %v", err)
	}

	if err := json.Unmarshal(content, &decoded); err != nil {
		t.Fatalf("failed to parse JSON: %v", err)
	}

	if decoded.Target != metrics.Target {
		t.Errorf("Target = %v, want %v", decoded.Target, metrics.Target)
	}
}
