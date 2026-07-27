package kmf1

import "github.com/KohanMathers/KmF1/go/internal/livetiming/feeds"

// ChampionshipPrediction holds the live-timing championship win-probability prediction for a race session, obtained via Session.ChampionshipPrediction or LiveSession.ChampionshipPrediction.
type ChampionshipPrediction struct {
	drivers map[string][]feeds.ChampionshipPredictionDriver
	teams   map[string][]feeds.ChampionshipPredictionTeam
}

// Drivers returns the latest championship prediction for each driver.
func (c *ChampionshipPrediction) Drivers() []feeds.ChampionshipPredictionDriver {
	out := make([]feeds.ChampionshipPredictionDriver, 0, len(c.drivers))
	for _, snapshots := range c.drivers {
		if len(snapshots) > 0 {
			out = append(out, snapshots[len(snapshots)-1])
		}
	}
	return out
}

// Teams returns the latest championship prediction for each constructor.
func (c *ChampionshipPrediction) Teams() []feeds.ChampionshipPredictionTeam {
	out := make([]feeds.ChampionshipPredictionTeam, 0, len(c.teams))
	for _, snapshots := range c.teams {
		if len(snapshots) > 0 {
			out = append(out, snapshots[len(snapshots)-1])
		}
	}
	return out
}
