package feeds

import (
	"strings"
	"time"
)

func AddTrackStatusToLaps(laps map[string][]LapData, statuses []TrackStatus) {
	if len(statuses) == 0 {
		return
	}
	for _, ls := range laps {
		for i := range ls {
			AddTrackStatusToLap(&ls[i], statuses)
		}
	}
}

func AddTrackStatusToLap(l *LapData, statuses []TrackStatus) {
	if len(statuses) == 0 {
		return
	}

	for i := 0; i < len(statuses)-1; i++ {
		t, next := statuses[i].Timestamp, statuses[i+1].Timestamp
		if lapOverlapsStatusWindow(l, t, next) {
			appendTrackStatusCode(l, statuses[i].Status)
		}
	}

	last := statuses[len(statuses)-1]
	if l.Timestamp >= last.Timestamp {
		appendTrackStatusCode(l, last.Status)
	}
}

func lapOverlapsStatusWindow(l *LapData, t, next time.Duration) bool {
	if t <= l.Timestamp && l.Timestamp <= next {
		return true
	}
	if l.LapStartTime != nil {
		if t <= *l.LapStartTime && *l.LapStartTime <= next {
			return true
		}
		if *l.LapStartTime <= t && next <= l.Timestamp {
			return true
		}
	}
	return false
}

func appendTrackStatusCode(l *LapData, status string) {
	if status == "" || strings.Contains(l.TrackStatus, status) {
		return
	}
	l.TrackStatus += status
}
