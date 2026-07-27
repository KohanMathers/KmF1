package feeds

import (
	"testing"
	"time"

	"github.com/KohanMathers/KmF1/go/internal/testfixture"
)

func TestParseTyreStints_RealCapture_StintAgeProgresses(t *testing.T) {
	data := testfixture.Read(t, "testdata/timingappdata.jsonstream")

	result, err := ParseTyreStints(data)
	if err != nil {
		t.Fatalf("ParseTyreStints: %v", err)
	}

	drv4 := result["4"]
	if len(drv4) < 4 {
		t.Fatalf("got %d entries for driver 4, want at least 4", len(drv4))
	}

	want := []TyreStint{
		{
			DriverNumber: "4", Compound: "SOFT", New: true, TyresChanged: true,
			Stint: 1, StintAge: 0, StartLap: 0, LapFlags: 0,
			Timestamp: 50*time.Minute + 55*time.Second + 29*time.Millisecond,
		},
		{

			DriverNumber: "4", Compound: "SOFT", New: true, TyresChanged: true,
			Stint: 1, StintAge: 1, StartLap: 0, LapFlags: 0,
			Timestamp: 58*time.Minute + 25*time.Second + 84*time.Millisecond,
		},
		{
			DriverNumber: "4", Compound: "SOFT", New: true, TyresChanged: true,
			Stint: 1, StintAge: 2, StartLap: 0, LapFlags: 0,
			Timestamp: 59*time.Minute + 45*time.Second + 92*time.Millisecond,
		},
		{

			DriverNumber: "4", Compound: "SOFT", New: true, TyresChanged: true,
			Stint: 1, StintAge: 2, StartLap: 0, LapFlags: 1,
			Timestamp: 59*time.Minute + 45*time.Second + 251*time.Millisecond,
		},
	}

	for i, w := range want {
		if drv4[i] != w {
			t.Errorf("entry %d = %+v, want %+v", i, drv4[i], w)
		}
	}
}

func TestParseTyreStints_RealCapture_PitStopStartsNewStint(t *testing.T) {
	data := testfixture.Read(t, "testdata/timingappdata.jsonstream")

	result, err := ParseTyreStints(data)
	if err != nil {
		t.Fatalf("ParseTyreStints: %v", err)
	}

	drv24 := result["24"]

	first := drv24[0]
	if first.Stint != 1 || first.Compound != "SOFT" || first.New != true {
		t.Errorf("first entry = %+v, want Stint 1 on fresh SOFT", first)
	}

	var foundSecondStint bool
	for _, ts := range drv24 {
		if ts.Stint == 2 {
			foundSecondStint = true
			if ts.Compound != "MEDIUM" || !ts.New || ts.StintAge != 0 || ts.StartLap != 0 {
				t.Errorf("second stint entry = %+v, want fresh MEDIUM at StintAge 0", ts)
			}
			break
		}
	}
	if !foundSecondStint {
		t.Fatal("expected to find a Stint 2 entry for driver 24 (their real pit stop to MEDIUM)")
	}
}

func TestParseTyreStints_TyresChangedIsInvertedFromRawField(t *testing.T) {

	data := []byte(
		`00:00:01.000{"Lines":{"1":{"Stints":[{"Compound":"SOFT","New":"false","TyresNotChanged":"1","TotalLaps":5,"StartLaps":5}]}}}` + "\r\n",
	)

	result, err := ParseTyreStints(data)
	if err != nil {
		t.Fatalf("ParseTyreStints: %v", err)
	}

	stints := result["1"]
	if len(stints) != 1 {
		t.Fatalf("got %d entries, want 1", len(stints))
	}
	if stints[0].TyresChanged {
		t.Error("TyresNotChanged:\"1\" should produce TyresChanged:false")
	}
}

func TestParseTyreStints_IgnoresPatchesWithoutStints(t *testing.T) {
	data := []byte(`00:00:01.000{"Lines":{"1":{"Line":3}}}` + "\r\n")

	result, err := ParseTyreStints(data)
	if err != nil {
		t.Fatalf("ParseTyreStints: %v", err)
	}
	if len(result["1"]) != 0 {
		t.Errorf("got %d entries, want 0 for a patch with no Stints field", len(result["1"]))
	}
}

func TestParseTyreStints_RejectsMalformed(t *testing.T) {
	if _, err := ParseTyreStints([]byte("00:00:00.000not json\r\n")); err == nil {
		t.Fatal("expected an error for a malformed payload")
	}
}
