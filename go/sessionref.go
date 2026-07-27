package kmf1

import (
	"context"
	"fmt"
	"log/slog"
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

	driverInfo, _ := fetchDriverList(ctx, lt, r.path, logger)
	s.drivers = driverInfo

	trackStatuses, _ := fetchTrackStatus(ctx, lt, r.path, logger)
	s.trackStatus = trackStatuses

	statusEvents, _ := fetchSessionStatus(ctx, lt, r.path, logger)
	s.statusEvents = statusEvents
	sessionStartTime, haveSessionStartTime := feeds.SessionStartTime(statusEvents)

	progress, _ := fetchLapCount(ctx, lt, r.path, logger)
	s.progress = progress
	totalLaps, haveTotalLaps := latestTotalLaps(progress)

	raceLike := isRaceLike(r.sessionType)
	qualiLike := isQualiLike(r.sessionType)

	laps, intervals := loadLapsAndIntervals(ctx, lt, r.path, logger)
	feeds.AlignLaps(laps, intervals)
	s.intervals = intervals

	stints, _ := fetchTyreStints(ctx, lt, r.path, logger)
	feeds.FixTyreStints(stints, laps, sessionStartTime, haveSessionStartTime, raceLike)
	s.stints = stints

	telemetry := loadTelemetry(ctx, lt, r.path, logger)
	feeds.PadTelemetry(telemetry)
	s.telemetry = telemetry

	feeds.DeriveLapStartTimes(laps, sessionStartTime, haveSessionStartTime, statusEvents, raceLike)
	feeds.CalculatePositions(laps, raceLike)
	feeds.AddTrackStatusToLaps(laps, trackStatuses)
	feeds.ForwardFillSpeedTraps(laps)
	feeds.CheckLapAccuracy(laps)

	results := fetchAndReconcileResults(ctx, jp, r.year, r.round, r.sessionType, driverInfo, logger)

	if r.sessionType == SessionRace {
		lap1Times, err := jp.FetchLapTimes(ctx, r.year, r.round, 1)
		if err != nil {
			logger.WarnContext(ctx, "kmf1: jolpica lap-1 times fetch failed, skipping first-lap backfill", "year", r.year, "round", r.round, "error", err)
		} else {
			logFeedLoaded(ctx, logger, "Jolpica LapTimes (lap 1)", r.path)
			session.BackfillFirstLapTime(laps, results, lap1Times)
		}
	}

	feeds.GenerateSyntheticRetirementLaps(laps, statusEvents, trackStatuses, telemetry, totalLaps, haveTotalLaps)

	raceControlMessages, _ := fetchRaceControlMessages(ctx, lt, r.path, logger)
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

	s.bestStats, _ = fetchBestStats(ctx, lt, r.path, logger)
	s.weather, _ = fetchWeather(ctx, lt, r.path, logger)
	s.teamRadio, _ = fetchTeamRadio(ctx, lt, r.path, logger)
	s.topThree, _ = fetchTopThree(ctx, lt, r.path, logger)
	s.clock, _ = fetchSessionClock(ctx, lt, r.path, logger)
	s.archiveStatus, _ = fetchArchiveStatus(ctx, lt, r.path, logger)
	s.audio, _ = fetchAudioStreams(ctx, lt, r.path, logger)
	s.content, _ = fetchContentStreams(ctx, lt, r.path, logger)

	if r.sessionType == SessionRace {
		s.championshipDrivers, s.championshipTeams, _ = fetchChampionshipPrediction(ctx, lt, r.path, logger)
	}

	if raceLike {
		pitStops, err := jp.FetchPitStops(ctx, r.year, r.round)
		if err != nil {
			logger.WarnContext(ctx, "kmf1: jolpica pit stops fetch failed, leaving PitStops empty", "year", r.year, "round", r.round, "error", err)
		} else {
			logFeedLoaded(ctx, logger, "Jolpica PitStops", r.path)
		}
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
