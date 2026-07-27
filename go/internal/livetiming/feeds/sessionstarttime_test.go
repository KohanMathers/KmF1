package feeds

import (
	"testing"
	"time"

	"github.com/KohanMathers/KmF1/go/internal/testfixture"
)

func TestSessionStartTime_RealCapture(t *testing.T) {
	data := testfixture.Read(t, "testdata/sessionstatus.jsonstream")
	events, err := ParseSessionStatus(data)
	if err != nil {
		t.Fatalf("ParseSessionStatus: %v", err)
	}

	got, ok := SessionStartTime(events)
	if !ok {
		t.Fatal("expected a Started transition to be found")
	}
	want := 56*time.Minute + 57*time.Second + 95*time.Millisecond
	if got != want {
		t.Errorf("SessionStartTime = %v, want %v", got, want)
	}
}

func TestSessionStartTime_NoStartedEvent(t *testing.T) {
	events := []SessionStatusEvent{
		{Status: "Inactive", Timestamp: 0},
		{Status: "Aborted", Timestamp: time.Second},
	}
	if _, ok := SessionStartTime(events); ok {
		t.Error("expected ok=false when no Started event is present")
	}
}
