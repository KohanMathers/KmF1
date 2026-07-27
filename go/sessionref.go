package kmf1

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/KohanMathers/KmF1/go/internal/jolpica"
	"github.com/KohanMathers/KmF1/go/internal/livetiming/feeds"
	"github.com/KohanMathers/KmF1/go/internal/session"
)

// SessionRef identifies a session (practice, qualifying, race, etc.) that has not yet had its data fetched. Obtain one via Event.Session, then call Load to fetch and assemble its full data set.
type SessionRef struct {
	client      *Client
	year        int
	round       int
	sessionType string
	path        string
}

// Load fetches and assembles every available live-timing and Jolpica feed for the referenced session into a Session. Individual feeds that fail to fetch or parse are logged and left empty rather than failing the whole call; only a failure to fetch the session's core SessionInfo feed aborts Load.
func (r *SessionRef) Load(ctx context.Context) (*Session, error) {
	lt := r.client.livetiming
	jp := r.client.jolpica
	logger := r.client.log()
	start := time.Now()

	logger.InfoContext(ctx, "kmf1: loading session", "session", r.path, "year", r.year, "round", r.round, "sessionType", r.sessionType)

	sessionInfoRaw, err := lt.FetchFeed(ctx, r.path, "SessionInfo.jsonStream")
	if err != nil {
		logger.ErrorContext(ctx, "kmf1: load session: fetch SessionInfo failed, aborting Load", "session", r.path, "error", err)
		return nil, fmt.Errorf("kmf1: load session: fetch SessionInfo: %w", err)
	}
	sessionInfo, err := feeds.ParseSessionInfo(sessionInfoRaw)
	if err != nil {
		logger.ErrorContext(ctx, "kmf1: load session: parse SessionInfo failed, aborting Load", "session", r.path, "error", err)
		return nil, fmt.Errorf("kmf1: load session: parse SessionInfo: %w", err)
	}
	logFeedLoaded(ctx, logger, "SessionInfo", r.path)

	s := &Session{
		client:      r.client,
		year:        r.year,
		round:       r.round,
		sessionType: r.sessionType,
		info:        sessionInfo,
	}

	raceLike := isRaceLike(r.sessionType)
	qualiLike := isQualiLike(r.sessionType)

	var (
		driverInfo          map[string]*feeds.DriverInfo
		trackStatuses       []feeds.TrackStatus
		statusEvents        []feeds.SessionStatusEvent
		progress            []feeds.SessionProgress
		laps                map[string][]feeds.LapData
		intervals           map[string][]feeds.TimingInterval
		stints              map[string][]feeds.TyreStint
		telemetry           map[string][]feeds.Telemetry
		raceControlMessages []feeds.RaceControlMessage
		results             []jolpica.SessionResults
		lap1Times           []jolpica.LapTiming
		haveLap1Times       bool
		bestStats           map[string][]feeds.DriverBestStats
		weather             []feeds.WeatherSample
		teamRadio           []feeds.TeamRadio
		topThree            map[int][]feeds.TopThreeLive
		clock               []feeds.SessionClock
		archiveStatus       feeds.ArchiveStatus
		audio               []feeds.AudioStream
		content             []feeds.ContentStream
		championshipDrivers map[string][]feeds.ChampionshipPredictionDriver
		championshipTeams   map[string][]feeds.ChampionshipPredictionTeam
		pitStops            []jolpica.PitStop
	)

	var wg sync.WaitGroup
	run := func(fn func()) {
		wg.Add(1)
		go func() { defer wg.Done(); fn() }()
	}

	run(func() {
		driverInfo, _ = fetchDriverList(ctx, lt, r.path, logger)
		results = fetchAndReconcileResults(ctx, jp, r.year, r.round, r.sessionType, driverInfo, logger)
	})
	run(func() { trackStatuses, _ = fetchTrackStatus(ctx, lt, r.path, logger) })
	run(func() { statusEvents, _ = fetchSessionStatus(ctx, lt, r.path, logger) })
	run(func() { progress, _ = fetchLapCount(ctx, lt, r.path, logger) })
	run(func() { laps, intervals = loadLapsAndIntervals(ctx, lt, r.path, logger) })
	run(func() { stints, _ = fetchTyreStints(ctx, lt, r.path, logger) })
	run(func() { telemetry = loadTelemetry(ctx, lt, r.path, logger) })
	run(func() { raceControlMessages, _ = fetchRaceControlMessages(ctx, lt, r.path, logger) })
	run(func() { bestStats, _ = fetchBestStats(ctx, lt, r.path, logger) })
	run(func() { weather, _ = fetchWeather(ctx, lt, r.path, logger) })
	run(func() { teamRadio, _ = fetchTeamRadio(ctx, lt, r.path, logger) })
	run(func() { topThree, _ = fetchTopThree(ctx, lt, r.path, logger) })
	run(func() { clock, _ = fetchSessionClock(ctx, lt, r.path, logger) })
	run(func() { archiveStatus, _ = fetchArchiveStatus(ctx, lt, r.path, logger) })
	run(func() { audio, _ = fetchAudioStreams(ctx, lt, r.path, logger) })
	run(func() { content, _ = fetchContentStreams(ctx, lt, r.path, logger) })

	if r.sessionType == SessionRace {
		run(func() {
			championshipDrivers, championshipTeams, _ = fetchChampionshipPrediction(ctx, lt, r.path, logger)
		})
		run(func() {
			times, err := jp.FetchLapTimes(ctx, r.year, r.round, 1)
			if err != nil {
				logger.WarnContext(ctx, "kmf1: jolpica lap-1 times fetch failed, skipping first-lap backfill", "year", r.year, "round", r.round, "error", err)
				return
			}
			logFeedLoaded(ctx, logger, "Jolpica LapTimes (lap 1)", r.path)
			lap1Times, haveLap1Times = times, true
		})
	}

	if raceLike {
		run(func() {
			pits, err := jp.FetchPitStops(ctx, r.year, r.round)
			if err != nil {
				logger.WarnContext(ctx, "kmf1: jolpica pit stops fetch failed, leaving PitStops empty", "year", r.year, "round", r.round, "error", err)
				return
			}
			logFeedLoaded(ctx, logger, "Jolpica PitStops", r.path)
			pitStops = pits
		})
	}

	wg.Wait()

	s.drivers = driverInfo
	s.trackStatus = trackStatuses
	s.statusEvents = statusEvents
	sessionStartTime, haveSessionStartTime := feeds.SessionStartTime(statusEvents)
	s.progress = progress
	totalLaps, haveTotalLaps := latestTotalLaps(progress)

	feeds.AlignLaps(laps, intervals)
	s.intervals = intervals

	feeds.FixTyreStints(stints, laps, sessionStartTime, haveSessionStartTime, raceLike)
	s.stints = stints

	feeds.PadTelemetry(telemetry)
	s.telemetry = telemetry

	feeds.DeriveLapStartTimes(laps, sessionStartTime, haveSessionStartTime, statusEvents, raceLike)
	feeds.CalculatePositions(laps, raceLike)
	feeds.AddTrackStatusToLaps(laps, trackStatuses)
	feeds.ForwardFillSpeedTraps(laps)
	feeds.CheckLapAccuracy(laps)

	if haveLap1Times {
		session.BackfillFirstLapTime(laps, results, lap1Times)
	}

	feeds.GenerateSyntheticRetirementLaps(laps, statusEvents, trackStatuses, telemetry, totalLaps, haveTotalLaps)

	s.raceControlMessages = raceControlMessages
	feeds.MarkDeletedLaps(laps, raceControlMessages)

	switch {
	case raceLike:
		results = session.CalculateRaceResultsFromLaps(results, laps, false)
	case qualiLike:
		results = session.CalculateQualiResultsFromLaps(results, laps, statusEvents, false)
	}
	s.results = results
	s.laps = laps

	s.bestStats = bestStats
	s.weather = weather
	s.teamRadio = teamRadio
	s.topThree = topThree
	s.clock = clock
	s.archiveStatus = archiveStatus
	s.audio = audio
	s.content = content

	if r.sessionType == SessionRace {
		s.championshipDrivers = championshipDrivers
		s.championshipTeams = championshipTeams
	}

	if raceLike {
		s.pitStops = pitStops
	}

	logger.InfoContext(ctx, "kmf1: session loaded", "session", r.path, "duration", time.Since(start), "laps", len(laps), "results", len(results))

	return s, nil
}

