package kmf1

import "github.com/KohanMathers/KmF1/go/internal/livetiming/feeds"

// Lap is a single lap's timing data, embedding feeds.LapData, obtained via Driver.Lap.
type Lap struct {
	feeds.LapData
	driver *Driver
}

// Telemetry returns the telemetry samples recorded during this lap, from the lap's start time up to its recorded timestamp.
func (l *Lap) Telemetry() []feeds.Telemetry {
	if l == nil || l.driver == nil || l.LapStartTime == nil {
		return nil
	}

	all := l.driver.session.telemetry[l.driver.driverNumber]
	var out []feeds.Telemetry
	for _, t := range all {
		if t.Timestamp >= *l.LapStartTime && t.Timestamp <= l.Timestamp {
			out = append(out, t)
		}
	}
	return out
}
