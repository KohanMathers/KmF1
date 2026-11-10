package jolpica

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"
)

type LapTiming struct {
	LapNumber int
	DriverId  string
	Time      time.Duration
	Position  int
}

type rawLapTiming struct {
	DriverId string `json:"driverId"`
	Time     string `json:"time"`
	Position string `json:"position"`
}

type rawLap struct {
	Number  string         `json:"number"`
	Timings []rawLapTiming `json:"Timings"`
}

func ParseLapTimes(data []byte) ([]LapTiming, error) {
	var env struct {
		MRData struct {
			RaceTable struct {
				Races []struct {
					Laps []rawLap `json:"Laps"`
				} `json:"Races"`
			} `json:"RaceTable"`
		} `json:"MRData"`
	}
	if err := json.Unmarshal(data, &env); err != nil {
		return nil, fmt.Errorf("jolpica: LapTiming: unmarshal: %w", err)
	}
	if len(env.MRData.RaceTable.Races) == 0 {
		return nil, nil
	}

	var timings []LapTiming
	for _, lap := range env.MRData.RaceTable.Races[0].Laps {
		lapNumber, _ := strconv.Atoi(lap.Number)
		for _, t := range lap.Timings {
			pos, _ := strconv.Atoi(t.Position)
			timings = append(timings, LapTiming{
				LapNumber: lapNumber,
				DriverId:  t.DriverId,
				Time:      parseErgastDuration1(t.Time),
				Position:  pos,
			})
		}
	}
	return timings, nil
}

func parseErgastDuration1(s string) time.Duration {
	d, err := parseErgastDuration(s)
	if err != nil {
		return 0
	}
	return d
}

func (c *Client) FetchLapTimes(ctx context.Context, year, round, lapNumber int) ([]LapTiming, error) {
	rows, err := fetchList(ctx, c, fmt.Sprintf("/%d/%d/laps/%d.json", year, round, lapNumber), ParseLapTimes)
	if err != nil {
		return nil, fmt.Errorf("jolpica: fetch lap times %d/%d lap %d: %w", year, round, lapNumber, err)
	}
	return rows, nil
}