func fetchAndReconcileResults(
	ctx context.Context,
	jp *jolpica.Client,
	year, round int,
	sessionType string,
	driverInfo map[string]*feeds.DriverInfo,
	logger *slog.Logger,
) []jolpica.SessionResults {
	var jolpicaResults []jolpica.SessionResults
	var err error
	var feed string

	switch {
	case sessionType == SessionRace:
		feed = "RaceResults"
		jolpicaResults, err = jp.FetchRaceResults(ctx, year, round)
	case sessionType == SessionSprint:
		feed = "SprintResults"
		jolpicaResults, err = jp.FetchSprintResults(ctx, year, round)
	case isQualiLike(sessionType):
		feed = "QualifyingResults"
		jolpicaResults, err = jp.FetchQualifyingResults(ctx, year, round)
	}
	if err != nil {
		logger.WarnContext(ctx, "kmf1: jolpica results fetch failed, reconciling without official results", "feed", feed, "year", year, "round", round, "error", err)
		jolpicaResults = nil
	} else if feed != "" {
		logger.InfoContext(ctx, "kmf1: jolpica results loaded", "feed", feed, "year", year, "round", round)
	}

	return session.ReconcileResults(driverInfo, jolpicaResults, ergastListAccurate(sessionType))
}

func latestTotalLaps(progress []feeds.SessionProgress) (int, bool) {
	for i := len(progress) - 1; i >= 0; i-- {
		if progress[i].TotalLaps > 0 {
			return progress[i].TotalLaps, true
		}
	}
	return 0, false
}
