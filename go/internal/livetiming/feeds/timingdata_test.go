package feeds

import (
	"reflect"
	"testing"
	"time"

	"github.com/KohanMathers/KmF1/go/internal/livetiming/wire"
	"github.com/KohanMathers/KmF1/go/internal/testfixture"
)

func TestSplitTimingDataByDriver_RealCapture(t *testing.T) {
	data := testfixture.Read(t, "testdata/timingdata.jsonstream")

	perDriver, err := SplitTimingDataByDriver(data)
	if err != nil {
		t.Fatalf("SplitTimingDataByDriver: %v", err)
	}

	if len(perDriver) != 20 {
		t.Fatalf("got %d drivers, want 20", len(perDriver))
	}

	drv4 := perDriver["4"]
	if len(drv4) != 193 {
		t.Fatalf("driver 4: got %d patches, want 193", len(drv4))
	}

	first := drv4[0]
	wantElapsed := 8*time.Second + 720*time.Millisecond
	if first.Elapsed != wantElapsed {
		t.Errorf("driver 4 first patch: Elapsed = %v, want %v", first.Elapsed, wantElapsed)
	}
	if first.Data["Position"] != "1" {
		t.Errorf("driver 4 first patch: Position = %v, want %q", first.Data["Position"], "1")
	}
	if first.Data["InPit"] != true {
		t.Errorf("driver 4 first patch: InPit = %v, want true", first.Data["InPit"])
	}

	if _, present := first.Data["NumberOfLaps"]; present {
		t.Error("driver 4 first patch should not carry NumberOfLaps")
	}

	var lap2 *TimingPatch
	for i := range drv4 {
		if n, ok := drv4[i].Data["NumberOfLaps"]; ok && n == float64(2) {
			lap2 = &drv4[i]
			break
		}
	}
	if lap2 == nil {
		t.Fatal("driver 4: no patch with NumberOfLaps=2 found")
	}
	wantLap2Elapsed := 59*time.Minute + 41*time.Second + 791*time.Millisecond
	if lap2.Elapsed != wantLap2Elapsed {
		t.Errorf("driver 4 NumberOfLaps=2 patch: Elapsed = %v, want %v", lap2.Elapsed, wantLap2Elapsed)
	}
	if _, present := lap2.Data["Position"]; present {
		t.Error("the NumberOfLaps=2 patch should not carry a Position field — it wasn't in the real record, so merging must not have introduced one")
	}
	sectors, ok := lap2.Data["Sectors"].(map[string]any)
	if !ok {
		t.Fatalf("NumberOfLaps=2 patch: Sectors is %T, want map[string]any", lap2.Data["Sectors"])
	}
	if _, ok := sectors["2"]; !ok {
		t.Error("NumberOfLaps=2 patch: expected sector index \"2\" in Sectors")
	}
}

func TestTimingDataDecoder_IncrementalFeedMatchesBatchParse(t *testing.T) {
	data := testfixture.Read(t, "testdata/timingdata.jsonstream")
	wantLaps, err := ParseLapData(data)
	if err != nil {
		t.Fatalf("ParseLapData: %v", err)
	}
	wantIntervals, err := ParseTimingIntervals(data)
	if err != nil {
		t.Fatalf("ParseTimingIntervals: %v", err)
	}

	records, err := wire.SplitFrames(data)
	if err != nil {
		t.Fatalf("split frames: %v", err)
	}

	var d TimingDataDecoder
	for i, rec := range records {
		if err := d.Feed(rec.Elapsed, rec.Payload); err != nil {
			t.Fatalf("Feed record %d: %v", i, err)
		}
	}

	gotLaps := d.Laps()
	if len(gotLaps) != len(wantLaps) {
		t.Fatalf("incremental Laps() has %d drivers, want %d", len(gotLaps), len(wantLaps))
	}
	for drv, want := range wantLaps {
		got, ok := gotLaps[drv]
		if !ok {
			t.Errorf("driver %s missing from incremental Laps()", drv)
			continue
		}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("driver %s Laps() mismatch:\n got  %+v\n want %+v", drv, got, want)
		}
	}

	gotIntervals := d.Intervals()
	if len(gotIntervals) != len(wantIntervals) {
		t.Fatalf("incremental Intervals() has %d drivers, want %d", len(gotIntervals), len(wantIntervals))
	}
	for drv, want := range wantIntervals {
		got, ok := gotIntervals[drv]
		if !ok {
			t.Errorf("driver %s missing from incremental Intervals()", drv)
			continue
		}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("driver %s Intervals() mismatch:\n got  %+v\n want %+v", drv, got, want)
		}
	}
}

func TestSplitTimingDataByDriver_SkipsEmptyLinesRecord(t *testing.T) {

	data := []byte(`00:00:07.609{"Lines":{},"Withheld":false}` + "\r\n")

	perDriver, err := SplitTimingDataByDriver(data)
	if err != nil {
		t.Fatalf("SplitTimingDataByDriver: %v", err)
	}
	if len(perDriver) != 0 {
		t.Fatalf("got %d drivers from an empty Lines record, want 0", len(perDriver))
	}
}

func TestSplitTimingDataByDriver_PreservesRecordOrderAcrossDrivers(t *testing.T) {

	data := []byte(`00:01:00.000{"Lines":{"1":{"Position":"1"},"2":{"Position":"2"}}}` + "\r\n")

	perDriver, err := SplitTimingDataByDriver(data)
	if err != nil {
		t.Fatalf("SplitTimingDataByDriver: %v", err)
	}
	if len(perDriver) != 2 {
		t.Fatalf("got %d drivers, want 2", len(perDriver))
	}
	for _, drv := range []string{"1", "2"} {
		patches := perDriver[drv]
		if len(patches) != 1 {
			t.Fatalf("driver %s: got %d patches, want 1", drv, len(patches))
		}
		if patches[0].Elapsed != time.Minute {
			t.Errorf("driver %s: Elapsed = %v, want 1m", drv, patches[0].Elapsed)
		}
	}
	if perDriver["1"][0].Data["Position"] != "1" || perDriver["2"][0].Data["Position"] != "2" {
		t.Error("patches were mixed up between drivers")
	}
}

func TestSplitTimingDataByDriver_RejectsMalformed(t *testing.T) {
	if _, err := SplitTimingDataByDriver([]byte("00:00:00.000not json\r\n")); err == nil {
		t.Fatal("expected an error for a malformed payload")
	}
}
