package feeds

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/KohanMathers/KmF1/go/internal/livetiming/wire"
)

type LapData struct {
	DriverNumber   string
	LapNo          int
	LapTime        *time.Duration
	Sector1Time    *time.Duration
	Sector2Time    *time.Duration
	Sector3Time    *time.Duration
	SpeedI1        *int
	SpeedI2        *int
	SpeedFL        *int
	SpeedST        *int
	PitStopNo      int
	PitInTime      *time.Duration
	PitOutTime     *time.Duration
	IsPersonalBest bool
	Timestamp      time.Duration

	LapStartTime    *time.Duration
	TrackStatus     string
	Position        *int
	FastF1Generated bool
	IsAccurate      bool
	Deleted         bool
	DeletedReason   string
}

type lapBuilder struct {
	driverNumber string

	time    time.Duration
	timeSet bool

	lapTime             *time.Duration
	lapTimeExplicitNone bool

	sector1, sector2, sector3                      *time.Duration
	sector1Session, sector2Session, sector3Session *time.Duration

	speedI1, speedI2, speedFL, speedST *int

	pitInTime, pitOutTime *time.Duration

	numberOfLaps     int
	numberOfPitStops int
	isPersonalBest   bool
}

type pbEntry struct {
	elapsed time.Duration
	value   time.Duration
}

func ParseLapData(data []byte) (map[string][]LapData, error) {
	records, err := wire.SplitFrames(data)
	if err != nil {
		return nil, fmt.Errorf("feeds: LapData: %w", err)
	}

	var d TimingDataDecoder
	for i, rec := range records {
		if err := d.Feed(rec.Elapsed, rec.Payload); err != nil {
			return nil, fmt.Errorf("feeds: LapData: record %d: %w", i, err)
		}
	}
	return d.Laps(), nil
}

func (b *lapBuilder) finalize() LapData {
	return LapData{
		DriverNumber:   b.driverNumber,
		LapNo:          b.numberOfLaps,
		LapTime:        b.lapTime,
		Sector1Time:    b.sector1,
		Sector2Time:    b.sector2,
		Sector3Time:    b.sector3,
		SpeedI1:        b.speedI1,
		SpeedI2:        b.speedI2,
		SpeedFL:        b.speedFL,
		SpeedST:        b.speedST,
		PitStopNo:      b.numberOfPitStops,
		PitInTime:      b.pitInTime,
		PitOutTime:     b.pitOutTime,
		IsPersonalBest: b.isPersonalBest,
		Timestamp:      b.time,
	}
}

func buildRawLaps(driverNumber string, patches []TimingPatch) ([]*lapBuilder, []pbEntry) {
	var laps []*lapBuilder
	lapcnt := 0
	pitstops := -1
	inPast := false

	ensure := func(idx int) *lapBuilder {
		for len(laps) <= idx {
			b := &lapBuilder{driverNumber: driverNumber}
			laps = append(laps, b)
		}
		return laps[idx]
	}

	var personalBests []pbEntry

	for _, patch := range patches {
		d := patch.Data

		var newLapcnt int
		hasNL := false
		if nlRaw, ok := d["NumberOfLaps"]; ok {
			hasNL = true
			newLapcnt = int(toFloat(nlRaw))
		}

		if hasNL {
			switch {
			case newLapcnt < lapcnt:

				inPast = true
			case inPast:
				if newLapcnt == lapcnt {
					inPast = false
				}
				continue
			}
		}
		if inPast {
			continue
		}

		cur := ensure(lapcnt)

		target := cur
		if lapcnt > 0 {
			prev := laps[lapcnt-1]
			if prev.timeSet && patch.Elapsed >= prev.time && patch.Elapsed-prev.time < 5*time.Second {
				target = prev
			}
		}

		applySectors(d, target, patch.Elapsed)
		applySpeeds(d, target, cur)
		applyLastLapTime(d, target)

		if blt, ok := d["BestLapTime"].(map[string]any); ok {
			if v, _ := blt["Value"].(string); v != "" {
				if dur, err := parseF1Duration(v); err == nil {
					personalBests = append(personalBests, pbEntry{elapsed: patch.Elapsed, value: dur})
				}
			}
		}

		if v, ok := d["InPit"].(bool); ok && v && pitstops >= 0 {
			t := patch.Elapsed
			cur.pitInTime = &t
		}
		if v, ok := d["InPit"].(bool); ok && !v {

			pitOutTarget := cur
			closeToLapEnd := cur.timeSet && patch.Elapsed-cur.time < 5*time.Second
			if hasNL || closeToLapEnd {
				pitOutTarget = ensure(lapcnt + 1)
			}
			t := patch.Elapsed
			pitOutTarget.pitOutTime = &t
			pitstops++
		}

		if hasNL {
			cur.numberOfLaps = lapcnt + 1
			cur.numberOfPitStops = pitstops
			cur.time = patch.Elapsed
			cur.timeSet = true

			if newLapcnt > lapcnt && pitstops >= 0 {
				lapcnt = newLapcnt
			}
		}
	}

	return laps, personalBests
}

