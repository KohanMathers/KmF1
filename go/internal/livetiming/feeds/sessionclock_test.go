package feeds

import (
	"testing"
	"time"

	"github.com/KohanMathers/KmF1/go/internal/testfixture"
)

func TestParseSessionClock_RealCapture(t *testing.T) {
	data := testfixture.Read(t, "testdata/extrapolatedclock.jsonstream")

	clocks, err := ParseSessionClock(data)
	if err != nil {
		t.Fatalf("ParseSessionClock: %v", err)
	}
	if len(clocks) != 3 {
		t.Fatalf("got %d records, want 3 (genuinely sparse real feed)", len(clocks))
	}

	if clocks[0].Remaining != 0 || clocks[0].Extrapolating != false {
		t.Errorf("record 0 = %+v, want Remaining=0, Extrapolating=false", clocks[0])
	}

	if clocks[1].Remaining != 2*time.Hour {
		t.Errorf("record 1 Remaining = %v, want 2h", clocks[1].Remaining)
	}
	if clocks[1].Extrapolating != false {
		t.Errorf("record 1 Extrapolating = %v, want false (carried forward, key absent)", clocks[1].Extrapolating)
	}

	if clocks[2].Extrapolating != true {
		t.Errorf("record 2 Extrapolating = %v, want true", clocks[2].Extrapolating)
	}
	want := time.Hour + 59*time.Minute + 59*time.Second
	if clocks[2].Remaining != want {
		t.Errorf("record 2 Remaining = %v, want %v", clocks[2].Remaining, want)
	}
}

func TestParseSessionClock_RejectsMalformed(t *testing.T) {
	if _, err := ParseSessionClock([]byte("00:00:00.000not json\r\n")); err == nil {
		t.Fatal("expected an error for a malformed payload")
	}
}
