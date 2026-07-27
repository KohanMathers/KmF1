package feeds

import (
	"reflect"
	"testing"
	"time"

	"github.com/KohanMathers/KmF1/go/internal/livetiming/wire"
	"github.com/KohanMathers/KmF1/go/internal/testfixture"
)

func readTelemetryFixtures(t *testing.T) (carData, posData []byte) {
	t.Helper()
	carData = testfixture.Read(t, "testdata/cardata.jsonstream")
	posData = testfixture.Read(t, "testdata/position.jsonstream")
	return carData, posData
}

func TestParseCarData_RealCapture(t *testing.T) {
	carData, _ := readTelemetryFixtures(t)

	cars, err := parseCarData(carData)
	if err != nil {
		t.Fatalf("parseCarData: %v", err)
	}
	if len(cars) != 20 {
		t.Fatalf("got %d drivers, want 20", len(cars))
	}

	drv4 := cars["4"]
	if len(drv4) != 235 {
		t.Fatalf("driver 4: got %d samples, want 235", len(drv4))
	}

	first := drv4[0]
	wantDate, _ := time.Parse(time.RFC3339Nano, "2024-06-23T13:05:16.8674205Z")
	if !first.date.Equal(wantDate) {
		t.Errorf("date = %v, want %v", first.date, wantDate)
	}
	checks := []struct {
		name string
		got  *int
		want int
	}{
		{"rpm", first.rpm, 12065},
		{"speed", first.speed, 192},
		{"gear", first.gear, 4},
		{"throttle", first.throttle, 100},
		{"drsNo", first.drsNo, 0},
	}
	for _, c := range checks {
		if c.got == nil || *c.got != c.want {
			t.Errorf("%s = %v, want %d", c.name, c.got, c.want)
		}
	}
	if first.braking == nil || *first.braking != false {
		t.Errorf("braking = %v, want false", first.braking)
	}
}

func TestParsePositionData_RealCapture(t *testing.T) {
	_, posData := readTelemetryFixtures(t)

	positions, err := parsePositionData(posData)
	if err != nil {
		t.Fatalf("parsePositionData: %v", err)
	}
	if len(positions) != 20 {
		t.Fatalf("got %d drivers, want 20", len(positions))
	}

	drv4 := positions["4"]
	if len(drv4) != 240 {
		t.Fatalf("driver 4: got %d samples, want 240", len(drv4))
	}

	first := drv4[0]
	wantDate, _ := time.Parse(time.RFC3339Nano, "2024-06-23T13:05:16.8156359Z")
	if !first.date.Equal(wantDate) {
		t.Errorf("date = %v, want %v", first.date, wantDate)
	}
	if first.x == nil || *first.x != -3995 {
		t.Errorf("x = %v, want -3995", first.x)
	}
	if first.y == nil || *first.y != -5879 {
		t.Errorf("y = %v, want -5879", first.y)
	}
	if first.z == nil || *first.z != 1696 {
		t.Errorf("z = %v, want 1696", first.z)
	}
}

func TestMapDRS(t *testing.T) {
	cases := []struct {
		raw  int
		want string
	}{
		{0, "DISABLED"},
		{1, "DISABLED"},
		{2, "UNKNOWN"},
		{3, "UNKNOWN"},
		{8, "ELIGIBLE"},
		{10, "ENABLED"},
		{12, "ENABLED"},
		{14, "ENABLED"},
		{99, "UNKNOWN"},
	}
	for _, c := range cases {
		if got := mapDRS(c.raw); got != c.want {
			t.Errorf("mapDRS(%d) = %q, want %q", c.raw, got, c.want)
		}
	}
}

