package feeds

import (
	"sort"
	"strings"
	"time"
)

func AlignLaps(laps map[string][]LapData, intervals map[string][]TimingInterval) []string {
	if !anyGapData(intervals) {
		return nil
	}

	maxLaps := 0
	for _, ls := range laps {
		if len(ls) > maxLaps {
			maxLaps = len(ls)
		}
	}

	expectedGap := make(map[string]time.Duration)
	delta := make(map[string]time.Duration)

	for offset := 0; offset < maxLaps; offset++ {
		activeDrivers, skip := activeAndSkipDrivers(laps, offset)

		leader := ""
		hasLeader := false

		for _, drv := range activeDrivers {
			if skip[drv] {
				continue
			}
			gapStr, ok := nearestGap(intervals[drv], laps[drv][offset].Timestamp)
			if !ok {
				continue
			}
			if strings.Contains(gapStr, "LAP") {
				leader = drv
				hasLeader = true
				continue
			}
			if _, already := delta[drv]; !already {
				if eg, err := parseGapString(gapStr); err == nil {
					expectedGap[drv] = eg
				}
			}
		}

		if !hasLeader {
			continue
		}

		leaderTime := laps[leader][offset].Timestamp
		if _, ok := delta[leader]; !ok {
			delta[leader] = 0
		}
		refZero := delta[leader]

		for drv, eg := range expectedGap {
			if _, already := delta[drv]; already {
				continue
			}
			ls, ok := laps[drv]
			if !ok || offset >= len(ls) {
				continue
			}
			otherTime := ls[offset].Timestamp
			isGap := otherTime - leaderTime
			delta[drv] = eg - isGap + refZero
		}

		if len(activeDrivers) <= len(delta) {
			break
		}
	}

	normalizeAndApplyDeltas(laps, delta)

	var unaligned []string
	for drv := range laps {
		if _, ok := delta[drv]; !ok {
			unaligned = append(unaligned, drv)
		}
	}
	sort.Strings(unaligned)
	return unaligned
}

func anyGapData(intervals map[string][]TimingInterval) bool {
	for _, ivs := range intervals {
		for _, iv := range ivs {
			if iv.GapToLeader != "" {
				return true
			}
		}
	}
	return false
}

func activeAndSkipDrivers(laps map[string][]LapData, offset int) ([]string, map[string]bool) {
	var active []string
	skip := make(map[string]bool)
	for drv, ls := range laps {
		if offset >= len(ls) {
			continue
		}
		active = append(active, drv)
		lap := ls[offset]
		if lap.PitInTime != nil || lap.PitOutTime != nil {
			skip[drv] = true
		}
	}
	sort.Strings(active)
	return active, skip
}

func nearestGap(intervals []TimingInterval, target time.Duration) (string, bool) {
	if len(intervals) == 0 {
		return "", false
	}
	best := intervals[0]
	bestDiff := absDuration(target - best.Timestamp)
	for _, iv := range intervals[1:] {
		if diff := absDuration(target - iv.Timestamp); diff < bestDiff {
			best, bestDiff = iv, diff
		}
	}
	return best.GapToLeader, true
}

func parseGapString(s string) (time.Duration, error) {
	return parseF1Duration(strings.TrimPrefix(s, "+"))
}

func absDuration(d time.Duration) time.Duration {
	if d < 0 {
		return -d
	}
	return d
}

func normalizeAndApplyDeltas(laps map[string][]LapData, delta map[string]time.Duration) {
	if len(delta) == 0 {
		return
	}

	var maxDelta time.Duration
	first := true
	for _, d := range delta {
		if first || d > maxDelta {
			maxDelta = d
			first = false
		}
	}
	if maxDelta < 0 {
		maxDelta = 0
	}

	for drv, d := range delta {
		d -= maxDelta
		ls := laps[drv]
		for i := range ls {
			ls[i].Timestamp += d
		}
	}
}
