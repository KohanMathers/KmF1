package kmf1

import (
	"context"
	"fmt"
	"log/slog"
	"sync"

	"github.com/KohanMathers/KmF1/go/internal/livetiming"
	"github.com/KohanMathers/KmF1/go/internal/livetiming/feeds"
)

func logFeedFetchFailure(ctx context.Context, logger *slog.Logger, feed, path string, err error) {
	logger.WarnContext(ctx, "kmf1: feed fetch failed, leaving its data empty", "feed", feed, "session", path, "error", err)
}

func logFeedParseFailure(ctx context.Context, logger *slog.Logger, feed, path string, err error) {
	logger.WarnContext(ctx, "kmf1: feed parse failed, leaving its data empty", "feed", feed, "session", path, "error", err)
}

func logFeedLoaded(ctx context.Context, logger *slog.Logger, feed, path string) {
	logger.InfoContext(ctx, "kmf1: feed loaded", "feed", feed, "session", path)
}

func fetchDriverList(ctx context.Context, lt *livetiming.Client, path string, logger *slog.Logger) (map[string]*feeds.DriverInfo, error) {
	raw, err := lt.FetchFeed(ctx, path, "DriverList.jsonStream")
	if err != nil {
		logFeedFetchFailure(ctx, logger, "DriverList", path, err)
		return nil, err
	}
	result, err := feeds.ParseDriverList(raw)
	if err != nil {
		logFeedParseFailure(ctx, logger, "DriverList", path, err)
		return result, err
	}
	logFeedLoaded(ctx, logger, "DriverList", path)
	return result, nil
}

func fetchTrackStatus(ctx context.Context, lt *livetiming.Client, path string, logger *slog.Logger) ([]feeds.TrackStatus, error) {
	raw, err := lt.FetchFeed(ctx, path, "TrackStatus.jsonStream")
	if err != nil {
		logFeedFetchFailure(ctx, logger, "TrackStatus", path, err)
		return nil, err
	}
	result, err := feeds.ParseTrackStatus(raw)
	if err != nil {
		logFeedParseFailure(ctx, logger, "TrackStatus", path, err)
		return result, err
	}
	logFeedLoaded(ctx, logger, "TrackStatus", path)
	return result, nil
}

func fetchSessionStatus(ctx context.Context, lt *livetiming.Client, path string, logger *slog.Logger) ([]feeds.SessionStatusEvent, error) {
	raw, err := lt.FetchFeed(ctx, path, "SessionStatus.jsonStream")
	if err != nil {
		logFeedFetchFailure(ctx, logger, "SessionStatus", path, err)
		return nil, err
	}
	result, err := feeds.ParseSessionStatus(raw)
	if err != nil {
		logFeedParseFailure(ctx, logger, "SessionStatus", path, err)
		return result, err
	}
	logFeedLoaded(ctx, logger, "SessionStatus", path)
	return result, nil
}

func fetchLapCount(ctx context.Context, lt *livetiming.Client, path string, logger *slog.Logger) ([]feeds.SessionProgress, error) {
	raw, err := lt.FetchFeed(ctx, path, "LapCount.jsonStream")
	if err != nil {
		logFeedFetchFailure(ctx, logger, "LapCount", path, err)
		return nil, err
	}
	result, err := feeds.ParseLapCount(raw)
	if err != nil {
		logFeedParseFailure(ctx, logger, "LapCount", path, err)
		return result, err
	}
	logFeedLoaded(ctx, logger, "LapCount", path)
	return result, nil
}

func loadLapsAndIntervals(ctx context.Context, lt *livetiming.Client, path string, logger *slog.Logger) (map[string][]feeds.LapData, map[string][]feeds.TimingInterval) {
	raw, err := lt.FetchFeed(ctx, path, "TimingData.jsonStream")
	if err != nil {
		logFeedFetchFailure(ctx, logger, "TimingData", path, err)
		return nil, nil
	}
	laps, err := feeds.ParseLapData(raw)
	if err != nil {
		logFeedParseFailure(ctx, logger, "TimingData (laps)", path, err)
		laps = nil
	} else {
		logFeedLoaded(ctx, logger, "TimingData (laps)", path)
	}
	intervals, err := feeds.ParseTimingIntervals(raw)
	if err != nil {
		logFeedParseFailure(ctx, logger, "TimingData (intervals)", path, err)
		intervals = nil
	} else {
		logFeedLoaded(ctx, logger, "TimingData (intervals)", path)
	}
	return laps, intervals
}

