package feeds

import "time"

const syntheticLapMargin = 0

func GenerateSyntheticRetirementLaps(
	laps map[string][]LapData,
	statusEvents []SessionStatusEvent,
	trackStatuses []TrackStatus,
	telemetry map[string][]Telemetry,
	totalLaps int,
	haveTotalLaps bool,
) {
	for drv, drvLaps := range laps {
		if len(drvLaps) == 0 {
			continue
		}

		if len(drvLaps) == 1 && drvLaps[0].FastF1Generated {
			continue
		}

		last := drvLaps[len(drvLaps)-1]
		refTime := last.Timestamp

		var nextStatuses, prevStatuses []SessionStatusEvent
		for _, e := range statusEvents {
			if e.Timestamp > refTime {
				nextStatuses = append(nextStatuses, e)
			} else {
				prevStatuses = append(prevStatuses, e)
			}
		}

		if haveTotalLaps && last.LapNo == totalLaps {
			continue
		}

		if anyStatus(prevStatuses, "Finished") {
			continue
		}

		if len(nextStatuses) == 0 || !anyStatus(nextStatuses, "Finished") {
			continue
		}

		if last.PitInTime != nil {
			continue
		}

		if haveTotalLaps && len(drvLaps) >= totalLaps {
			continue
		}

		if len(drvLaps) >= 2 && drvLaps[len(drvLaps)-2].PitInTime != nil && last.PitOutTime == nil {
			continue
		}

		nextStatus := nextStatuses[0]
		var assumedEndTime time.Duration
		switch {
		case nextStatus.Status == "Aborted":
			assumedEndTime = nextStatus.Timestamp
		default:
			if zeroSpeedTime, ok := firstZeroSpeedAfter(telemetry[drv], refTime); ok {
				assumedEndTime = zeroSpeedTime + syntheticLapMargin
			} else {
				assumedEndTime = refTime + 150*time.Second
			}
		}

		newLap := LapData{
			DriverNumber:    drv,
			LapNo:           last.LapNo + 1,
			LapStartTime:    durPtr(last.Timestamp),
			Timestamp:       assumedEndTime,
			Position:        nil,
			FastF1Generated: true,
			IsAccurate:      false,
		}
		AddTrackStatusToLap(&newLap, trackStatuses)

		laps[drv] = append(drvLaps, newLap)
	}
}

func anyStatus(events []SessionStatusEvent, status string) bool {
	for _, e := range events {
		if e.Status == status {
			return true
		}
	}
	return false
}

func firstZeroSpeedAfter(samples []Telemetry, t time.Duration) (time.Duration, bool) {
	for _, s := range samples {
		if s.Timestamp > t && s.Speed != nil && *s.Speed == 0 {
			return s.Timestamp, true
		}
	}
	return 0, false
}
