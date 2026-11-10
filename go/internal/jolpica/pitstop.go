package jolpica

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"
)

type PitStop struct {
	DriverId   string
	StopNumber int
	Lap        int
	Duration   time.Duration
	Time       time.Duration
}

type rawPitStop struct {
	DriverId string `json:"driverId"`
	Lap      string `json:"lap"`
	Stop     string `json:"stop"`
	Time     string `json:"time"`
	Duration string `json:"duration"`
}

func ParsePitStops(data []byte) ([]PitStop, error) {
	var env struct {
		MRData struct {
			RaceTable struct {
				Races []struct {
					PitStops []rawPitStop `json:"PitStops"`
				} `json:"Races"`
			} `json:"RaceTable"`
		} `json:"MRData"`
	}
	if err := json.Unmarshal(data, &env); err != nil {
		return nil, fmt.Errorf("jolpica: PitStop: unmarshal: %w", err)
	}
	if len(env.MRData.RaceTable.Races) == 0 {
		return nil, nil
	}

	entries := env.MRData.RaceTable.Races[0].PitStops
	stops := make([]PitStop, 0, len(entries))
	for _, e := range entries {
		lap, _ := strconv.Atoi(e.Lap)
		stop, _ := strconv.Atoi(e.Stop)
		duration, _ := parseSeconds(e.Duration)
		timeOfDay, _ := parseTimeOfDay(e.Time)

		stops = append(stops, PitStop{
			DriverId:   e.DriverId,
			StopNumber: stop,
			Lap:        lap,
			Duration:   duration,
			Time:       timeOfDay,
		})
	}
	return stops, nil
}

func parseSeconds(s string) (time.Duration, error) {
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0, err
	}
	return time.Duration(f * float64(time.Second)), nil
}

func parseTimeOfDay(s string) (time.Duration, error) {
	parts := strings.Split(s, ":")
	if len(parts) != 3 {
		return 0, fmt.Errorf("unrecognized time-of-day format %q", s)
	}
	h, err := strconv.Atoi(parts[0])
	if err != nil {
		return 0, err
	}
	m, err := strconv.Atoi(parts[1])
	if err != nil {
		return 0, err
	}
	sec, err := strconv.Atoi(parts[2])
	if err != nil {
		return 0, err
	}
	return time.Duration(h)*time.Hour + time.Duration(m)*time.Minute + time.Duration(sec)*time.Second, nil
}

func (c *Client) FetchPitStops(ctx context.Context, year, round int) ([]PitStop, error) {
	rows, err := fetchList(ctx, c, fmt.Sprintf("/%d/%d/pitstops.json", year, round), ParsePitStops)
	if err != nil {
		return nil, fmt.Errorf("jolpica: fetch pit stops %d/%d: %w", year, round, err)
	}
	return rows, nil
}