func fetchTyreStints(ctx context.Context, lt *livetiming.Client, path string, logger *slog.Logger) (map[string][]feeds.TyreStint, error) {
	raw, err := lt.FetchFeed(ctx, path, "TimingAppData.jsonStream")
	if err != nil {
		logFeedFetchFailure(ctx, logger, "TimingAppData", path, err)
		return nil, err
	}
	result, err := feeds.ParseTyreStints(raw)
	if err != nil {
		logFeedParseFailure(ctx, logger, "TimingAppData", path, err)
		return result, err
	}
	logFeedLoaded(ctx, logger, "TimingAppData", path)
	return result, nil
}

func loadTelemetry(ctx context.Context, lt *livetiming.Client, path string, logger *slog.Logger) map[string][]feeds.Telemetry {
	var carRaw, posRaw []byte
	var carErr, posErr error

	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		carRaw, carErr = lt.FetchFeed(ctx, path, "CarData.z.jsonStream")
	}()
	go func() {
		defer wg.Done()
		posRaw, posErr = lt.FetchFeed(ctx, path, "Position.z.jsonStream")
	}()
	wg.Wait()

	if carErr != nil {
		logFeedFetchFailure(ctx, logger, "CarData.z", path, carErr)
		return nil
	}
	if posErr != nil {
		logFeedFetchFailure(ctx, logger, "Position.z", path, posErr)
		return nil
	}

	telemetry, err := feeds.ParseTelemetry(carRaw, posRaw)
	if err != nil {
		logFeedParseFailure(ctx, logger, "CarData.z+Position.z (telemetry)", path, err)
		return nil
	}
	logFeedLoaded(ctx, logger, "CarData.z+Position.z (telemetry)", path)
	return telemetry
}

func fetchBestStats(ctx context.Context, lt *livetiming.Client, path string, logger *slog.Logger) (map[string][]feeds.DriverBestStats, error) {
	raw, err := lt.FetchFeed(ctx, path, "TimingStats.jsonStream")
	if err != nil {
		logFeedFetchFailure(ctx, logger, "TimingStats", path, err)
		return nil, err
	}
	result, err := feeds.ParseDriverBestStats(raw)
	if err != nil {
		logFeedParseFailure(ctx, logger, "TimingStats", path, err)
		return result, err
	}
	logFeedLoaded(ctx, logger, "TimingStats", path)
	return result, nil
}

func fetchWeather(ctx context.Context, lt *livetiming.Client, path string, logger *slog.Logger) ([]feeds.WeatherSample, error) {
	raw, err := lt.FetchFeed(ctx, path, "WeatherData.jsonStream")
	if err != nil {
		logFeedFetchFailure(ctx, logger, "WeatherData", path, err)
		return nil, err
	}
	result, err := feeds.ParseWeatherSamples(raw)
	if err != nil {
		logFeedParseFailure(ctx, logger, "WeatherData", path, err)
		return result, err
	}
	logFeedLoaded(ctx, logger, "WeatherData", path)
	return result, nil
}

func fetchRaceControlMessages(ctx context.Context, lt *livetiming.Client, path string, logger *slog.Logger) ([]feeds.RaceControlMessage, error) {
	raw, err := lt.FetchFeed(ctx, path, "RaceControlMessages.jsonStream")
	if err != nil {
		logFeedFetchFailure(ctx, logger, "RaceControlMessages", path, err)
		return nil, err
	}
	result, err := feeds.ParseRaceControlMessages(raw)
	if err != nil {
		logFeedParseFailure(ctx, logger, "RaceControlMessages", path, err)
		return result, err
	}
	logFeedLoaded(ctx, logger, "RaceControlMessages", path)
	return result, nil
}

func fetchTeamRadio(ctx context.Context, lt *livetiming.Client, path string, logger *slog.Logger) ([]feeds.TeamRadio, error) {
	raw, err := lt.FetchFeed(ctx, path, "TeamRadio.jsonStream")
	if err != nil {
		logFeedFetchFailure(ctx, logger, "TeamRadio", path, err)
		return nil, err
	}
	result, err := feeds.ParseTeamRadio(raw)
	if err != nil {
		logFeedParseFailure(ctx, logger, "TeamRadio", path, err)
		return result, err
	}
	logFeedLoaded(ctx, logger, "TeamRadio", path)
	return result, nil
}

