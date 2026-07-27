package feeds

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/KohanMathers/KmF1/go/internal/livetiming/wire"
)

type DriverBestStats struct {
	DriverNumber string

	BestLapTime     *time.Duration
	BestLapNumber   *int
	BestLapPosition *int

	BestSector1Time     *time.Duration
	BestSector1Position *int
	BestSector2Time     *time.Duration
	BestSector2Position *int
	BestSector3Time     *time.Duration
	BestSector3Position *int

	BestSpeedI1         *int
	BestSpeedI1Position *int
	BestSpeedI2         *int
	BestSpeedI2Position *int
	BestSpeedFL         *int
	BestSpeedFLPosition *int
	BestSpeedST         *int
	BestSpeedSTPosition *int

	Timestamp time.Duration
}

type driverBestStatsBuilder struct {
	DriverBestStats
}

func ParseDriverBestStats(data []byte) (map[string][]DriverBestStats, error) {
	records, err := wire.SplitFrames(data)
	if err != nil {
		return nil, fmt.Errorf("feeds: TimingStats: %w", err)
	}

	var d DriverBestStatsDecoder
	for i, rec := range records {
		if err := d.Feed(rec.Elapsed, rec.Payload); err != nil {
			return nil, fmt.Errorf("feeds: TimingStats: record %d: %w", i, err)
		}
	}
	return d.Snapshot(), nil
}

type DriverBestStatsDecoder struct {
	builders map[string]*driverBestStatsBuilder
	result   map[string][]DriverBestStats
}

func (d *DriverBestStatsDecoder) Feed(elapsed time.Duration, payload string) error {
	decoded, err := wire.Decode(payload, false)
	if err != nil {
		return err
	}

	var top struct {
		Lines map[string]map[string]any `json:"Lines"`
	}
	if err := json.Unmarshal(decoded, &top); err != nil {
		return fmt.Errorf("unmarshal: %w", err)
	}

	if d.builders == nil {
		d.builders = make(map[string]*driverBestStatsBuilder)
		d.result = make(map[string][]DriverBestStats)
	}
	for drv, patch := range top.Lines {
		b, ok := d.builders[drv]
		if !ok {
			b = &driverBestStatsBuilder{DriverBestStats: DriverBestStats{DriverNumber: drv}}
			d.builders[drv] = b
		}
		if applyBestStatsPatch(patch, b) {
			snapshot := b.DriverBestStats
			snapshot.Timestamp = elapsed
			d.result[drv] = append(d.result[drv], snapshot)
		}
	}
	return nil
}

func (d *DriverBestStatsDecoder) Snapshot() map[string][]DriverBestStats {
	out := make(map[string][]DriverBestStats, len(d.result))
	for drv, snapshots := range d.result {
		cp := make([]DriverBestStats, len(snapshots))
		copy(cp, snapshots)
		out[drv] = cp
	}
	return out
}

func applyBestStatsPatch(patch map[string]any, b *driverBestStatsBuilder) bool {
	changed := false

	if blt, ok := patch["PersonalBestLapTime"].(map[string]any); ok {
		if applyBestValue(blt, &b.BestLapTime, &b.BestLapPosition) {
			changed = true
		}
		if lap, ok := blt["Lap"]; ok {
			n := int(toFloat(lap))
			b.BestLapNumber = &n
			changed = true
		}
	}

	if sectors := indexedEntries(patch["BestSectors"]); len(sectors) > 0 {
		targets := []struct {
			val **time.Duration
			pos **int
		}{
			{&b.BestSector1Time, &b.BestSector1Position},
			{&b.BestSector2Time, &b.BestSector2Position},
			{&b.BestSector3Time, &b.BestSector3Position},
		}
		for idx, fields := range sectors {
			if idx < 0 || idx >= len(targets) {
				continue
			}
			if applyBestValue(fields, targets[idx].val, targets[idx].pos) {
				changed = true
			}
		}
	}

	if speeds, ok := patch["BestSpeeds"].(map[string]any); ok {
		targets := map[string]struct {
			val **int
			pos **int
		}{
			"I1": {&b.BestSpeedI1, &b.BestSpeedI1Position},
			"I2": {&b.BestSpeedI2, &b.BestSpeedI2Position},
			"FL": {&b.BestSpeedFL, &b.BestSpeedFLPosition},
			"ST": {&b.BestSpeedST, &b.BestSpeedSTPosition},
		}
		for key, t := range targets {
			fields, ok := speeds[key].(map[string]any)
			if !ok {
				continue
			}
			if applyBestSpeed(fields, t.val, t.pos) {
				changed = true
			}
		}
	}

	return changed
}

func applyBestValue(fields map[string]any, val **time.Duration, pos **int) bool {
	changed := false
	if v, ok := fields["Value"].(string); ok && v != "" {
		if dur, err := parseF1Duration(v); err == nil {
			*val = &dur
			changed = true
		}
	}
	if p, ok := fields["Position"]; ok {
		n := int(toFloat(p))
		*pos = &n
		changed = true
	}
	return changed
}

func applyBestSpeed(fields map[string]any, val **int, pos **int) bool {
	changed := false
	if v, ok := fields["Value"].(string); ok && v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			*val = &n
			changed = true
		}
	}
	if p, ok := fields["Position"]; ok {
		n := int(toFloat(p))
		*pos = &n
		changed = true
	}
	return changed
}
