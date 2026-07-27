package jolpica

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"
)

type SessionResults struct {
	DriverNumber string
	DriverId     string
	Abbreviation string
	TeamName     string

	Position     *int
	PositionText string
	GridPosition *int

	Q1 *time.Duration
	Q2 *time.Duration
	Q3 *time.Duration

	Laps   *int
	Status string
	Points float64

	Time *time.Duration

	FastestLapRank     *int
	FastestLapNumber   *int
	FastestLapTime     *time.Duration
	FastestLapAvgSpeed *float64
}

type rawResultEntry struct {
	Number       string `json:"number"`
	Position     string `json:"position"`
	PositionText string `json:"positionText"`
	Points       string `json:"points"`
	Driver       struct {
		DriverId string `json:"driverId"`
		Code     string `json:"code"`
	} `json:"Driver"`
	Constructor struct {
		Name string `json:"name"`
	} `json:"Constructor"`
	Grid   string `json:"grid"`
	Laps   string `json:"laps"`
	Status string `json:"status"`
	Time   *struct {
		Time string `json:"time"`
	} `json:"Time"`
	FastestLap *struct {
		Rank string `json:"rank"`
		Lap  string `json:"lap"`
		Time struct {
			Time string `json:"time"`
		} `json:"Time"`
		AverageSpeed *struct {
			Speed string `json:"speed"`
		} `json:"AverageSpeed"`
	} `json:"FastestLap"`
}

type rawQualifyingEntry struct {
	Number string `json:"number"`
	Driver struct {
		DriverId string `json:"driverId"`
		Code     string `json:"code"`
	} `json:"Driver"`
	Constructor struct {
		Name string `json:"name"`
	} `json:"Constructor"`
	Position string `json:"position"`
	Q1       string `json:"Q1"`
	Q2       string `json:"Q2"`
	Q3       string `json:"Q3"`
}

func ParseRaceResults(data []byte) ([]SessionResults, error) {
	var env struct {
		MRData struct {
			RaceTable struct {
				Races []struct {
					Results []rawResultEntry `json:"Results"`
				} `json:"Races"`
			} `json:"RaceTable"`
		} `json:"MRData"`
	}
	if err := json.Unmarshal(data, &env); err != nil {
		return nil, fmt.Errorf("jolpica: RaceResults: unmarshal: %w", err)
	}
	if len(env.MRData.RaceTable.Races) == 0 {
		return nil, nil
	}
	return resultEntriesToSessionResults(env.MRData.RaceTable.Races[0].Results), nil
}

func ParseSprintResults(data []byte) ([]SessionResults, error) {
	var env struct {
		MRData struct {
			RaceTable struct {
				Races []struct {
					SprintResults []rawResultEntry `json:"SprintResults"`
				} `json:"Races"`
			} `json:"RaceTable"`
		} `json:"MRData"`
	}
	if err := json.Unmarshal(data, &env); err != nil {
		return nil, fmt.Errorf("jolpica: SprintResults: unmarshal: %w", err)
	}
	if len(env.MRData.RaceTable.Races) == 0 {
		return nil, nil
	}
	return resultEntriesToSessionResults(env.MRData.RaceTable.Races[0].SprintResults), nil
}

func ParseQualifyingResults(data []byte) ([]SessionResults, error) {
	var env struct {
		MRData struct {
			RaceTable struct {
				Races []struct {
					QualifyingResults []rawQualifyingEntry `json:"QualifyingResults"`
				} `json:"Races"`
			} `json:"RaceTable"`
		} `json:"MRData"`
	}
	if err := json.Unmarshal(data, &env); err != nil {
		return nil, fmt.Errorf("jolpica: QualifyingResults: unmarshal: %w", err)
	}
	if len(env.MRData.RaceTable.Races) == 0 {
		return nil, nil
	}

	entries := env.MRData.RaceTable.Races[0].QualifyingResults
	results := make([]SessionResults, 0, len(entries))
	for _, e := range entries {
		var pos *int
		if n, err := strconv.Atoi(e.Position); err == nil {
			pos = &n
		}
		results = append(results, SessionResults{
			DriverNumber: e.Number,
			DriverId:     e.Driver.DriverId,
			Abbreviation: e.Driver.Code,
			TeamName:     e.Constructor.Name,
			Position:     pos,
			PositionText: e.Position,
			Q1:           parseErgastDurationPtr(e.Q1),
			Q2:           parseErgastDurationPtr(e.Q2),
			Q3:           parseErgastDurationPtr(e.Q3),
		})
	}
	return results, nil
}