func TestParseTelemetry_RealCapture_MergesAndForwardFills(t *testing.T) {
	carData, posData := readTelemetryFixtures(t)

	result, err := ParseTelemetry(carData, posData)
	if err != nil {
		t.Fatalf("ParseTelemetry: %v", err)
	}

	drv4 := result["4"]

	if len(drv4) != 235+240 {
		t.Fatalf("got %d merged samples, want %d", len(drv4), 235+240)
	}

	for i := 1; i < len(drv4); i++ {
		if drv4[i].Timestamp < drv4[i-1].Timestamp {
			t.Fatalf("sample %d out of order: %v before %v", i, drv4[i].Timestamp, drv4[i-1].Timestamp)
		}
	}

	if drv4[0].RPM != nil || drv4[0].Speed != nil || drv4[0].DRS != "" {
		t.Errorf("first sample should have nil car fields (no car data yet), got %+v", drv4[0])
	}
	if drv4[0].PositionX == nil || *drv4[0].PositionX != -3995 {
		t.Errorf("first sample PositionX = %v, want -3995", drv4[0].PositionX)
	}

	var firstWithCar *Telemetry
	for i := range drv4 {
		if drv4[i].RPM != nil {
			firstWithCar = &drv4[i]
			break
		}
	}
	if firstWithCar == nil {
		t.Fatal("expected at least one sample with car data")
	}
	if *firstWithCar.RPM != 12065 || *firstWithCar.Speed != 192 {
		t.Errorf("first car sample = %+v, want RPM 12065 / Speed 192", firstWithCar)
	}
	if firstWithCar.DRS != "DISABLED" {
		t.Errorf("DRS = %q, want %q", firstWithCar.DRS, "DISABLED")
	}
	if firstWithCar.PositionX == nil {
		t.Error("first car sample should have forward-filled position data, got nil")
	}
}

func TestParseTelemetry_T0AnchorsToLeastDelayedSample(t *testing.T) {

	t0 := computeT0(
		map[string][]carSample{
			"1": {{date: mustParseRFC3339(t, "2024-01-01T00:00:10Z"), frameElapsed: 10 * time.Second}},
		},
		map[string][]positionSample{
			"1": {{date: mustParseRFC3339(t, "2024-01-01T00:00:09Z"), frameElapsed: 8 * time.Second}},
		},
	)

	want := mustParseRFC3339(t, "2024-01-01T00:00:01Z")
	if !t0.Equal(want) {
		t.Errorf("t0 = %v, want %v", t0, want)
	}
}

func mustParseRFC3339(t *testing.T, s string) time.Time {
	t.Helper()
	tm, err := time.Parse(time.RFC3339, s)
	if err != nil {
		t.Fatalf("parse %q: %v", s, err)
	}
	return tm
}

func TestParseCarData_RejectsMalformed(t *testing.T) {
	if _, err := parseCarData([]byte("00:00:00.000not json\r\n")); err == nil {
		t.Fatal("expected an error for a malformed payload")
	}
}

func TestParsePositionData_RejectsMalformed(t *testing.T) {
	if _, err := parsePositionData([]byte("00:00:00.000not json\r\n")); err == nil {
		t.Fatal("expected an error for a malformed payload")
	}
}

func TestTelemetryDecoder_IncrementalFeedMatchesBatchParse(t *testing.T) {
	carData, posData := readTelemetryFixtures(t)
	want, err := ParseTelemetry(carData, posData)
	if err != nil {
		t.Fatalf("ParseTelemetry: %v", err)
	}

	carRecords, err := wire.SplitFrames(carData)
	if err != nil {
		t.Fatalf("split car frames: %v", err)
	}
	posRecords, err := wire.SplitFrames(posData)
	if err != nil {
		t.Fatalf("split position frames: %v", err)
	}

	var d TelemetryDecoder

	i, j := 0, 0
	for i < len(carRecords) || j < len(posRecords) {
		if i < len(carRecords) {
			if err := d.FeedCarData(carRecords[i].Elapsed, carRecords[i].Payload); err != nil {
				t.Fatalf("FeedCarData record %d: %v", i, err)
			}
			i++
		}
		if j < len(posRecords) {
			if err := d.FeedPosition(posRecords[j].Elapsed, posRecords[j].Payload); err != nil {
				t.Fatalf("FeedPosition record %d: %v", j, err)
			}
			j++
		}
	}

	got := d.Snapshot()
	if len(got) != len(want) {
		t.Fatalf("incremental Snapshot() has %d drivers, want %d", len(got), len(want))
	}
	for drv, wantTelemetry := range want {
		gotTelemetry, ok := got[drv]
		if !ok {
			t.Errorf("driver %s missing from incremental Snapshot()", drv)
			continue
		}
		if !reflect.DeepEqual(gotTelemetry, wantTelemetry) {
			t.Errorf("driver %s telemetry mismatch:\n got  %+v\n want %+v", drv, gotTelemetry, wantTelemetry)
		}
	}
}
