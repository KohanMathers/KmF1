package kmf1

import (
	"context"
	"log/slog"
	"sort"
	"sync"

	"github.com/KohanMathers/KmF1/go/internal/livetiming/feeds"
	"github.com/KohanMathers/KmF1/go/internal/livetiming/live"
)

// LiveSession accumulates decoded state from a live SignalR timing feed as messages are fed into it via Consume. Create one with NewLiveSession, then read its accumulated state at any time from another goroutine using its accessor methods (Info, Drivers, Laps, etc.), which are safe for concurrent use.
type LiveSession struct {
	mu     sync.RWMutex
	anchor live.TimeAnchor
	logger *slog.Logger

	sessionInfo   feeds.SessionInfoDecoder
	driverList    feeds.DriverListDecoder
	trackStatus   feeds.TrackStatusDecoder
	sessionStatus feeds.SessionStatusDecoder
	lapCount      feeds.SessionProgressDecoder
	timingData    feeds.TimingDataDecoder
	tyreStints    feeds.TyreStintDecoder
	telemetry     feeds.TelemetryDecoder
	weather       feeds.WeatherSampleDecoder
	rcm           feeds.RaceControlMessageDecoder
	teamRadio     feeds.TeamRadioDecoder
	topThree      feeds.TopThreeLiveDecoder
	bestStats     feeds.DriverBestStatsDecoder
	clock         feeds.SessionClockDecoder
	championship  feeds.ChampionshipPredictionDecoder
	audio         feeds.AudioStreamDecoder
	content       feeds.ContentStreamDecoder
}

// LiveSessionOption configures a LiveSession created by NewLiveSession.
type LiveSessionOption func(*LiveSession)

// WithLiveSessionLogger sets the logger a LiveSession uses to report malformed feed messages. If unset, log messages are discarded.
func WithLiveSessionLogger(logger *slog.Logger) LiveSessionOption {
	return func(s *LiveSession) { s.logger = logger }
}

// NewLiveSession creates an empty LiveSession ready to Consume messages from a live feed.
func NewLiveSession(opts ...LiveSessionOption) *LiveSession {
	s := &LiveSession{}
	for _, opt := range opts {
		opt(s)
	}
	if s.logger == nil {
		s.logger = slog.New(slog.DiscardHandler)
	}
	return s
}

// Consume reads messages from in, feeding each into the session's accumulated state, until ctx is cancelled or in is closed. It is intended to be run in its own goroutine alongside Client.Live.
func (s *LiveSession) Consume(ctx context.Context, in <-chan live.Message) {
	for {
		select {
		case <-ctx.Done():
			return
		case msg, ok := <-in:
			if !ok {
				return
			}
			s.feed(msg)
		}
	}
}

func (s *LiveSession) feed(msg live.Message) {
	elapsed := s.anchor.Elapsed(msg.Timestamp)

	s.mu.Lock()
	defer s.mu.Unlock()

	var err error
	switch msg.Topic {
	case "SessionInfo":
		err = s.sessionInfo.Feed(elapsed, msg.Payload)
	case "DriverList":
		err = s.driverList.Feed(elapsed, msg.Payload)
	case "TrackStatus":
		err = s.trackStatus.Feed(elapsed, msg.Payload)
	case "SessionStatus":
		err = s.sessionStatus.Feed(elapsed, msg.Payload)
	case "LapCount":
		err = s.lapCount.Feed(elapsed, msg.Payload)
	case "TimingData":
		err = s.timingData.Feed(elapsed, msg.Payload)
	case "TimingAppData":
		err = s.tyreStints.Feed(elapsed, msg.Payload)
	case "CarData.z":
		err = s.telemetry.FeedCarData(elapsed, msg.Payload)
	case "Position.z":
		err = s.telemetry.FeedPosition(elapsed, msg.Payload)
	case "WeatherData":
		err = s.weather.Feed(elapsed, msg.Payload)
	case "RaceControlMessages":
		err = s.rcm.Feed(elapsed, msg.Payload)
	case "TeamRadio":
		err = s.teamRadio.Feed(elapsed, msg.Payload)
	case "TopThree":
		err = s.topThree.Feed(elapsed, msg.Payload)
	case "TimingStats":
		err = s.bestStats.Feed(elapsed, msg.Payload)
	case "ExtrapolatedClock":
		err = s.clock.Feed(elapsed, msg.Payload)
	case "ChampionshipPrediction":
		err = s.championship.Feed(elapsed, msg.Payload)
	case "AudioStreams":
		err = s.audio.Feed(elapsed, msg.Payload)
	case "ContentStreams":
		err = s.content.Feed(elapsed, msg.Payload)
	default:
		return
	}

	if err != nil {
		s.logger.Warn("kmf1: live feed message failed, dropping it", "topic", msg.Topic, "error", err)
	}
}