func applySectors(d map[string]any, target *lapBuilder, elapsed time.Duration) {
	sectors := sectorsFromPatch(d)
	for idx, entry := range sectors {
		v, _ := entry["Value"].(string)
		if v == "" {
			continue
		}
		dur, err := parseF1Duration(v)
		if err != nil {
			continue
		}
		switch idx {
		case 0:
			target.sector1 = &dur
			target.sector1Session = durPtr(elapsed)
		case 1:
			target.sector2 = &dur
			target.sector2Session = durPtr(elapsed)
		case 2:
			target.sector3 = &dur
			target.sector3Session = durPtr(elapsed)
		}
	}
}

func sectorsFromPatch(d map[string]any) map[int]map[string]any {
	return indexedEntries(d["Sectors"])
}

func indexedEntries(raw any) map[int]map[string]any {
	out := make(map[int]map[string]any)
	switch v := raw.(type) {
	case []any:
		for i, item := range v {
			if m, ok := item.(map[string]any); ok {
				out[i] = m
			}
		}
	case map[string]any:
		for k, item := range v {
			idx, err := strconv.Atoi(k)
			if err != nil {
				continue
			}
			if m, ok := item.(map[string]any); ok {
				out[idx] = m
			}
		}
	}
	return out
}

func applySpeeds(d map[string]any, target, current *lapBuilder) {
	speeds, ok := d["Speeds"].(map[string]any)
	if !ok {
		return
	}

	parse := func(key string) (int, bool) {
		v, ok := speeds[key].(map[string]any)
		if !ok {
			return 0, false
		}
		s, _ := v["Value"].(string)
		if s == "" {
			return 0, false
		}
		n, err := strconv.Atoi(s)
		if err != nil {
			return 0, false
		}
		return n, true
	}

	if n, ok := parse("I1"); ok {
		target.speedI1 = &n
	}
	if n, ok := parse("I2"); ok {
		target.speedI2 = &n
	}
	if n, ok := parse("FL"); ok {
		target.speedFL = &n
	}
	if n, ok := parse("ST"); ok {
		current.speedST = &n
	}
}

func applyLastLapTime(d map[string]any, target *lapBuilder) {
	llt, ok := d["LastLapTime"].(map[string]any)
	if !ok {
		return
	}
	v, present := llt["Value"]
	if !present {
		return
	}
	s, _ := v.(string)
	if s == "" {
		target.lapTimeExplicitNone = true
		target.lapTime = nil
		return
	}
	dur, err := parseF1Duration(s)
	if err != nil {
		return
	}
	if dur >= 150*time.Second {

		return
	}
	target.lapTime = &dur
	target.lapTimeExplicitNone = false
}

func parseF1Duration(s string) (time.Duration, error) {
	parts := strings.Split(s, ":")
	var h, m int
	var secStr string

	switch len(parts) {
	case 1:
		secStr = parts[0]
	case 2:
		mm, err := strconv.Atoi(parts[0])
		if err != nil {
			return 0, fmt.Errorf("minutes in %q: %w", s, err)
		}
		m, secStr = mm, parts[1]
	case 3:
		hh, err := strconv.Atoi(parts[0])
		if err != nil {
			return 0, fmt.Errorf("hours in %q: %w", s, err)
		}
		mm, err := strconv.Atoi(parts[1])
		if err != nil {
			return 0, fmt.Errorf("minutes in %q: %w", s, err)
		}
		h, m, secStr = hh, mm, parts[2]
	default:
		return 0, fmt.Errorf("unrecognized duration format %q", s)
	}

	sec, err := strconv.ParseFloat(secStr, 64)
	if err != nil {
		return 0, fmt.Errorf("seconds in %q: %w", s, err)
	}

	return time.Duration(h)*time.Hour +
		time.Duration(m)*time.Minute +
		time.Duration(sec*float64(time.Second)), nil
}

func durPtr(d time.Duration) *time.Duration { return &d }

func toFloat(v any) float64 {
	switch n := v.(type) {
	case float64:
		return n
	case int:
		return float64(n)
	default:
		return 0
	}
}
