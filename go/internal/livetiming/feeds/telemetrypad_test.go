package feeds

import (
	"testing"
)

func TestPadTelemetry_RealCapture_AllDriversReachReferenceLength(t *testing.T) {
	carData, posData := readTelemetryFixtures(t)
	telemetry, err := ParseTelemetry(carData, posData)
	if err != nil {
		t.Fatalf("ParseTelemetry: %v", err)
	}

	maxLen := 0
	for _, ts := range telemetry {
		if len(ts) > maxLen {
			maxLen = len(ts)
		}
	}

	PadTelemetry(telemetry)

	for drv, ts := range telemetry {
		if len(ts) != maxLen {
			t.Errorf("driver %s: length = %d after padding, want %d", drv, len(ts), maxLen)
		}

		for i := 1; i < len(ts); i++ {
			if ts[i].Timestamp < ts[i-1].Timestamp {
				t.Fatalf("driver %s: sample %d out of order after padding", drv, i)
			}
		}
	}
}

func TestPadTelemetry_PadsShorterDriverToReferenceTimestamps(t *testing.T) {
	telemetry := map[string][]Telemetry{
		"1": {
			{DriverNumber: "1", Timestamp: 0, RPM: intPtr(10000)},
			{DriverNumber: "1", Timestamp: 10, RPM: intPtr(10500)},
			{DriverNumber: "1", Timestamp: 20, RPM: intPtr(11000)},
		},
		"2": {
			{DriverNumber: "2", Timestamp: 0, RPM: intPtr(9000)},
			{DriverNumber: "2", Timestamp: 20, RPM: intPtr(9500)},
		},
	}

	PadTelemetry(telemetry)

	drv2 := telemetry["2"]
	if len(drv2) != 3 {
		t.Fatalf("driver 2: length = %d, want 3", len(drv2))
	}
	if *drv2[0].RPM != 9000 || *drv2[2].RPM != 9500 {
		t.Errorf("real samples were altered: %+v", drv2)
	}
	padded := drv2[1]
	if padded.Timestamp != 10 {
		t.Fatalf("padded sample Timestamp = %v, want 10", padded.Timestamp)
	}
	if padded.DriverNumber != "2" {
		t.Errorf("padded sample DriverNumber = %q, want %q", padded.DriverNumber, "2")
	}
	if *padded.RPM != 0 || *padded.Speed != 0 || *padded.Gear != 0 || *padded.Throttle != 0 {
		t.Errorf("padded sample car channels = %+v, want all zero", padded)
	}
	if *padded.Braking {
		t.Error("padded sample Braking = true, want false")
	}
	if padded.DRS != "DISABLED" {
		t.Errorf("padded sample DRS = %q, want DISABLED (mapped from raw 0)", padded.DRS)
	}
	if *padded.PositionX != 0 || *padded.PositionY != 0 || *padded.PositionZ != 0 {
		t.Errorf("padded sample position = %+v, want all zero", padded)
	}
}

func TestPadTelemetry_ReferenceDriverUnchanged(t *testing.T) {
	telemetry := map[string][]Telemetry{
		"1": {{DriverNumber: "1", Timestamp: 0, RPM: intPtr(10000)}, {DriverNumber: "1", Timestamp: 10, RPM: intPtr(11000)}},
		"2": {{DriverNumber: "2", Timestamp: 0, RPM: intPtr(9000)}},
	}
	PadTelemetry(telemetry)
	if len(telemetry["1"]) != 2 {
		t.Errorf("reference driver length changed: got %d, want 2", len(telemetry["1"]))
	}
}

func TestPadTelemetry_SingleDriverIsNoop(t *testing.T) {
	telemetry := map[string][]Telemetry{
		"1": {{DriverNumber: "1", Timestamp: 0}},
	}
	PadTelemetry(telemetry)
	if len(telemetry["1"]) != 1 {
		t.Errorf("single-driver map should be untouched, got length %d", len(telemetry["1"]))
	}
}
