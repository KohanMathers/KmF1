package feeds

import (
	"testing"
	"time"

	"github.com/KohanMathers/KmF1/go/internal/testfixture"
)

func TestParseSessionStatus_RealCapture(t *testing.T) {
	data := testfixture.Read(t, "testdata/sessionstatus.jsonstream")

	events, err := ParseSessionStatus(data)
	if err != nil {
		t.Fatalf("ParseSessionStatus: %v", err)
	}

	want := []SessionStatusEvent{
		{Status: "Inactive", Timestamp: 4*time.Second + 966*time.Millisecond},
		{Status: "Started", Timestamp: 56*time.Minute + 57*time.Second + 95*time.Millisecond},
		{Status: "Finished", Timestamp: 2*time.Hour + 25*time.Minute + 17*time.Second + 563*time.Millisecond},
		{Status: "Finalised", Timestamp: 2*time.Hour + 33*time.Minute + 40*time.Second + 490*time.Millisecond},
		{Status: "Ends", Timestamp: 2*time.Hour + 33*time.Minute + 41*time.Second + 969*time.Millisecond},
	}
	if len(events) != len(want) {
		t.Fatalf("got %d events, want %d", len(events), len(want))
	}
	for i, w := range want {
		if events[i] != w {
			t.Errorf("event %d = %+v, want %+v", i, events[i], w)
		}
	}
}

func TestParseSessionStatus_RejectsMalformed(t *testing.T) {
	if _, err := ParseSessionStatus([]byte("00:00:00.000not json\r\n")); err == nil {
		t.Fatal("expected an error for a malformed payload")
	}
}
