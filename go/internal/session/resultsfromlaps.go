package session

import (
	"sort"
	"time"

	"github.com/KohanMathers/KmF1/go/internal/jolpica"
	"github.com/KohanMathers/KmF1/go/internal/livetiming/feeds"
)

const quicklapThreshold = 1.07

func CalculateRaceResultsFromLaps(results []jolpica.SessionResults, laps map[string][]feeds.LapData, force bool) []jolpica.SessionResults {
	if !force && anyPositionKnown(results) {
		return results
	}
	if len(laps) == 0 {
		return results
	}

	type finalLap struct {
		driverNumber string
		lapNumber    int
		time         time.Duration
	}
	finals := make([]finalLap, 0, len(laps))
	for drv, ls := range laps {
		if len(ls) == 0 {
			continue
		}
		last := ls[len(ls)-1]
		finals = append(finals, finalLap{drv, last.LapNo, last.Timestamp})
	}

	sort.SliceStable(finals, func(i, j int) bool {
		if finals[i].lapNumber != finals[j].lapNumber {
			return finals[i].lapNumber > finals[j].lapNumber
		}
		return finals[i].time < finals[j].time
	})

	positionByDriver := make(map[string]int, len(finals))
	for i, f := range finals {
		positionByDriver[f.driverNumber] = i + 1
	}

	out := make([]jolpica.SessionResults, len(results))
	copy(out, results)
	for i := range out {
		if pos, ok := positionByDriver[out[i].DriverNumber]; ok {
			p := pos
			out[i].Position = &p
		}
	}
	sortResultsByPosition(out)
	return out
}

func CalculateQualiResultsFromLaps(
	results []jolpica.SessionResults,
	laps map[string][]feeds.LapData,
	statusEvents []feeds.SessionStatusEvent,
	force bool,
) []jolpica.SessionResults {
	if !force && anyPositionKnown(results) {
		return results
	}
	if len(laps) == 0 || len(statusEvents) == 0 {
		return results
	}

	splitTimes := qualifyingSplitTimes(statusEvents)
	if len(splitTimes) < 2 {
		return results
	}

	buckets := splitQualifyingSessions(filterAccurateLaps(laps), splitTimes)

	var qTimes [3]map[string]time.Duration
	for i := range buckets {
		qTimes[i] = minLapTimePerDriver(pickQuicklaps(buckets[i]))
	}

	driverSet := make(map[string]bool, len(laps))
	for drv := range laps {
		driverSet[drv] = true
	}

	type qualiRow struct {
		driverNumber string
		q            [3]*time.Duration
	}
	rows := make([]qualiRow, 0, len(driverSet))
	for drv := range driverSet {
		row := qualiRow{driverNumber: drv}
		for i := 0; i < 3; i++ {
			if t, ok := qTimes[i][drv]; ok {
				tt := t
				row.q[i] = &tt
			}
		}
		rows = append(rows, row)
	}

	sort.SliceStable(rows, func(i, j int) bool {
		if c := compareOptionalDuration(rows[i].q[2], rows[j].q[2]); c != 0 {
			return c < 0
		}
		if c := compareOptionalDuration(rows[i].q[1], rows[j].q[1]); c != 0 {
			return c < 0
		}
		if c := compareOptionalDuration(rows[i].q[0], rows[j].q[0]); c != 0 {
			return c < 0
		}
		return rows[i].driverNumber < rows[j].driverNumber
	})

	positionByDriver := make(map[string]int, len(rows))
	qByDriver := make(map[string][3]*time.Duration, len(rows))
	for i, r := range rows {
		positionByDriver[r.driverNumber] = i + 1
		qByDriver[r.driverNumber] = r.q
	}

	out := make([]jolpica.SessionResults, len(results))
	copy(out, results)
	for i := range out {
		q, ok := qByDriver[out[i].DriverNumber]
		if !ok {
			continue
		}
		p := positionByDriver[out[i].DriverNumber]
		out[i].Position = &p
		out[i].Q1, out[i].Q2, out[i].Q3 = q[0], q[1], q[2]
	}
	sortResultsByPosition(out)
	return out
}

