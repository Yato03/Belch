package ui_test

import (
	"encoding/csv"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"fuzzer/ui"
)

var sampleRows = []ui.ResultRow{
	{N: 1, StatusCode: 200, Length: 1234, Duration: 45 * time.Millisecond, Payloads: "admin"},
	{N: 2, StatusCode: 302, Length: 0, Duration: 12 * time.Millisecond, Payloads: "root | password"},
	{N: 3, StatusCode: 404, Length: 500, Duration: 30 * time.Millisecond, Payloads: "guest", Error: "not found"},
}

func TestExportCSV_CreatesFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "out.csv")
	if err := ui.ExportCSV(sampleRows, path); err != nil {
		t.Fatalf("ExportCSV: %v", err)
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("file not created: %v", err)
	}
}

func TestExportCSV_RowCount(t *testing.T) {
	path := filepath.Join(t.TempDir(), "out.csv")
	_ = ui.ExportCSV(sampleRows, path)

	f, _ := os.Open(path)
	defer f.Close()
	records, _ := csv.NewReader(f).ReadAll()
	// header + 3 data rows
	if len(records) != 4 {
		t.Errorf("CSV rows: got %d, want 4", len(records))
	}
}

func TestExportCSV_Header(t *testing.T) {
	path := filepath.Join(t.TempDir(), "out.csv")
	_ = ui.ExportCSV(sampleRows, path)

	f, _ := os.Open(path)
	defer f.Close()
	records, _ := csv.NewReader(f).ReadAll()
	header := records[0]
	if header[0] != "#" || header[1] != "status" || header[4] != "payloads" {
		t.Errorf("unexpected CSV header: %v", header)
	}
}

func TestExportCSV_ErrorRow(t *testing.T) {
	path := filepath.Join(t.TempDir(), "out.csv")
	_ = ui.ExportCSV(sampleRows, path)

	f, _ := os.Open(path)
	defer f.Close()
	records, _ := csv.NewReader(f).ReadAll()
	// row 3 (index 3 in records) has Error set — status should be ERR
	if records[3][1] != "ERR" {
		t.Errorf("error row status: got %q, want ERR", records[3][1])
	}
}

func TestExportJSON_Valid(t *testing.T) {
	path := filepath.Join(t.TempDir(), "out.json")
	if err := ui.ExportJSON(sampleRows, path); err != nil {
		t.Fatalf("ExportJSON: %v", err)
	}
	data, _ := os.ReadFile(path)
	var out []map[string]any
	if err := json.Unmarshal(data, &out); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if len(out) != 3 {
		t.Errorf("JSON rows: got %d, want 3", len(out))
	}
}

func TestExportJSON_Fields(t *testing.T) {
	path := filepath.Join(t.TempDir(), "out.json")
	_ = ui.ExportJSON(sampleRows, path)
	data, _ := os.ReadFile(path)
	var out []map[string]any
	_ = json.Unmarshal(data, &out)

	row := out[0]
	for _, field := range []string{"n", "status", "length", "time_ms", "payloads"} {
		if _, ok := row[field]; !ok {
			t.Errorf("JSON row missing field %q", field)
		}
	}
}

func TestExportText_CreatesFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "out.txt")
	if err := ui.ExportText(sampleRows, path); err != nil {
		t.Fatalf("ExportText: %v", err)
	}
	data, _ := os.ReadFile(path)
	if len(data) == 0 {
		t.Error("text file is empty")
	}
}

func TestExportText_ContainsPayload(t *testing.T) {
	path := filepath.Join(t.TempDir(), "out.txt")
	_ = ui.ExportText(sampleRows, path)
	data, _ := os.ReadFile(path)
	if !strings.Contains(string(data), "admin") {
		t.Error("text export missing payload 'admin'")
	}
}

func TestExportText_ContainsHeader(t *testing.T) {
	path := filepath.Join(t.TempDir(), "out.txt")
	_ = ui.ExportText(sampleRows, path)
	data, _ := os.ReadFile(path)
	if !strings.Contains(string(data), "status") {
		t.Error("text export missing header")
	}
}
