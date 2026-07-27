package feeds

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/KohanMathers/KmF1/go/internal/livetiming/wire"
)

type ChampionshipPredictionDriver struct {
	DriverNumber      string
	CurrentPosition   int
	CurrentPoints     float64
	PredictedPosition int
	PredictedPoints   float64
	Timestamp         time.Duration
}

type ChampionshipPredictionTeam struct {
	TeamId            string
	TeamName          string
	CurrentPosition   int
	CurrentPoints     float64
	PredictedPosition int
	PredictedPoints   float64
	Timestamp         time.Duration
}

type championshipPredictionRaw struct {
	Drivers map[string]map[string]any `json:"Drivers"`
	Teams   map[string]map[string]any `json:"Teams"`
}

func ParseChampionshipPrediction(data []byte) (drivers map[string][]ChampionshipPredictionDriver, teams map[string][]ChampionshipPredictionTeam, err error) {
	records, err := wire.SplitFrames(data)
	if err != nil {
		return nil, nil, fmt.Errorf("feeds: ChampionshipPrediction: %w", err)
	}

	var d ChampionshipPredictionDecoder
	for i, rec := range records {
		if err := d.Feed(rec.Elapsed, rec.Payload); err != nil {
			return nil, nil, fmt.Errorf("feeds: ChampionshipPrediction: record %d: %w", i, err)
		}
	}
	drivers, teams = d.Snapshot()
	return drivers, teams, nil
}

type ChampionshipPredictionDecoder struct {
	driverBuilders map[string]*ChampionshipPredictionDriver
	teamBuilders   map[string]*ChampionshipPredictionTeam
	drivers        map[string][]ChampionshipPredictionDriver
	teams          map[string][]ChampionshipPredictionTeam
}

func (d *ChampionshipPredictionDecoder) Feed(elapsed time.Duration, payload string) error {
	decoded, err := wire.Decode(payload, false)
	if err != nil {
		return err
	}

	var raw championshipPredictionRaw
	if err := json.Unmarshal(decoded, &raw); err != nil {
		return fmt.Errorf("unmarshal: %w", err)
	}

	if d.driverBuilders == nil {
		d.driverBuilders = make(map[string]*ChampionshipPredictionDriver)
		d.teamBuilders = make(map[string]*ChampionshipPredictionTeam)
		d.drivers = make(map[string][]ChampionshipPredictionDriver)
		d.teams = make(map[string][]ChampionshipPredictionTeam)
	}

	for drv, patch := range raw.Drivers {
		b, ok := d.driverBuilders[drv]
		if !ok {
			b = &ChampionshipPredictionDriver{DriverNumber: drv}
			d.driverBuilders[drv] = b
		}
		if applyChampionshipPredictionFields(patch, &b.CurrentPosition, &b.CurrentPoints, &b.PredictedPosition, &b.PredictedPoints) {
			snapshot := *b
			snapshot.Timestamp = elapsed
			d.drivers[drv] = append(d.drivers[drv], snapshot)
		}
	}

	for team, patch := range raw.Teams {
		b, ok := d.teamBuilders[team]
		if !ok {
			b = &ChampionshipPredictionTeam{TeamId: team}
			d.teamBuilders[team] = b
		}
		nameChanged := false
		if v, ok := patch["TeamName"].(string); ok {
			b.TeamName = v
			nameChanged = true
		}
		fieldsChanged := applyChampionshipPredictionFields(patch, &b.CurrentPosition, &b.CurrentPoints, &b.PredictedPosition, &b.PredictedPoints)
		if nameChanged || fieldsChanged {
			snapshot := *b
			snapshot.Timestamp = elapsed
			d.teams[team] = append(d.teams[team], snapshot)
		}
	}
	return nil
}

func (d *ChampionshipPredictionDecoder) Snapshot() (drivers map[string][]ChampionshipPredictionDriver, teams map[string][]ChampionshipPredictionTeam) {
	drivers = make(map[string][]ChampionshipPredictionDriver, len(d.drivers))
	for drv, snapshots := range d.drivers {
		cp := make([]ChampionshipPredictionDriver, len(snapshots))
		copy(cp, snapshots)
		drivers[drv] = cp
	}
	teams = make(map[string][]ChampionshipPredictionTeam, len(d.teams))
	for team, snapshots := range d.teams {
		cp := make([]ChampionshipPredictionTeam, len(snapshots))
		copy(cp, snapshots)
		teams[team] = cp
	}
	return drivers, teams
}

func applyChampionshipPredictionFields(patch map[string]any, curPos *int, curPts *float64, predPos *int, predPts *float64) bool {
	changed := false
	if v, ok := patch["CurrentPosition"]; ok {
		*curPos = int(toFloat(v))
		changed = true
	}
	if v, ok := patch["CurrentPoints"]; ok {
		*curPts = toFloat(v)
		changed = true
	}
	if v, ok := patch["PredictedPosition"]; ok {
		*predPos = int(toFloat(v))
		changed = true
	}
	if v, ok := patch["PredictedPoints"]; ok {
		*predPts = toFloat(v)
		changed = true
	}
	return changed
}