// Info returns the session's metadata as decoded so far. It reports false if no SessionInfo message has been received yet.
func (s *LiveSession) Info() (*feeds.SessionInfo, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.sessionInfo.Snapshot()
}

// Drivers returns the current driver list as decoded so far, sorted by driver number.
func (s *LiveSession) Drivers() []feeds.DriverInfo {
	s.mu.RLock()
	defer s.mu.RUnlock()
	drivers, err := s.driverList.Snapshot()
	if err != nil {
		return nil
	}
	out := make([]feeds.DriverInfo, 0, len(drivers))
	for _, d := range drivers {
		out = append(out, *d)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].DriverNumber < out[j].DriverNumber })
	return out
}

// TrackStatus returns every track status change received so far.
func (s *LiveSession) TrackStatus() []feeds.TrackStatus {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.trackStatus.Snapshot()
}

// StatusEvents returns every session status change received so far.
func (s *LiveSession) StatusEvents() []feeds.SessionStatusEvent {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.sessionStatus.Snapshot()
}

// Progress returns every lap/time progress update received so far.
func (s *LiveSession) Progress() []feeds.SessionProgress {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.lapCount.Snapshot()
}

// Laps returns every lap decoded so far, across all drivers, in no particular order.
func (s *LiveSession) Laps() []feeds.LapData {
	s.mu.RLock()
	defer s.mu.RUnlock()
	laps := s.timingData.Laps()
	out := make([]feeds.LapData, 0)
	for _, ls := range laps {
		out = append(out, ls...)
	}
	return out
}

// Intervals returns every timing interval decoded so far, across all drivers, in no particular order.
func (s *LiveSession) Intervals() []feeds.TimingInterval {
	s.mu.RLock()
	defer s.mu.RUnlock()
	intervals := s.timingData.Intervals()
	out := make([]feeds.TimingInterval, 0)
	for _, is := range intervals {
		out = append(out, is...)
	}
	return out
}

// Stints returns every tyre stint decoded so far, across all drivers, in no particular order.
func (s *LiveSession) Stints() []feeds.TyreStint {
	s.mu.RLock()
	defer s.mu.RUnlock()
	stints := s.tyreStints.Snapshot()
	out := make([]feeds.TyreStint, 0)
	for _, ss := range stints {
		out = append(out, ss...)
	}
	return out
}

// Telemetry returns every telemetry sample decoded so far, across all drivers, in no particular order.
func (s *LiveSession) Telemetry() []feeds.Telemetry {
	s.mu.RLock()
	defer s.mu.RUnlock()
	telemetry := s.telemetry.Snapshot()
	out := make([]feeds.Telemetry, 0)
	for _, ts := range telemetry {
		out = append(out, ts...)
	}
	return out
}

// Weather returns every weather sample received so far.
func (s *LiveSession) Weather() []feeds.WeatherSample {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.weather.Snapshot()
}

// RaceControlMessages returns every race control message received so far.
func (s *LiveSession) RaceControlMessages() []feeds.RaceControlMessage {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.rcm.Snapshot()
}

// TeamRadio returns every team radio clip received so far.
func (s *LiveSession) TeamRadio() []feeds.TeamRadio {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.teamRadio.Snapshot()
}

// TopThree returns the latest snapshot of the top three drivers as shown on the live-timing broadcast overlay.
func (s *LiveSession) TopThree() []feeds.TopThreeLive {
	s.mu.RLock()
	defer s.mu.RUnlock()
	slots := s.topThree.Snapshot()
	out := make([]feeds.TopThreeLive, 0, 3)
	for slot := 0; slot < 3; slot++ {
		snapshots := slots[slot]
		if len(snapshots) > 0 {
			out = append(out, snapshots[len(snapshots)-1])
		}
	}
	return out
}

// BestStats returns each driver's best lap and sector times decoded so far, sorted by driver number.
func (s *LiveSession) BestStats() []feeds.DriverBestStats {
	s.mu.RLock()
	defer s.mu.RUnlock()
	stats := s.bestStats.Snapshot()
	out := make([]feeds.DriverBestStats, 0, len(stats))
	for _, snapshots := range stats {
		if len(snapshots) > 0 {
			out = append(out, snapshots[len(snapshots)-1])
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].DriverNumber < out[j].DriverNumber })
	return out
}

// Clock returns every extrapolated session clock sample received so far.
func (s *LiveSession) Clock() []feeds.SessionClock {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.clock.Snapshot()
}

// ChampionshipPrediction returns the latest championship win-probability prediction decoded so far.
func (s *LiveSession) ChampionshipPrediction() *ChampionshipPrediction {
	s.mu.RLock()
	defer s.mu.RUnlock()
	drivers, teams := s.championship.Snapshot()
	return &ChampionshipPrediction{drivers: drivers, teams: teams}
}

// Broadcast returns the audio and content streams advertised so far.
func (s *LiveSession) Broadcast() *Broadcast {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return &Broadcast{audio: s.audio.Snapshot(), content: s.content.Snapshot()}
}
