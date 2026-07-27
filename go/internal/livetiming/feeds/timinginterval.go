package feeds

import (
	"fmt"
	"strconv"
	"time"

	"github.com/KohanMathers/KmF1/go/internal/livetiming/wire"
)

type TimingInterval struct {
	DriverNumber  string
	Position      int
	GapToLeader   string
	GapToCarAhead string
	Timestamp     time.Duration
}

func ParseTimingIntervals(data []byte) (map[string][]TimingInterval, error) {
	records, err := wire.SplitFrames(data)
	if err != nil {
		return nil, fmt.Errorf("feeds: TimingInterval: %w", err)
	}

	var d TimingDataDecoder
	for i, rec := range records {
		if err := d.Feed(rec.Elapsed, rec.Payload); err != nil {
			return nil, fmt.Errorf("feeds: TimingInterval: record %d: %w", i, err)
		}
	}
	return d.Intervals(), nil
}

func timingIntervalsForDriver(driverNumber string, patches []TimingPatch) []TimingInterval {
	var out []TimingInterval
	var position int
	var gapToLeader, gapToCarAhead string

	for _, patch := range patches {
		d := patch.Data
		changed := false

		if v, ok := d["Position"].(string); ok {
			if n, err := strconv.Atoi(v); err == nil && n != position {
				position = n
				changed = true
			}
		}
		if v, ok := d["GapToLeader"].(string); ok && v != gapToLeader {
			gapToLeader = v
			changed = true
		}
		if ipa, ok := d["IntervalToPositionAhead"].(map[string]any); ok {
			if v, ok := ipa["Value"].(string); ok && v != gapToCarAhead {
				gapToCarAhead = v
				changed = true
			}
		}

		if changed {
			out = append(out, TimingInterval{
				DriverNumber:  driverNumber,
				Position:      position,
				GapToLeader:   gapToLeader,
				GapToCarAhead: gapToCarAhead,
				Timestamp:     patch.Elapsed,
			})
		}
	}

	return out
}
