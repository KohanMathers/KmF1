package feeds

import (
	"os"
	"testing"
	"time"
)

func TestParseTeamRadio_RealCapture(t *testing.T) {
	data, err := os.ReadFile("testdata/teamradio.jsonstream")
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}

	captures, err := ParseTeamRadio(data)
	if err != nil {
		t.Fatalf("ParseTeamRadio: %v", err)
	}
	if len(captures) != 8 {
		t.Fatalf("got %d captures, want 8", len(captures))
	}

	first := captures[0]
	if first.DriverNumber != "27" {
		t.Errorf("first capture DriverNumber = %q, want 27", first.DriverNumber)
	}
	if first.Path != "TeamRadio/NICHUL01_27_20240623_142127.mp3" {
		t.Errorf("first capture Path = %q", first.Path)
	}
	wantUtc, _ := time.Parse(time.RFC3339Nano, "2024-06-23T12:21:52.397Z")
	if !first.Utc.Equal(wantUtc) {
		t.Errorf("first capture Utc = %v, want %v", first.Utc, wantUtc)
	}
	wantTimestamp := 15*time.Minute + 30*time.Second + 475*time.Millisecond
	if first.Timestamp != wantTimestamp {
		t.Errorf("first capture Timestamp = %v, want %v", first.Timestamp, wantTimestamp)
	}

	drv20, drv63 := captures[1], captures[2]
	if drv20.DriverNumber != "20" || drv63.DriverNumber != "63" {
		t.Fatalf("expected captures[1]/[2] to be drivers 20/63, got %s/%s", drv20.DriverNumber, drv63.DriverNumber)
	}
	if drv20.Timestamp != drv63.Timestamp {
		t.Errorf("expected both batched captures to share one frame Timestamp, got %v vs %v", drv20.Timestamp, drv63.Timestamp)
	}
	if drv20.Utc.Equal(drv63.Utc) {
		t.Error("expected the two batched captures to have genuinely different Utc values")
	}
	if diff := drv63.Utc.Sub(drv20.Utc); diff != 107*time.Millisecond {
		t.Errorf("Utc difference = %v, want 107ms", diff)
	}
}

func TestParseTeamRadio_RejectsMalformed(t *testing.T) {
	if _, err := ParseTeamRadio([]byte("00:00:00.000not json\r\n")); err == nil {
		t.Fatal("expected an error for a malformed payload")
	}
}
