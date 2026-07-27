package feeds

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/KohanMathers/KmF1/go/internal/livetiming/wire"
	"github.com/KohanMathers/KmF1/go/internal/testfixture"
)

func TestParseDriverList_RealCapture(t *testing.T) {
	data := testfixture.Read(t, "testdata/driverlist.jsonstream")

	drivers, err := ParseDriverList(data)
	if err != nil {
		t.Fatalf("ParseDriverList: %v", err)
	}

	if len(drivers) != 20 {
		t.Fatalf("got %d drivers, want 20", len(drivers))
	}

	ver, ok := drivers["1"]
	if !ok {
		t.Fatal("driver 1 (VER) not found")
	}
	if ver.DriverNumber != "1" {
		t.Errorf("VER: DriverNumber = %q, want %q", ver.DriverNumber, "1")
	}
	if ver.Abbreviation != "VER" {
		t.Errorf("VER: Abbreviation = %q, want %q", ver.Abbreviation, "VER")
	}
	if ver.FullName != "Max VERSTAPPEN" {
		t.Errorf("VER: FullName = %q", ver.FullName)
	}
	if ver.TeamName != "Red Bull Racing" {
		t.Errorf("VER: TeamName = %q", ver.TeamName)
	}
	if ver.TeamColour != "3671C6" {
		t.Errorf("VER: TeamColour = %q", ver.TeamColour)
	}
	if ver.Reference != "MAXVER01" {
		t.Errorf("VER: Reference = %q", ver.Reference)
	}
	if ver.CountryCode != "NED" {
		t.Errorf("VER: CountryCode = %q", ver.CountryCode)
	}

	if ver.Line != 1 {
		t.Errorf("VER: Line = %d, want 1 (final patched value, not the initial 2)", ver.Line)
	}

	nor, ok := drivers["4"]
	if !ok {
		t.Fatal("driver 4 (NOR) not found")
	}
	if nor.Line != 2 {
		t.Errorf("NOR: Line = %d, want 2 (final patched value)", nor.Line)
	}
}

func TestParseDriverList_DropsLateDeclaration(t *testing.T) {

	record0 := `00:00:01.000` + mustJSON(t, map[string]any{
		"1": map[string]any{"RacingNumber": "1", "Tla": "VER"},
	})
	lateRecord := `01:30:00.000` + mustJSON(t, map[string]any{
		"99": map[string]any{"RacingNumber": "99", "Tla": "ZZZ"},
	})
	data := []byte(record0 + "\r\n" + lateRecord + "\r\n")

	drivers, err := ParseDriverList(data)
	if err != nil {
		t.Fatalf("ParseDriverList: %v", err)
	}

	if _, ok := drivers["1"]; !ok {
		t.Error("driver 1 (present from the start) should not be dropped")
	}
	if _, ok := drivers["99"]; ok {
		t.Error("driver 99 (first seen 90 minutes in) should be dropped as a late declaration")
	}
	if len(drivers) != 1 {
		t.Errorf("got %d drivers, want 1", len(drivers))
	}
}

func TestParseDriverList_KeepsDeclarationJustUnderCutoff(t *testing.T) {
	record := `00:59:59.000` + mustJSON(t, map[string]any{
		"1": map[string]any{"RacingNumber": "1", "Tla": "VER"},
	})
	data := []byte(record + "\r\n")

	drivers, err := ParseDriverList(data)
	if err != nil {
		t.Fatalf("ParseDriverList: %v", err)
	}
	if _, ok := drivers["1"]; !ok {
		t.Error("a driver first seen just under the 60-minute cutoff should be kept")
	}
}

func TestDriverListDecoder_IncrementalFeedMatchesBatchParse(t *testing.T) {
	data := testfixture.Read(t, "testdata/driverlist.jsonstream")
	want, err := ParseDriverList(data)
	if err != nil {
		t.Fatalf("ParseDriverList: %v", err)
	}

	records, err := wire.SplitFrames(data)
	if err != nil {
		t.Fatalf("split frames: %v", err)
	}

	var d DriverListDecoder
	for i, rec := range records {
		if err := d.Feed(rec.Elapsed, rec.Payload); err != nil {
			t.Fatalf("Feed record %d: %v", i, err)
		}
	}

	got, err := d.Snapshot()
	if err != nil {
		t.Fatalf("Snapshot: %v", err)
	}
	if len(got) != len(want) {
		t.Fatalf("incremental Snapshot() has %d drivers, want %d (from batch ParseDriverList)", len(got), len(want))
	}
	for drv, wantInfo := range want {
		gotInfo, ok := got[drv]
		if !ok {
			t.Errorf("driver %s missing from incremental Snapshot()", drv)
			continue
		}
		if !reflect.DeepEqual(*gotInfo, *wantInfo) {
			t.Errorf("driver %s = %+v, want %+v", drv, *gotInfo, *wantInfo)
		}
	}
}

func mustJSON(t *testing.T, v any) string {
	t.Helper()
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("marshal test fixture: %v", err)
	}
	return string(b)
}
