package feeds

import (
	"os"
	"testing"
)

func TestParseLapCount_RealCapture(t *testing.T) {
	data, err := os.ReadFile("testdata/lapcount.jsonstream")
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}

	progress, err := ParseLapCount(data)
	if err != nil {
		t.Fatalf("ParseLapCount: %v", err)
	}

	if len(progress) != 66 {
		t.Fatalf("got %d progress entries, want 66", len(progress))
	}

	first := progress[0]
	if first.CurrentLap != 1 {
		t.Errorf("first: CurrentLap = %d, want 1", first.CurrentLap)
	}
	if first.TotalLaps != 66 {
		t.Errorf("first: TotalLaps = %d, want 66", first.TotalLaps)
	}

	last := progress[len(progress)-1]
	if last.CurrentLap != 66 {
		t.Errorf("last: CurrentLap = %d, want 66", last.CurrentLap)
	}
	if last.TotalLaps != 66 {
		t.Errorf("last: TotalLaps = %d, want 66 (carried forward from record 0)", last.TotalLaps)
	}

	mid := progress[32]
	if mid.TotalLaps != 66 {
		t.Errorf("middle entry: TotalLaps = %d, want 66 (carried forward)", mid.TotalLaps)
	}
}

func TestParseLapCount_CarriesForwardAcrossAChange(t *testing.T) {

	data := []byte(
		`00:00:01.000{"CurrentLap":1,"TotalLaps":66}` + "\r\n" +
			`00:01:00.000{"CurrentLap":2}` + "\r\n" +
			`00:02:00.000{"CurrentLap":3,"TotalLaps":60}` + "\r\n" +
			`00:03:00.000{"CurrentLap":4}` + "\r\n",
	)

	progress, err := ParseLapCount(data)
	if err != nil {
		t.Fatalf("ParseLapCount: %v", err)
	}
	if len(progress) != 4 {
		t.Fatalf("got %d entries, want 4", len(progress))
	}

	if progress[0].TotalLaps != 66 || progress[1].TotalLaps != 66 {
		t.Errorf("entries before the revision should keep TotalLaps=66, got %d and %d",
			progress[0].TotalLaps, progress[1].TotalLaps)
	}
	if progress[2].TotalLaps != 60 || progress[3].TotalLaps != 60 {
		t.Errorf("entries at/after the revision should have TotalLaps=60, got %d and %d",
			progress[2].TotalLaps, progress[3].TotalLaps)
	}
}

func TestParseLapCount_RejectsMalformed(t *testing.T) {
	if _, err := ParseLapCount([]byte("00:00:00.000not json\r\n")); err == nil {
		t.Fatal("expected an error for a malformed payload")
	}
}
