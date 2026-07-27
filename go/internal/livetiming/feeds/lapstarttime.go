package feeds

import "time"

func DeriveLapStartTimes(
	laps map[string][]LapData,
	sessionStartTime time.Duration,
	haveSessionStartTime bool,
	statusEvents []SessionStatusEvent,
	raceLike bool,
) {
	for _, ls := range laps {
		for i := range ls {
			if i == 0 {
				if raceLike && haveSessionStartTime {
					t := sessionStartTime
					ls[i].LapStartTime = &t
				}
				continue
			}
			t := ls[i-1].Timestamp
			ls[i].LapStartTime = &t
		}
	}

	applyRedFlagRestarts(laps, statusEvents, raceLike)

	for _, ls := range laps {
		for i := range ls {
			if ls[i].LapStartTime == nil && ls[i].PitOutTime != nil {
				t := *ls[i].PitOutTime
				ls[i].LapStartTime = &t
			}
		}
	}
}

func applyRedFlagRestarts(laps map[string][]LapData, statusEvents []SessionStatusEvent, raceLike bool) {
	isAborted := false
	for _, e := range statusEvents {
		switch {
		case e.Status == "Aborted":
			isAborted = true
		case e.Status == "Started" && isAborted:
			isAborted = false
			restartTime := e.Timestamp

			for _, ls := range laps {
				restartIdx := -1
				for i, l := range ls {
					if l.PitOutTime != nil && *l.PitOutTime > restartTime {
						restartIdx = i
						break
					}
				}
				if restartIdx == -1 {
					continue
				}

				if raceLike {
					if ls[restartIdx].LapTime == nil {
						t := restartTime
						ls[restartIdx].LapStartTime = &t
					}
				} else {
					ls[restartIdx].LapStartTime = nil
				}
			}
		}
	}
}
