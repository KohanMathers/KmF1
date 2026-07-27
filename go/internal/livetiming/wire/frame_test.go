package wire

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/KohanMathers/KmF1/go/internal/testfixture"
)

func readTestdata(t *testing.T, name string) []byte {
	t.Helper()
	data := testfixture.Read(t, "testdata/"+name)
	return data
}

func TestSplitFrames_RealCarDataCapture(t *testing.T) {
	data := readTestdata(t, "cardata.jsonstream")

	records, err := SplitFrames(data)
	if err != nil {
		t.Fatalf("SplitFrames: %v", err)
	}
	if len(records) != 3 {
		t.Fatalf("got %d records, want 3", len(records))
	}

	want := []time.Duration{
		1*time.Minute + 16*time.Second + 579*time.Millisecond,
		1*time.Minute + 17*time.Second + 579*time.Millisecond,
		1*time.Minute + 18*time.Second + 858*time.Millisecond,
	}
	for i, r := range records {
		if r.Elapsed != want[i] {
			t.Errorf("record %d: elapsed = %v, want %v", i, r.Elapsed, want[i])
		}
		if len(r.Payload) == 0 {
			t.Errorf("record %d: empty payload", i)
		}
	}
}

func TestSplitFrames_RejectsShortLine(t *testing.T) {
	_, err := SplitFrames([]byte("00:01\r\n"))
	if err == nil {
		t.Fatal("expected an error for a line shorter than a frame prefix")
	}
}

func TestSplitFrames_RejectsMalformedPrefix(t *testing.T) {
	_, err := SplitFrames([]byte("AA:BB:CC.DDD{}\r\n"))
	if err == nil {
		t.Fatal("expected an error for a non-numeric elapsed-time prefix")
	}
}

func TestSplitFrames_SkipsTrailingBlankLine(t *testing.T) {
	records, err := SplitFrames([]byte("00:00:01.000{}\r\n"))
	if err != nil {
		t.Fatalf("SplitFrames: %v", err)
	}
	if len(records) != 1 {
		t.Fatalf("got %d records, want 1", len(records))
	}
}

func TestTrimBOM(t *testing.T) {
	withBOM := append([]byte{0xEF, 0xBB, 0xBF}, []byte(`{"Status":"Complete"}`)...)
	got := TrimBOM(withBOM)
	if string(got) != `{"Status":"Complete"}` {
		t.Fatalf("got %q after TrimBOM", got)
	}

	noBOM := []byte(`{"Status":"Complete"}`)
	if got := TrimBOM(noBOM); string(got) != `{"Status":"Complete"}` {
		t.Fatalf("got %q, TrimBOM should be a no-op without a BOM", got)
	}
}

func TestTrimBOM_RealArchiveStatus(t *testing.T) {
	data := readTestdata(t, "archivestatus.json")

	var doc struct {
		Status string `json:"Status"`
	}
	if err := json.Unmarshal(TrimBOM(data), &doc); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if doc.Status != "Complete" {
		t.Fatalf("Status = %q, want %q", doc.Status, "Complete")
	}
}

func TestTrimBOM_RealSessionData(t *testing.T) {
	data := readTestdata(t, "sessiondata.json")

	var doc struct {
		Series []struct {
			Utc string `json:"Utc"`
			Lap int    `json:"Lap"`
		} `json:"Series"`
		StatusSeries []map[string]any `json:"StatusSeries"`
	}
	if err := json.Unmarshal(TrimBOM(data), &doc); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(doc.Series) == 0 {
		t.Fatal("expected a non-empty Series (lap progress) in the real fixture")
	}
	if len(doc.StatusSeries) == 0 {
		t.Fatal("expected a non-empty StatusSeries (track/session status log) in the real fixture")
	}
	if doc.Series[0].Lap != 1 {
		t.Errorf("first Series entry: Lap = %d, want 1", doc.Series[0].Lap)
	}
}