func fetchTopThree(ctx context.Context, lt *livetiming.Client, path string, logger *slog.Logger) (map[int][]feeds.TopThreeLive, error) {
	raw, err := lt.FetchFeed(ctx, path, "TopThree.jsonStream")
	if err != nil {
		logFeedFetchFailure(ctx, logger, "TopThree", path, err)
		return nil, err
	}
	result, err := feeds.ParseTopThreeLive(raw)
	if err != nil {
		logFeedParseFailure(ctx, logger, "TopThree", path, err)
		return result, err
	}
	logFeedLoaded(ctx, logger, "TopThree", path)
	return result, nil
}

func fetchSessionClock(ctx context.Context, lt *livetiming.Client, path string, logger *slog.Logger) ([]feeds.SessionClock, error) {
	raw, err := lt.FetchFeed(ctx, path, "ExtrapolatedClock.jsonStream")
	if err != nil {
		logFeedFetchFailure(ctx, logger, "ExtrapolatedClock", path, err)
		return nil, err
	}
	result, err := feeds.ParseSessionClock(raw)
	if err != nil {
		logFeedParseFailure(ctx, logger, "ExtrapolatedClock", path, err)
		return result, err
	}
	logFeedLoaded(ctx, logger, "ExtrapolatedClock", path)
	return result, nil
}

func fetchArchiveStatus(ctx context.Context, lt *livetiming.Client, path string, logger *slog.Logger) (feeds.ArchiveStatus, error) {
	raw, err := lt.FetchFeed(ctx, path, "ArchiveStatus.json")
	if err != nil {
		logFeedFetchFailure(ctx, logger, "ArchiveStatus", path, err)
		return feeds.ArchiveStatus{}, err
	}
	result, err := feeds.ParseArchiveStatus(raw)
	if err != nil {
		logFeedParseFailure(ctx, logger, "ArchiveStatus", path, err)
		return result, err
	}
	logFeedLoaded(ctx, logger, "ArchiveStatus", path)
	return result, nil
}

func fetchAudioStreams(ctx context.Context, lt *livetiming.Client, path string, logger *slog.Logger) ([]feeds.AudioStream, error) {
	raw, err := lt.FetchFeed(ctx, path, "AudioStreams.jsonStream")
	if err != nil {
		logFeedFetchFailure(ctx, logger, "AudioStreams", path, err)
		return nil, err
	}
	result, err := feeds.ParseAudioStreams(raw)
	if err != nil {
		logFeedParseFailure(ctx, logger, "AudioStreams", path, err)
		return result, err
	}
	logFeedLoaded(ctx, logger, "AudioStreams", path)
	return result, nil
}

func fetchContentStreams(ctx context.Context, lt *livetiming.Client, path string, logger *slog.Logger) ([]feeds.ContentStream, error) {
	raw, err := lt.FetchFeed(ctx, path, "ContentStreams.jsonStream")
	if err != nil {
		logFeedFetchFailure(ctx, logger, "ContentStreams", path, err)
		return nil, err
	}
	result, err := feeds.ParseContentStreams(raw)
	if err != nil {
		logFeedParseFailure(ctx, logger, "ContentStreams", path, err)
		return result, err
	}
	logFeedLoaded(ctx, logger, "ContentStreams", path)
	return result, nil
}

func fetchChampionshipPrediction(ctx context.Context, lt *livetiming.Client, path string, logger *slog.Logger) (map[string][]feeds.ChampionshipPredictionDriver, map[string][]feeds.ChampionshipPredictionTeam, error) {
	raw, err := lt.FetchFeed(ctx, path, "ChampionshipPrediction.jsonStream")
	if err != nil {
		logFeedFetchFailure(ctx, logger, "ChampionshipPrediction", path, err)
		return nil, nil, fmt.Errorf("fetch ChampionshipPrediction: %w", err)
	}
	drivers, teams, err := feeds.ParseChampionshipPrediction(raw)
	if err != nil {
		logFeedParseFailure(ctx, logger, "ChampionshipPrediction", path, err)
		return drivers, teams, err
	}
	logFeedLoaded(ctx, logger, "ChampionshipPrediction", path)
	return drivers, teams, nil
}
