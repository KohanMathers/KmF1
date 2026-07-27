package feeds

import (
	"regexp"
	"strings"
)

var (
	lapDeletionRe    = regexp.MustCompile(`CAR (\d{1,2}) .* TIME (\d:\d\d\.\d\d\d) DELETED - (.*)`)
	lapReinstatedRe  = regexp.MustCompile(`CAR (\d{1,2}) .* TIME (\d:\d\d\.\d\d\d) .*REINSTATED.*`)
	localTrackTimeRe = regexp.MustCompile(`\d\d:\d\d:\d\d`)
)

func MarkDeletedLaps(laps map[string][]LapData, messages []RaceControlMessage) {
	if len(messages) == 0 {
		return
	}

	type key struct{ driver, lapTime string }

	reinstated := make(map[key]bool)
	for _, m := range messages {
		match := lapReinstatedRe.FindStringSubmatch(m.Message)
		if match == nil {
			continue
		}
		reinstated[key{match[1], match[2]}] = true
	}

	for _, m := range messages {
		match := lapDeletionRe.FindStringSubmatch(m.Message)
		if match == nil {
			continue
		}
		drv, lapTimeStr, reason := match[1], match[2], match[3]
		if reinstated[key{drv, lapTimeStr}] {
			continue
		}

		deletedTime, err := parseF1Duration(lapTimeStr)
		if err != nil {
			continue
		}
		reason = strings.TrimSpace(localTrackTimeRe.ReplaceAllString(reason, ""))

		for i := range laps[drv] {
			l := &laps[drv][i]
			if l.LapTime != nil && *l.LapTime == deletedTime {
				l.Deleted = true
				l.IsPersonalBest = false
				l.DeletedReason = reason
			}
		}
	}
}
