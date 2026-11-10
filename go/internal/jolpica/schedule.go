package jolpica

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"
)

type Schedule struct {
	Season            int
	RoundNumber       int
	EventName         string
	OfficialEventName string
	Country           string
	Location          string
	EventFormat       string

	CircuitKey int

	Sessions []ScheduleSession
}

type ScheduleSession struct {
	SessionType  string
	StartDate    time.Time
	StartDateUtc time.Time

	Path           string
	PathCandidates []string `json:",omitempty"`
}

type rawScheduleEnvelope struct {
	MRData struct {
		RaceTable struct {
			Season string            `json:"season"`
			Races  []rawScheduleRace `json:"Races"`
		} `json:"RaceTable"`
	} `json:"MRData"`
}

type rawScheduleRace struct {
	Round    string `json:"round"`
	RaceName string `json:"raceName"`
	Circuit  struct {
		Location struct {
			Locality string `json:"locality"`
			Country  string `json:"country"`
		} `json:"Location"`
	} `json:"Circuit"`
	Date             string           `json:"date"`
	Time             string           `json:"time"`
	FirstPractice    *rawSessionBlock `json:"FirstPractice"`
	SecondPractice   *rawSessionBlock `json:"SecondPractice"`
	ThirdPractice    *rawSessionBlock `json:"ThirdPractice"`
	SprintQualifying *rawSessionBlock `json:"SprintQualifying"`
	Sprint           *rawSessionBlock `json:"Sprint"`
	Qualifying       *rawSessionBlock `json:"Qualifying"`
}

type rawSessionBlock struct {
	Date string `json:"date"`
	Time string `json:"time"`
}

func ParseSchedule(data []byte) ([]Schedule, error) {
	var env rawScheduleEnvelope
	if err := json.Unmarshal(data, &env); err != nil {
		return nil, fmt.Errorf("jolpica: Schedule: unmarshal: %w", err)
	}

	season, _ := strconv.Atoi(env.MRData.RaceTable.Season)

	schedules := make([]Schedule, 0, len(env.MRData.RaceTable.Races))
	for _, r := range env.MRData.RaceTable.Races {
		round, _ := strconv.Atoi(r.Round)

		s := Schedule{
			Season:      season,
			RoundNumber: round,
			EventName:   r.RaceName,
			Country:     r.Circuit.Location.Country,
			Location:    r.Circuit.Location.Locality,
		}

		type namedBlock struct {
			sessionType string
			block       *rawSessionBlock
		}
		blocks := []namedBlock{
			{"Practice 1", r.FirstPractice},
			{"Practice 2", r.SecondPractice},
			{"Practice 3", r.ThirdPractice},
			{"Sprint Qualifying", r.SprintQualifying},
			{"Sprint", r.Sprint},
			{"Qualifying", r.Qualifying},
		}
		for _, b := range blocks {
			if b.block == nil {
				continue
			}
			utc, err := parseScheduleDateTime(b.block.Date, b.block.Time)
			if err != nil {
				continue
			}
			s.Sessions = append(s.Sessions, ScheduleSession{SessionType: b.sessionType, StartDateUtc: utc})
		}
		if utc, err := parseScheduleDateTime(r.Date, r.Time); err == nil {
			s.Sessions = append(s.Sessions, ScheduleSession{SessionType: "Race", StartDateUtc: utc})
		}

		schedules = append(schedules, s)
	}

	return schedules, nil
}

func parseScheduleDateTime(date, t string) (time.Time, error) {
	if date == "" || t == "" {
		return time.Time{}, fmt.Errorf("jolpica: missing date/time")
	}
	return time.Parse(time.RFC3339, date+"T"+t)
}

func (c *Client) FetchSchedule(ctx context.Context, year int) ([]Schedule, error) {
	rows, err := fetchList(ctx, c, fmt.Sprintf("/%d.json", year), ParseSchedule)
	if err != nil {
		return nil, fmt.Errorf("jolpica: fetch schedule %d: %w", year, err)
	}
	return rows, nil
}
