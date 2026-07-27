package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"text/tabwriter"
	"time"
)

func output(jsonOut bool, v any, headers []string, rows [][]string) error {
	if jsonOut {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(v)
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 2, 2, ' ', 0)
	fmt.Fprintln(w, strings.Join(headers, "\t"))
	for _, row := range rows {
		fmt.Fprintln(w, strings.Join(row, "\t"))
	}
	return w.Flush()
}

const naCell = "-"

func durOrDash(d *time.Duration) string {
	if d == nil {
		return naCell
	}
	return d.String()
}

func intOrDash(n *int) string {
	if n == nil {
		return naCell
	}
	return strconv.Itoa(*n)
}

func floatOrDash(f *float64) string {
	if f == nil {
		return naCell
	}
	return strconv.FormatFloat(*f, 'f', 3, 64)
}

func strOrDash(s string) string {
	if s == "" {
		return naCell
	}
	return s
}

func boolYesNo(b bool) string {
	if b {
		return "yes"
	}
	return "no"
}
