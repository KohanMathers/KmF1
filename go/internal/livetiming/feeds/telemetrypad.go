package feeds

import (
	"sort"
	"time"
)

func PadTelemetry(telemetry map[string][]Telemetry) {
	if len(telemetry) < 2 {
		return
	}

	refDrv := ""
	maxLen := -1
	for drv, ts := range telemetry {
		if len(ts) > maxLen {
			maxLen = len(ts)
			refDrv = drv
		}
	}

	hasCar, hasPos := telemetryHalves(telemetry)

	refTimestamps := make([]time.Duration, len(telemetry[refDrv]))
	for i, t := range telemetry[refDrv] {
		refTimestamps[i] = t.Timestamp
	}

	for drv, ts := range telemetry {
		if drv == refDrv {
			continue
		}

		existing := make(map[time.Duration]bool, len(ts))
		for _, t := range ts {
			existing[t.Timestamp] = true
		}

		padded := false
		out := make([]Telemetry, len(ts), len(ts)+len(refTimestamps))
		copy(out, ts)
		for _, refTs := range refTimestamps {
			if !existing[refTs] {
				out = append(out, paddingSample(drv, refTs, hasCar, hasPos))
				padded = true
			}
		}

		if padded {
			sort.Slice(out, func(i, j int) bool { return out[i].Timestamp < out[j].Timestamp })
			telemetry[drv] = out
		}
	}
}

func telemetryHalves(telemetry map[string][]Telemetry) (hasCar, hasPos bool) {
	for _, ts := range telemetry {
		for _, t := range ts {
			if t.RPM != nil || t.Speed != nil || t.Gear != nil || t.Throttle != nil || t.Braking != nil || t.DRSNo != nil {
				hasCar = true
			}
			if t.PositionX != nil || t.PositionY != nil || t.PositionZ != nil {
				hasPos = true
			}
			if hasCar && hasPos {
				return
			}
		}
	}
	return
}

func paddingSample(driverNumber string, ts time.Duration, hasCar, hasPos bool) Telemetry {
	t := Telemetry{DriverNumber: driverNumber, Timestamp: ts}
	if hasCar {
		rpm, speed, gear, throttle, drsNo := 0, 0, 0, 0, 0
		braking := false
		drs := mapDRS(0)
		t.RPM, t.Speed, t.Gear, t.Throttle = &rpm, &speed, &gear, &throttle
		t.Braking, t.DRS, t.DRSNo = &braking, &drs, &drsNo
	}
	if hasPos {
		x, y, z := 0, 0, 0
		t.PositionX, t.PositionY, t.PositionZ = &x, &y, &z
	}
	return t
}
