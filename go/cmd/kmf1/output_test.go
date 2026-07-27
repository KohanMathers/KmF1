package main

import (
	"io"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"
)

func TestDurOrDash(t *testing.T) {
	if got := durOrDash(nil); got != naCell {
		t.Errorf("durOrDash(nil) = %q, want %q", got, naCell)
	}
	d := 90 * time.Second
	if got := durOrDash(&d); got != d.String() {
		t.Errorf("durOrDash(&90s) = %q, want %q", got, d.String())
	}
}

func TestIntOrDash(t *testing.T) {
	if got := intOrDash(nil); got != naCell {
		t.Errorf("intOrDash(nil) = %q, want %q", got, naCell)
	}
	n := 3
	if got := intOrDash(&n); got != "3" {
		t.Errorf("intOrDash(&3) = %q, want %q", got, "3")
	}
}

func TestFloatOrDash(t *testing.T) {
	if got := floatOrDash(nil); got != naCell {
		t.Errorf("floatOrDash(nil) = %q, want %q", got, naCell)
	}
	f := 123.4
	if got := floatOrDash(&f); got != strconv.FormatFloat(f, 'f', 3, 64) {
		t.Errorf("floatOrDash(&123.4) = %q", got)
	}
}

func TestStrOrDash(t *testing.T) {
	if got := strOrDash(""); got != naCell {
		t.Errorf("strOrDash(\"\") = %q, want %q", got, naCell)
	}
	if got := strOrDash("VER"); got != "VER" {
		t.Errorf("strOrDash(\"VER\") = %q, want %q", got, "VER")
	}
}

func TestBoolYesNo(t *testing.T) {
	if got := boolYesNo(true); got != "yes" {
		t.Errorf("boolYesNo(true) = %q, want yes", got)
	}
	if got := boolYesNo(false); got != "no" {
		t.Errorf("boolYesNo(false) = %q, want no", got)
	}
}

func TestOutput_JSONMode(t *testing.T) {
	type row struct{ Name string }
	captured := captureStdout(t, func() {
		if err := output(true, []row{{Name: "Verstappen"}}, nil, nil); err != nil {
			t.Fatalf("output: %v", err)
		}
	})
	if !strings.Contains(captured, `"Name": "Verstappen"`) {
		t.Errorf("JSON output missing expected field, got:\n%s", captured)
	}
}

func TestOutput_TableMode(t *testing.T) {
	captured := captureStdout(t, func() {
		headers := []string{"NUM", "NAME"}
		rows := [][]string{{"1", "Verstappen"}, {"44", "Hamilton"}}
		if err := output(false, nil, headers, rows); err != nil {
			t.Fatalf("output: %v", err)
		}
	})
	lines := strings.Split(strings.TrimRight(captured, "\n"), "\n")
	if len(lines) != 3 {
		t.Fatalf("table output = %d lines, want 3 (header + 2 rows), got:\n%s", len(lines), captured)
	}
	if !strings.Contains(lines[0], "NUM") || !strings.Contains(lines[0], "NAME") {
		t.Errorf("header line = %q, want it to contain NUM and NAME", lines[0])
	}
	if !strings.Contains(lines[1], "Verstappen") {
		t.Errorf("row line = %q, want it to contain Verstappen", lines[1])
	}
}

func captureStdout(t *testing.T, fn func()) string {
	t.Helper()
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe: %v", err)
	}
	orig := os.Stdout
	os.Stdout = w
	defer func() { os.Stdout = orig }()

	fn()
	_ = w.Close()

	data, err := io.ReadAll(r)
	if err != nil {
		t.Fatalf("read captured stdout: %v", err)
	}
	return string(data)
}
