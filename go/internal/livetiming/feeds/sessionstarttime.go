package feeds

import "time"

func SessionStartTime(events []SessionStatusEvent) (time.Duration, bool) {
	for _, e := range events {
		if e.Status == "Started" {
			return e.Timestamp, true
		}
	}
	return 0, false
}
