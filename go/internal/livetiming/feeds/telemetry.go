package feeds

import (
	"encoding/json"
	"fmt"
	"sort"
	"time"

	"github.com/KohanMathers/KmF1/go/internal/livetiming/wire"
)

type Telemetry struct {
	DriverNumber string
	RPM          *int
	Speed        *int
	Gear         *int
	Throttle     *int
	Braking      *bool
	DRS          string
	DRSNo        *int
	PositionX    *int
	PositionY    *int
	PositionZ    *int
	Timestamp    time.Duration
}

type carSample struct {
	driverNumber string
	date         time.Time
	frameElapsed time.Duration
	rpm          *int
	speed        *int
	gear         *int
	throttle     *int
	braking      *bool
	drsNo        *int
}

type positionSample struct {
	driverNumber string
	date         time.Time
	frameElapsed time.Duration
	x, y, z      *int
}

func ParseTelemetry(carData, positionData []byte) (map[string][]Telemetry, error) {
	var d TelemetryDecoder

	carRecords, err := wire.SplitFrames(carData)
	if err != nil {
		return nil, fmt.Errorf("feeds: CarData: %w", err)
	}
	for i, rec := range carRecords {
		if err := d.FeedCarData(rec.Elapsed, rec.Payload); err != nil {
			return nil, fmt.Errorf("feeds: CarData: record %d: %w", i, err)
		}
	}

	posRecords, err := wire.SplitFrames(positionData)
	if err != nil {
		return nil, fmt.Errorf("feeds: Position: %w", err)
	}
	for i, rec := range posRecords {
		if err := d.FeedPosition(rec.Elapsed, rec.Payload); err != nil {
			return nil, fmt.Errorf("feeds: Position: record %d: %w", i, err)
		}
	}

	return d.Snapshot(), nil
}

type TelemetryDecoder struct {
	cars      map[string][]carSample
	positions map[string][]positionSample
}

func (d *TelemetryDecoder) FeedCarData(elapsed time.Duration, payload string) error {
	decoded, err := wire.Decode(payload, true)
	if err != nil {
		return err
	}

	var raw struct {
		Entries []struct {
			Utc  string `json:"Utc"`
			Cars map[string]struct {
				Channels map[string]float64 `json:"Channels"`
			} `json:"Cars"`
		} `json:"Entries"`
	}
	if err := json.Unmarshal(decoded, &raw); err != nil {
		return fmt.Errorf("unmarshal: %w", err)
	}

	if d.cars == nil {
		d.cars = make(map[string][]carSample)
	}
	for _, entry := range raw.Entries {
		date, err := time.Parse(time.RFC3339Nano, entry.Utc)
		if err != nil {
			continue
		}
		for drv, car := range entry.Cars {
			s := carSample{driverNumber: drv, date: date, frameElapsed: elapsed}
			if v, ok := car.Channels["0"]; ok {
				n := int(v)
				s.rpm = &n
			}
			if v, ok := car.Channels["2"]; ok {
				n := int(v)
				s.speed = &n
			}
			if v, ok := car.Channels["3"]; ok {
				n := int(v)
				s.gear = &n
			}
			if v, ok := car.Channels["4"]; ok {
				n := int(v)
				s.throttle = &n
			}
			if v, ok := car.Channels["5"]; ok {
				b := v != 0
				s.braking = &b
			}
			if v, ok := car.Channels["45"]; ok {
				n := int(v)
				s.drsNo = &n
			}
			d.cars[drv] = append(d.cars[drv], s)
		}
	}
	return nil
}

func (d *TelemetryDecoder) FeedPosition(elapsed time.Duration, payload string) error {
	decoded, err := wire.Decode(payload, true)
	if err != nil {
		return err
	}

	var raw struct {
		Position []struct {
			Timestamp string `json:"Timestamp"`
			Entries   map[string]struct {
				Status string  `json:"Status"`
				X      float64 `json:"X"`
				Y      float64 `json:"Y"`
				Z      float64 `json:"Z"`
			} `json:"Entries"`
		} `json:"Position"`
	}
	if err := json.Unmarshal(decoded, &raw); err != nil {
		return fmt.Errorf("unmarshal: %w", err)
	}

	if d.positions == nil {
		d.positions = make(map[string][]positionSample)
	}
	for _, entry := range raw.Position {
		date, err := time.Parse(time.RFC3339Nano, entry.Timestamp)
		if err != nil {
			continue
		}
		for drv, pos := range entry.Entries {
			x, y, z := int(pos.X), int(pos.Y), int(pos.Z)
			d.positions[drv] = append(d.positions[drv], positionSample{
				driverNumber: drv,
				date:         date,
				frameElapsed: elapsed,
				x:            &x,
				y:            &y,
				z:            &z,
			})
		}
	}
	return nil
}