func resultEntriesToSessionResults(entries []rawResultEntry) []SessionResults {
	results := make([]SessionResults, 0, len(entries))
	for _, e := range entries {
		var pos *int
		if n, err := strconv.Atoi(e.Position); err == nil {
			pos = &n
		}
		points, _ := strconv.ParseFloat(e.Points, 64)

		var grid, laps *int
		if n, err := strconv.Atoi(e.Grid); err == nil {
			grid = &n
		}
		if n, err := strconv.Atoi(e.Laps); err == nil {
			laps = &n
		}

		var raceTime *time.Duration
		if e.Time != nil {
			raceTime = parseErgastDurationPtr(e.Time.Time)
		}

		var fastestLapRank, fastestLapNumber *int
		var fastestLapTime *time.Duration
		var fastestLapAvgSpeed *float64
		if e.FastestLap != nil {
			if n, err := strconv.Atoi(e.FastestLap.Rank); err == nil {
				fastestLapRank = &n
			}
			if n, err := strconv.Atoi(e.FastestLap.Lap); err == nil {
				fastestLapNumber = &n
			}
			fastestLapTime = parseErgastDurationPtr(e.FastestLap.Time.Time)
			if e.FastestLap.AverageSpeed != nil {
				if f, err := strconv.ParseFloat(e.FastestLap.AverageSpeed.Speed, 64); err == nil {
					fastestLapAvgSpeed = &f
				}
			}
		}

		results = append(results, SessionResults{
			DriverNumber:       e.Number,
			DriverId:           e.Driver.DriverId,
			Abbreviation:       e.Driver.Code,
			TeamName:           e.Constructor.Name,
			Position:           pos,
			PositionText:       e.PositionText,
			GridPosition:       grid,
			Laps:               laps,
			Status:             e.Status,
			Points:             points,
			Time:               raceTime,
			FastestLapRank:     fastestLapRank,
			FastestLapNumber:   fastestLapNumber,
			FastestLapTime:     fastestLapTime,
			FastestLapAvgSpeed: fastestLapAvgSpeed,
		})
	}
	return results
}

func parseErgastDurationPtr(s string) *time.Duration {
	s = strings.TrimPrefix(s, "+")
	if s == "" {
		return nil
	}
	d, err := parseErgastDuration(s)
	if err != nil {
		return nil
	}
	return &d
}

func parseErgastDuration(s string) (time.Duration, error) {
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

func (c *Client) FetchRaceResults(ctx context.Context, year, round int) ([]SessionResults, error) {
	body, err := c.fetch(ctx, fmt.Sprintf("/%d/%d/results.json", year, round))
	if err != nil {
		return nil, fmt.Errorf("jolpica: fetch race results %d/%d: %w", year, round, err)
	}
	return ParseRaceResults(body)
}

func (c *Client) FetchSprintResults(ctx context.Context, year, round int) ([]SessionResults, error) {
	body, err := c.fetch(ctx, fmt.Sprintf("/%d/%d/sprint.json", year, round))
	if err != nil {
		return nil, fmt.Errorf("jolpica: fetch sprint results %d/%d: %w", year, round, err)
	}
	return ParseSprintResults(body)
}

func (c *Client) FetchQualifyingResults(ctx context.Context, year, round int) ([]SessionResults, error) {
	body, err := c.fetch(ctx, fmt.Sprintf("/%d/%d/qualifying.json", year, round))
	if err != nil {
		return nil, fmt.Errorf("jolpica: fetch qualifying results %d/%d: %w", year, round, err)
	}
	return ParseQualifyingResults(body)
}
