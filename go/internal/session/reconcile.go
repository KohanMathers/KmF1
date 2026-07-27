package session

import (
	"sort"

	"github.com/KohanMathers/KmF1/go/internal/jolpica"
	"github.com/KohanMathers/KmF1/go/internal/livetiming/feeds"
)

func ReconcileResults(
	driverInfo map[string]*feeds.DriverInfo,
	jolpicaResults []jolpica.SessionResults,
	ergastListAccurate bool,
) []jolpica.SessionResults {
	switch {
	case len(driverInfo) == 0 && len(jolpicaResults) == 0:
		return nil
	case len(driverInfo) == 0:
		out := append([]jolpica.SessionResults(nil), jolpicaResults...)
		sortResultsByPosition(out)
		return out
	case len(jolpicaResults) == 0:
		out := make([]jolpica.SessionResults, 0, len(driverInfo))
		for drv, di := range driverInfo {
			out = append(out, jolpica.SessionResults{
				DriverNumber: drv,
				Abbreviation: di.Abbreviation,
				TeamName:     di.TeamName,
			})
		}
		sortResultsByPosition(out)
		return out
	}

	jolpicaByDriver := make(map[string]jolpica.SessionResults, len(jolpicaResults))
	for _, r := range jolpicaResults {
		jolpicaByDriver[r.DriverNumber] = r
	}

	seen := make(map[string]bool, len(driverInfo))
	out := make([]jolpica.SessionResults, 0, len(driverInfo))
	for drv, di := range driverInfo {
		seen[drv] = true
		out = append(out, mergeResultRow(drv, di, jolpicaByDriver[drv]))
	}

	if ergastListAccurate {
		for drv, jr := range jolpicaByDriver {
			if seen[drv] {
				continue
			}
			row := jr
			row.Position = nil
			row.GridPosition = nil
			out = append(out, row)
		}
	}

	sortResultsByPosition(out)
	return out
}

func mergeResultRow(driverNumber string, di *feeds.DriverInfo, jr jolpica.SessionResults) jolpica.SessionResults {
	row := jolpica.SessionResults{
		DriverNumber:       driverNumber,
		DriverId:           jr.DriverId,
		Abbreviation:       di.Abbreviation,
		TeamName:           di.TeamName,
		Position:           jr.Position,
		PositionText:       jr.PositionText,
		GridPosition:       jr.GridPosition,
		Q1:                 jr.Q1,
		Q2:                 jr.Q2,
		Q3:                 jr.Q3,
		Laps:               jr.Laps,
		Status:             jr.Status,
		Points:             jr.Points,
		Time:               jr.Time,
		FastestLapRank:     jr.FastestLapRank,
		FastestLapNumber:   jr.FastestLapNumber,
		FastestLapTime:     jr.FastestLapTime,
		FastestLapAvgSpeed: jr.FastestLapAvgSpeed,
	}
	if row.Abbreviation == "" {
		row.Abbreviation = jr.Abbreviation
	}
	if row.TeamName == "" {
		row.TeamName = jr.TeamName
	}
	return row
}

func sortResultsByPosition(results []jolpica.SessionResults) {
	sort.SliceStable(results, func(i, j int) bool {
		pi, pj := results[i].Position, results[j].Position
		switch {
		case pi == nil && pj == nil:
			return results[i].DriverNumber < results[j].DriverNumber
		case pi == nil:
			return false
		case pj == nil:
			return true
		default:
			return *pi < *pj
		}
	})
}
