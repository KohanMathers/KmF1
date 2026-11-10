package jolpica

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
)

type DriverStanding struct {
	Season         int
	Round          int
	Position       int
	PositionText   string
	Points         float64
	Wins           int
	DriverId       string
	Abbreviation   string
	ConstructorIds []string
}

type ConstructorStanding struct {
	Season        int
	Round         int
	Position      int
	PositionText  string
	Points        float64
	Wins          int
	ConstructorId string
}

type rawDriverStandingEntry struct {
	Position     string `json:"position"`
	PositionText string `json:"positionText"`
	Points       string `json:"points"`
	Wins         string `json:"wins"`
	Driver       struct {
		DriverId string `json:"driverId"`
		Code     string `json:"code"`
	} `json:"Driver"`
	Constructors []struct {
		ConstructorId string `json:"constructorId"`
	} `json:"Constructors"`
}

type rawConstructorStandingEntry struct {
	Position     string `json:"position"`
	PositionText string `json:"positionText"`
	Points       string `json:"points"`
	Wins         string `json:"wins"`
	Constructor  struct {
		ConstructorId string `json:"constructorId"`
	} `json:"Constructor"`
}

func ParseDriverStandings(data []byte) ([]DriverStanding, error) {
	var env struct {
		MRData struct {
			StandingsTable struct {
				StandingsLists []struct {
					Season          string                   `json:"season"`
					Round           string                   `json:"round"`
					DriverStandings []rawDriverStandingEntry `json:"DriverStandings"`
				} `json:"StandingsLists"`
			} `json:"StandingsTable"`
		} `json:"MRData"`
	}
	if err := json.Unmarshal(data, &env); err != nil {
		return nil, fmt.Errorf("jolpica: DriverStanding: unmarshal: %w", err)
	}
	if len(env.MRData.StandingsTable.StandingsLists) == 0 {
		return nil, nil
	}

	list := env.MRData.StandingsTable.StandingsLists[0]
	season, _ := strconv.Atoi(list.Season)
	round, _ := strconv.Atoi(list.Round)

	standings := make([]DriverStanding, 0, len(list.DriverStandings))
	for _, e := range list.DriverStandings {
		pos, _ := strconv.Atoi(e.Position)
		points, _ := strconv.ParseFloat(e.Points, 64)
		wins, _ := strconv.Atoi(e.Wins)

		constructorIds := make([]string, 0, len(e.Constructors))
		for _, c := range e.Constructors {
			constructorIds = append(constructorIds, c.ConstructorId)
		}

		standings = append(standings, DriverStanding{
			Season:         season,
			Round:          round,
			Position:       pos,
			PositionText:   e.PositionText,
			Points:         points,
			Wins:           wins,
			DriverId:       e.Driver.DriverId,
			Abbreviation:   e.Driver.Code,
			ConstructorIds: constructorIds,
		})
	}
	return standings, nil
}

func ParseConstructorStandings(data []byte) ([]ConstructorStanding, error) {
	var env struct {
		MRData struct {
			StandingsTable struct {
				StandingsLists []struct {
					Season               string                        `json:"season"`
					Round                string                        `json:"round"`
					ConstructorStandings []rawConstructorStandingEntry `json:"ConstructorStandings"`
				} `json:"StandingsLists"`
			} `json:"StandingsTable"`
		} `json:"MRData"`
	}
	if err := json.Unmarshal(data, &env); err != nil {
		return nil, fmt.Errorf("jolpica: ConstructorStanding: unmarshal: %w", err)
	}
	if len(env.MRData.StandingsTable.StandingsLists) == 0 {
		return nil, nil
	}

	list := env.MRData.StandingsTable.StandingsLists[0]
	season, _ := strconv.Atoi(list.Season)
	round, _ := strconv.Atoi(list.Round)

	standings := make([]ConstructorStanding, 0, len(list.ConstructorStandings))
	for _, e := range list.ConstructorStandings {
		pos, _ := strconv.Atoi(e.Position)
		points, _ := strconv.ParseFloat(e.Points, 64)
		wins, _ := strconv.Atoi(e.Wins)

		standings = append(standings, ConstructorStanding{
			Season:        season,
			Round:         round,
			Position:      pos,
			PositionText:  e.PositionText,
			Points:        points,
			Wins:          wins,
			ConstructorId: e.Constructor.ConstructorId,
		})
	}
	return standings, nil
}

func (c *Client) FetchDriverStandings(ctx context.Context, year, round int) ([]DriverStanding, error) {
	rows, err := fetchList(ctx, c, fmt.Sprintf("/%d/%d/driverStandings.json", year, round), ParseDriverStandings)
	if err != nil {
		return nil, fmt.Errorf("jolpica: fetch driver standings %d/%d: %w", year, round, err)
	}
	return rows, nil
}

func (c *Client) FetchConstructorStandings(ctx context.Context, year, round int) ([]ConstructorStanding, error) {
	rows, err := fetchList(ctx, c, fmt.Sprintf("/%d/%d/constructorStandings.json", year, round), ParseConstructorStandings)
	if err != nil {
		return nil, fmt.Errorf("jolpica: fetch constructor standings %d/%d: %w", year, round, err)
	}
	return rows, nil
}
