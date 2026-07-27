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
				out = append(out, zeroTelemetrySample(drv, refTs))
				padded = true
			}
		}

		if padded {
			sort.Slice(out, func(i, j int) bool { return out[i].Timestamp < out[j].Timestamp })
			telemetry[drv] = out
		}
	}
}

func zeroTelemetrySample(driverNumber string, ts time.Duration) Telemetry {
	rpm, speed, gear, throttle, drsNo := 0, 0, 0, 0, 0
	braking := false
	x, y, z := 0, 0, 0
	return Telemetry{
		DriverNumber: driverNumber,
		RPM:          &rpm,
		Speed:        &speed,
		Gear:         &gear,
		Throttle:     &throttle,
		Braking:      &braking,
		DRS:          mapDRS(0),
		DRSNo:        &drsNo,
		PositionX:    &x,
		PositionY:    &y,
		PositionZ:    &z,
		Timestamp:    ts,
	}
}
