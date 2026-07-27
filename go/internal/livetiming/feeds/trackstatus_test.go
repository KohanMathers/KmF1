package feeds

import (
	"os"
	"testing"
	"time"

	"github.com/KohanMathers/KmF1/go/internal/livetiming/wire"
)

func TestParseTrackStatus_RealCapture(t *testing.T) {
	data, err := os.ReadFile("testdata/trackstatus.jsonstream")
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}

	events, err := ParseTrackStatus(data)
	if err != nil {
		t.Fatalf("ParseTrackStatus: %v", err)
	}

	want := []TrackStatus{
		{Status: "2", Message: "Yellow", Timestamp: 0},
		{Status: "1", Message: "AllClear", Timestamp: 10*time.Minute + 895*time.Millisecond},
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

func TestParseTrackStatus_RejectsMalformed(t *testing.T) {
	if _, err := ParseTrackStatus([]byte("00:00:00.000not json\r\n")); err == nil {
		t.Fatal("expected an error for a malformed payload")
	}
}

func TestTrackStatusDecoder_IncrementalFeedMatchesBatchParse(t *testing.T) {
	data, err := os.ReadFile("testdata/trackstatus.jsonstream")
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}
	want, err := ParseTrackStatus(data)
	if err != nil {
		t.Fatalf("ParseTrackStatus: %v", err)
	}

	records, err := wire.SplitFrames(data)
	if err != nil {
		t.Fatalf("split frames: %v", err)
	}

	var d TrackStatusDecoder
	if got := d.Snapshot(); len(got) != 0 {
		t.Fatalf("zero-value decoder Snapshot() = %v, want empty", got)
	}

	for i, rec := range records {
		if err := d.Feed(rec.Elapsed, rec.Payload); err != nil {
			t.Fatalf("Feed record %d: %v", i, err)
		}

		if got := d.Snapshot(); len(got) != i+1 {
			t.Fatalf("after feeding record %d, Snapshot() has %d events, want %d", i, len(got), i+1)
		}
	}

	got := d.Snapshot()
	if len(got) != len(want) {
		t.Fatalf("incremental Snapshot() has %d events, want %d (from batch ParseTrackStatus)", len(got), len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("event %d = %+v, want %+v", i, got[i], want[i])
		}
	}
}