func anyPositionKnown(results []jolpica.SessionResults) bool {
	for _, r := range results {
		if r.Position != nil {
			return true
		}
	}
	return false
}

func qualifyingSplitTimes(statusEvents []feeds.SessionStatusEvent) []time.Duration {
	if len(statusEvents) == 0 {
		return nil
	}

	var splits []time.Duration
	suspended := false
	for _, e := range statusEvents {
		switch e.Status {
		case "Started":
			if !suspended {
				splits = append(splits, e.Timestamp)
			} else {
				suspended = false
			}
		case "Aborted":
			suspended = true
		case "Finished":
			suspended = false
		}
	}
	splits = append(splits, statusEvents[len(statusEvents)-1].Timestamp)
	return splits
}

func splitQualifyingSessions(laps map[string][]feeds.LapData, splitTimes []time.Duration) [3]map[string][]feeds.LapData {
	var buckets [3]map[string][]feeds.LapData
	for i := range buckets {
		buckets[i] = make(map[string][]feeds.LapData)
	}

	numWindows := len(splitTimes) - 1
	if numWindows > 3 {
		numWindows = 3
	}

	for drv, ls := range laps {
		prevEarly := make([]bool, len(ls))
		for i := 0; i < numWindows; i++ {
			start, end := splitTimes[i], splitTimes[i+1]
			isEarly := make([]bool, len(ls))

			for idx, l := range ls {
				if l.LapStartTime == nil {
					continue
				}
				cond := (*l.LapStartTime > start && *l.LapStartTime < end) || prevEarly[idx]
				early := *l.LapStartTime < end && l.Timestamp > end && l.PitOutTime != nil
				if early {
					isEarly[idx] = true
				}
				if cond && !early {
					buckets[i][drv] = append(buckets[i][drv], l)
				}
			}

			prevEarly = isEarly
		}
	}

	return buckets
}

func filterAccurateLaps(laps map[string][]feeds.LapData) map[string][]feeds.LapData {
	out := make(map[string][]feeds.LapData, len(laps))
	for drv, ls := range laps {
		var kept []feeds.LapData
		for _, l := range ls {
			if l.IsAccurate {
				kept = append(kept, l)
			}
		}
		if len(kept) > 0 {
			out[drv] = kept
		}
	}
	return out
}

func pickQuicklaps(bucket map[string][]feeds.LapData) map[string][]feeds.LapData {
	var fastest time.Duration
	has := false
	for _, ls := range bucket {
		for _, l := range ls {
			if l.LapTime == nil || l.Deleted {
				continue
			}
			if !has || *l.LapTime < fastest {
				fastest = *l.LapTime
				has = true
			}
		}
	}
	if !has {
		return nil
	}
	threshold := time.Duration(float64(fastest) * quicklapThreshold)

	out := make(map[string][]feeds.LapData, len(bucket))
	for drv, ls := range bucket {
		var kept []feeds.LapData
		for _, l := range ls {
			if l.LapTime == nil || l.Deleted || *l.LapTime > threshold {
				continue
			}
			kept = append(kept, l)
		}
		if len(kept) > 0 {
			out[drv] = kept
		}
	}
	return out
}

func minLapTimePerDriver(bucket map[string][]feeds.LapData) map[string]time.Duration {
	out := make(map[string]time.Duration, len(bucket))
	for drv, ls := range bucket {
		var min time.Duration
		has := false
		for _, l := range ls {
			if l.LapTime == nil {
				continue
			}
			if !has || *l.LapTime < min {
				min = *l.LapTime
				has = true
			}
		}
		if has {
			out[drv] = min
		}
	}
	return out
}

func compareOptionalDuration(a, b *time.Duration) int {
	switch {
	case a == nil && b == nil:
		return 0
	case a == nil:
		return 1
	case b == nil:
		return -1
	case *a < *b:
		return -1
	case *a > *b:
		return 1
	default:
		return 0
	}
}
