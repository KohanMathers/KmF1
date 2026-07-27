package kmf1

import (
	"sort"
	"strings"

	"github.com/KohanMathers/KmF1/go/internal/jolpica"
	"github.com/KohanMathers/KmF1/go/internal/livetiming/feeds"
)

// Session holds the fully-loaded data for a single practice, qualifying, sprint, or race session, obtained by calling SessionRef.Load.
type Session struct {
	client      *Client
	year, round int
	sessionType string

	info         *feeds.SessionInfo
	drivers      map[string]*feeds.DriverInfo
	trackStatus  []feeds.TrackStatus
	statusEvents []feeds.SessionStatusEvent
	progress     []feeds.SessionProgress

	laps      map[string][]feeds.LapData
	intervals map[string][]feeds.TimingInterval
	stints    map[string][]feeds.TyreStint
	telemetry map[string][]feeds.Telemetry
	results   []jolpica.SessionResults

	bestStats           map[string][]feeds.DriverBestStats
	weather             []feeds.WeatherSample
	raceControlMessages []feeds.RaceControlMessage
	teamRadio           []feeds.TeamRadio
	topThree            map[int][]feeds.TopThreeLive
	clock               []feeds.SessionClock
	archiveStatus       feeds.ArchiveStatus
	audio               []feeds.AudioStream
	content             []feeds.ContentStream
	championshipDrivers map[string][]feeds.ChampionshipPredictionDriver
	championshipTeams   map[string][]feeds.ChampionshipPredictionTeam
	pitStops            []jolpica.PitStop
}

// Info returns the session's metadata (name, type, timing).
func (s *Session) Info() *feeds.SessionInfo { return s.info }

// Drivers returns every driver who participated in the session, sorted by driver number.
func (s *Session) Drivers() []feeds.DriverInfo {
	out := make([]feeds.DriverInfo, 0, len(s.drivers))
	for _, d := range s.drivers {
		out = append(out, *d)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].DriverNumber < out[j].DriverNumber })
	return out
}

// Driver looks up a single driver in the session by driver number or three-letter abbreviation (e.g. "44" or "HAM"). The returned Driver always has Exists false if no match is found.
func (s *Session) Driver(identifier string) *Driver {
	if di, ok := s.drivers[identifier]; ok {
		return &Driver{session: s, driverNumber: di.DriverNumber, exists: true}
	}
	for num, di := range s.drivers {
		if strings.EqualFold(di.Abbreviation, identifier) {
			return &Driver{session: s, driverNumber: num, exists: true}
		}
	}
	return &Driver{session: s, exists: false}
}

// Laps returns every lap recorded during the session, across all drivers, in no particular order.
func (s *Session) Laps() []feeds.LapData {
	out := make([]feeds.LapData, 0)
	for _, ls := range s.laps {
		out = append(out, ls...)
	}
	return out
}

// Results returns the official classification for the session, reconciled between Jolpica's results and the live-timing feed.
func (s *Session) Results() []jolpica.SessionResults { return s.results }

// BestStats returns each driver's best lap and sector times recorded during the session, sorted by driver number.
func (s *Session) BestStats() []feeds.DriverBestStats {
	out := make([]feeds.DriverBestStats, 0, len(s.bestStats))
	for _, snapshots := range s.bestStats {
		if len(snapshots) > 0 {
			out = append(out, snapshots[len(snapshots)-1])
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].DriverNumber < out[j].DriverNumber })
	return out
}

// Weather returns the weather samples recorded throughout the session.
func (s *Session) Weather() []feeds.WeatherSample { return s.weather }

// TrackStatus returns the track status changes (green, yellow, safety car, etc.) recorded throughout the session.
func (s *Session) TrackStatus() []feeds.TrackStatus { return s.trackStatus }

// Progress returns the lap/time progress updates recorded throughout the session.
func (s *Session) Progress() []feeds.SessionProgress { return s.progress }

// StatusEvents returns the session status changes (started, suspended, finalised, etc.) recorded throughout the session.
func (s *Session) StatusEvents() []feeds.SessionStatusEvent { return s.statusEvents }

// Clock returns the extrapolated session clock samples recorded throughout the session.
func (s *Session) Clock() []feeds.SessionClock { return s.clock }

// ArchiveStatus returns the live-timing archive's completeness status for the session.
func (s *Session) ArchiveStatus() feeds.ArchiveStatus { return s.archiveStatus }

// RaceControlMessages returns the race control messages issued during the session.
func (s *Session) RaceControlMessages() []feeds.RaceControlMessage { return s.raceControlMessages }

// TeamRadio returns the team radio clips captured during the session.
func (s *Session) TeamRadio() []feeds.TeamRadio { return s.teamRadio }

// PitStops returns the pit stops made during the session, as reported by Jolpica. It is only populated for race-like sessions.
func (s *Session) PitStops() []jolpica.PitStop { return s.pitStops }

// TopThree returns the latest snapshot of the top three drivers as shown on the live-timing broadcast overlay.
func (s *Session) TopThree() []feeds.TopThreeLive {
	out := make([]feeds.TopThreeLive, 0, 3)
	for slot := 0; slot < 3; slot++ {
		snapshots := s.topThree[slot]
		if len(snapshots) > 0 {
			out = append(out, snapshots[len(snapshots)-1])
		}
	}
	return out
}

// ChampionshipPrediction returns the live championship win-probability prediction for the session. It is only available for race sessions and returns nil otherwise.
func (s *Session) ChampionshipPrediction() *ChampionshipPrediction {
	if s.sessionType != SessionRace {
		return nil
	}
	return &ChampionshipPrediction{drivers: s.championshipDrivers, teams: s.championshipTeams}
}

// Broadcast returns the audio and content streams available for the session.
func (s *Session) Broadcast() *Broadcast {
	return &Broadcast{audio: s.audio, content: s.content}
}
