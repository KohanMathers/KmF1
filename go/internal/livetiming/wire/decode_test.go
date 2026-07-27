package wire

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestDecode_RealCompressedCarData(t *testing.T) {
	data := readTestdata(t, "cardata.jsonstream")
	records, err := SplitFrames(data)
	if err != nil {
		t.Fatalf("SplitFrames: %v", err)
	}

	wantUtc := []string{
		"2024-06-23T12:07:33.8963104Z",
		"2024-06-23T12:07:34.2162628Z",
		"2024-06-23T12:07:35.1761198Z",
	}

	for i, r := range records {
		raw, err := Decode(r.Payload, true)
		if err != nil {
			t.Fatalf("record %d: Decode: %v", i, err)
		}

		var doc struct {
			Entries []struct {
				Utc  string `json:"Utc"`
				Cars map[string]struct{ Channels map[string]int }
			} `json:"Entries"`
		}
		if err := json.Unmarshal(raw, &doc); err != nil {
			t.Fatalf("record %d: unmarshal decoded JSON: %v\nraw: %s", i, err, raw)
		}
		if len(doc.Entries) == 0 {
			t.Fatalf("record %d: no Entries in decoded payload", i)
		}
		if doc.Entries[0].Utc != wantUtc[i] {
			t.Errorf("record %d: Utc = %q, want %q", i, doc.Entries[0].Utc, wantUtc[i])
		}
		if len(doc.Entries[0].Cars) == 0 {
			t.Errorf("record %d: no Cars in decoded payload", i)
		}
	}
}

func TestDecode_UncompressedRawObject(t *testing.T) {

	raw, err := Decode(`{"Remaining":"02:00:00"}`, false)
	if err != nil {
		t.Fatalf("Decode: %v", err)
	}
	if !strings.Contains(string(raw), "02:00:00") {
		t.Fatalf("got %q", raw)
	}
}

func TestDecode_RealExtrapolatedClockRecords(t *testing.T) {
	data := readTestdata(t, "extrapolatedclock.jsonstream")
	records, err := SplitFrames(data)
	if err != nil {
		t.Fatalf("SplitFrames: %v", err)
	}
	if len(records) != 3 {
		t.Fatalf("got %d records, want 3", len(records))
	}

	for i, r := range records {
		raw, err := Decode(r.Payload, false)
		if err != nil {
			t.Fatalf("record %d: Decode: %v", i, err)
		}

		var doc struct {
			Utc           string `json:"Utc"`
			Remaining     string `json:"Remaining"`
			Extrapolating *bool  `json:"Extrapolating"`
		}
		if err := json.Unmarshal(raw, &doc); err != nil {
			t.Fatalf("record %d: unmarshal: %v", i, err)
		}
		if doc.Utc == "" || doc.Remaining == "" {
			t.Errorf("record %d: missing Utc/Remaining: %+v", i, doc)
		}
	}

	raw, _ := Decode(records[1].Payload, false)
	var generic map[string]any
	_ = json.Unmarshal(raw, &generic)
	if _, present := generic["Extrapolating"]; present {
		t.Fatal("record 1 should not carry an Extrapolating key at all")
	}
}

func TestDecode_RejectsEmptyPayload(t *testing.T) {
	if _, err := Decode("", false); err == nil {
		t.Fatal("expected an error for an empty payload")
	}
}

func TestDecode_RejectsGarbagePayload(t *testing.T) {
	if _, err := Decode("not json at all", false); err == nil {
		t.Fatal("expected an error for a payload that's neither an object/array nor a quoted string")
	}
}
