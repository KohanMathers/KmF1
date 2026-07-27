package feeds

import "time"

const lapAccuracyTolerance = 3 * time.Millisecond

func CheckLapAccuracy(laps map[string][]LapData) int {
	integrityErrors := 0

	for _, ls := range laps {
		var prev *LapData
		for i := range ls {
			l := &ls[i]

			check1 := lapAccuracyCheck1(l)
			check2 := check1 && lapAccuracySectorSumMatches(l)
			check3 := prev == nil || prev.TrackStatus != "4"
			check4 := true
			if prev != nil && l.LapTime != nil {
				check4 = closeEnoughDuration(l.Timestamp-prev.Timestamp, *l.LapTime)
			}

			if !check2 || !check4 {
				integrityErrors++
			}
			l.IsAccurate = check1 && check2 && check3 && check4

			prev = l
		}
	}

	return integrityErrors
}

func lapAccuracyCheck1(l *LapData) bool {
	if l.PitInTime != nil || l.PitOutTime != nil {
		return false
	}
	if l.FastF1Generated {
		return false
	}
	switch l.TrackStatus {
	case "1", "2", "12", "21":
	default:
		return false
	}
	return l.LapTime != nil && l.Sector1Time != nil && l.Sector2Time != nil && l.Sector3Time != nil
}

func lapAccuracySectorSumMatches(l *LapData) bool {
	if l.Sector1Time == nil || l.Sector2Time == nil || l.Sector3Time == nil || l.LapTime == nil {
		return false
	}
	sum := *l.Sector1Time + *l.Sector2Time + *l.Sector3Time
	return closeEnoughDuration(sum, *l.LapTime)
}

func closeEnoughDuration(a, b time.Duration) bool {
	return absDuration(a-b) <= lapAccuracyTolerance
}