func parseCarData(data []byte) (map[string][]carSample, error) {
	records, err := wire.SplitFrames(data)
	if err != nil {
		return nil, fmt.Errorf("feeds: CarData: %w", err)
	}
	var d TelemetryDecoder
	for i, rec := range records {
		if err := d.FeedCarData(rec.Elapsed, rec.Payload); err != nil {
			return nil, fmt.Errorf("feeds: CarData: record %d: %w", i, err)
		}
	}
	return d.cars, nil
}

func parsePositionData(data []byte) (map[string][]positionSample, error) {
	records, err := wire.SplitFrames(data)
	if err != nil {
		return nil, fmt.Errorf("feeds: Position: %w", err)
	}
	var d TelemetryDecoder
	for i, rec := range records {
		if err := d.FeedPosition(rec.Elapsed, rec.Payload); err != nil {
			return nil, fmt.Errorf("feeds: Position: record %d: %w", i, err)
		}
	}
	return d.positions, nil
}

func (d *TelemetryDecoder) Snapshot() map[string][]Telemetry {
	t0 := computeT0(d.cars, d.positions)

	drivers := make(map[string]bool)
	for drv := range d.cars {
		drivers[drv] = true
	}
	for drv := range d.positions {
		drivers[drv] = true
	}

	result := make(map[string][]Telemetry, len(drivers))
	for drv := range drivers {
		result[drv] = mergeDriverTelemetry(drv, d.cars[drv], d.positions[drv], t0)
	}
	return result
}

func computeT0(cars map[string][]carSample, positions map[string][]positionSample) time.Time {
	var t0 time.Time
	has := false
	consider := func(date time.Time, frameElapsed time.Duration) {
		candidate := date.Add(-frameElapsed)
		if !has || candidate.After(t0) {
			t0 = candidate
			has = true
		}
	}

	for _, samples := range cars {
		for _, s := range samples {
			consider(s.date, s.frameElapsed)
		}
	}
	for _, samples := range positions {
		for _, s := range samples {
			consider(s.date, s.frameElapsed)
		}
	}

	return t0
}

func mergeDriverTelemetry(drv string, cars []carSample, positions []positionSample, t0 time.Time) []Telemetry {
	type event struct {
		date time.Time
		car  *carSample
		pos  *positionSample
	}

	events := make([]event, 0, len(cars)+len(positions))
	for i := range cars {
		events = append(events, event{date: cars[i].date, car: &cars[i]})
	}
	for i := range positions {
		events = append(events, event{date: positions[i].date, pos: &positions[i]})
	}
	sort.Slice(events, func(i, j int) bool { return events[i].date.Before(events[j].date) })

	out := make([]Telemetry, 0, len(events))
	var lastCar *carSample
	var lastPos *positionSample

	for _, e := range events {
		if e.car != nil {
			lastCar = e.car
		}
		if e.pos != nil {
			lastPos = e.pos
		}

		t := Telemetry{DriverNumber: drv, Timestamp: e.date.Sub(t0)}
		if lastCar != nil {
			t.RPM = lastCar.rpm
			t.Speed = lastCar.speed
			t.Gear = lastCar.gear
			t.Throttle = lastCar.throttle
			t.Braking = lastCar.braking
			t.DRSNo = lastCar.drsNo
			if lastCar.drsNo != nil {
				t.DRS = mapDRS(*lastCar.drsNo)
			}
		}
		if lastPos != nil {
			t.PositionX = lastPos.x
			t.PositionY = lastPos.y
			t.PositionZ = lastPos.z
		}
		out = append(out, t)
	}

	return out
}

func mapDRS(raw int) string {
	switch raw {
	case 0, 1:
		return "DISABLED"
	case 2, 3:
		return "UNKNOWN"
	case 8:
		return "ELIGIBLE"
	case 10, 12, 14:
		return "ENABLED"
	default:
		return "UNKNOWN"
	}
}
