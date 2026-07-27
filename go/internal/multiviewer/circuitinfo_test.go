package multiviewer

import (
	"os"
	"testing"
)

func TestParseCircuitInfo_RealCapture_Catalunya(t *testing.T) {
	data, err := os.ReadFile("testdata/catalunya_2024.json")
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}

	info, err := ParseCircuitInfo(data)
	if err != nil {
		t.Fatalf("ParseCircuitInfo: %v", err)
	}

	if len(info.Corners) != 14 {
		t.Errorf("got %d corners, want 14", len(info.Corners))
	}
	if len(info.MarshalLights) != 16 {
		t.Errorf("got %d marshal lights, want 16", len(info.MarshalLights))
	}
	if len(info.MarshalSectors) != 16 {
		t.Errorf("got %d marshal sectors, want 16", len(info.MarshalSectors))
	}
	if info.Rotation != 303 {
		t.Errorf("Rotation = %v, want 303", info.Rotation)
	}

	first := info.Corners[0]
	want := Corner{
		Number: 1, Letter: "",
		X: -3102.4705973153305, Y: -6966.730800740276,
		Angle: -68.8993899041057, Length: 8118.684730529785,
	}
	if first != want {
		t.Errorf("first corner = %+v, want %+v", first, want)
	}

	for _, c := range info.Corners {
		if c.Letter != "" {
			t.Errorf("unexpected lettered corner in Catalunya: %+v", c)
		}
	}
}

func TestParseCircuitInfo_RealCapture_HungaroringLetteredCorners(t *testing.T) {
	data, err := os.ReadFile("testdata/hungaroring_2024.json")
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}

	info, err := ParseCircuitInfo(data)
	if err != nil {
		t.Fatalf("ParseCircuitInfo: %v", err)
	}

	var lettered []Corner
	for _, c := range info.Corners {
		if c.Letter != "" {
			lettered = append(lettered, c)
		}
	}
	if len(lettered) != 2 {
		t.Fatalf("got %d lettered corners, want 2 (real capture: corners 1A and 12A)", len(lettered))
	}
	for _, c := range lettered {
		if c.Letter != "A" {
			t.Errorf("lettered corner %+v, want Letter = A", c)
		}
		if c.Number != 1 && c.Number != 12 {
			t.Errorf("lettered corner Number = %d, want 1 or 12", c.Number)
		}
	}
}

func TestParseCircuitInfo_RejectsMalformed(t *testing.T) {
	if _, err := ParseCircuitInfo([]byte("not json")); err == nil {
		t.Fatal("expected an error for a malformed payload")
	}
}
