package ui

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"
)

func ExportCSV(rows []ResultRow, path string) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	w := csv.NewWriter(f)
	defer w.Flush()

	_ = w.Write([]string{"#", "status", "length", "time_ms", "payloads", "error"})
	for _, r := range rows {
		status := fmt.Sprintf("%d", r.StatusCode)
		if r.Error != "" {
			status = "ERR"
		}
		_ = w.Write([]string{
			fmt.Sprintf("%d", r.N),
			status,
			fmt.Sprintf("%d", r.Length),
			fmt.Sprintf("%d", r.Duration.Milliseconds()),
			r.Payloads,
			r.Error,
		})
	}
	return nil
}

type jsonRow struct {
	N          int    `json:"n"`
	StatusCode int    `json:"status"`
	Length     int    `json:"length"`
	DurationMs int64  `json:"time_ms"`
	Payloads   string `json:"payloads"`
	Error      string `json:"error,omitempty"`
}

func ExportJSON(rows []ResultRow, path string) error {
	out := make([]jsonRow, len(rows))
	for i, r := range rows {
		out[i] = jsonRow{
			N:          r.N,
			StatusCode: r.StatusCode,
			Length:     r.Length,
			DurationMs: r.Duration.Milliseconds(),
			Payloads:   r.Payloads,
			Error:      r.Error,
		}
	}
	data, err := json.MarshalIndent(out, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

func ExportText(rows []ResultRow, path string) error {
	var sb strings.Builder
	fmt.Fprintf(&sb, "%-6s %-8s %-8s %-12s %s\n", "#", "status", "length", "time", "payload")
	sb.WriteString(strings.Repeat("─", 60) + "\n")
	for _, r := range rows {
		status := fmt.Sprintf("%d", r.StatusCode)
		if r.Error != "" {
			status = "ERR"
		}
		fmt.Fprintf(&sb, "%-6d %-8s %-8d %-12s %s\n",
			r.N, status, r.Length,
			r.Duration.Round(time.Millisecond),
			r.Payloads)
	}
	return os.WriteFile(path, []byte(sb.String()), 0o644)
}
