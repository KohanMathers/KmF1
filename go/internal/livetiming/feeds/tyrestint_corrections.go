package feeds

import (
	"sort"
	"time"
)

func FixTyreStints(
	stints map[string][]TyreStint,
	laps map[string][]LapData,
	sessionStartTime time.Duration,
	haveSessionStartTime bool,
	raceLike bool,
) []string {
	var fixed []string

	for drv, drvStints := range stints {
		if len(drvStints) == 0 {
			continue
		}

		splitTimes := stintSplitTimes(laps[drv])
		brackets := stintBrackets(splitTimes)

		if raceLike && len(splitTimes) > 0 && haveSessionStartTime &&
			!anyBefore(drvStints, sessionStartTime) {
			fixBunchedMessages(drvStints, brackets)
		}

		if fixStintCounter(drvStints, brackets) {
			fixed = append(fixed, drv)
		}
	}

	sort.Strings(fixed)
	return fixed
}

func stintSplitTimes(drvLaps []LapData) []time.Duration {
	var times []time.Duration
	for _, l := range drvLaps {
		if l.PitInTime != nil {
			times = append(times, *l.PitInTime)
		}
	}
	return times
}

func stintBrackets(splitTimes []time.Duration) []time.Duration {
	brackets := make([]time.Duration, 0, len(splitTimes)+2)
	brackets = append(brackets, 0)
	brackets = append(brackets, splitTimes...)
	brackets = append(brackets, 24*time.Hour)
	return brackets
}

func anyBefore(drvStints []TyreStint, t time.Duration) bool {
	for _, s := range drvStints {
		if s.Timestamp < t {
			return true
		}
	}
	return false
}

func fixBunchedMessages(drvStints []TyreStint, brackets []time.Duration) {
	firstTs := drvStints[0].Timestamp
	for _, s := range drvStints {
		if s.Timestamp < firstTs {
			firstTs = s.Timestamp
		}
	}

	distinctStints := make(map[int]bool)
	for _, s := range drvStints {
		if s.Timestamp == firstTs {
			distinctStints[s.Stint] = true
		}
	}
	if len(distinctStints) <= 1 {
		return
	}

	for i := range drvStints {
		s := &drvStints[i]
		if s.Timestamp != firstTs {
			continue
		}
		rawStint := s.Stint - 1
		if rawStint < 0 || rawStint >= len(brackets) {
			continue
		}
		s.Timestamp = brackets[rawStint] + time.Millisecond
	}
}

func fixStintCounter(drvStints []TyreStint, brackets []time.Duration) bool {
	fixedAny := false
	stint := 0

	for i := 0; i < len(brackets)-1; i++ {
		tStart, tEnd := brackets[i], brackets[i+1]

		matchedExpected := false
		for j := range drvStints {
			s := &drvStints[j]
			if s.Timestamp < tStart || s.Timestamp > tEnd {
				continue
			}
			rawStint := s.Stint - 1
			if rawStint > stint {
				s.Stint = stint + 1
				fixedAny = true
			} else if rawStint == stint {
				matchedExpected = true
			}
		}

		if matchedExpected {
			stint++
		}
	}

	return fixedAny
}
