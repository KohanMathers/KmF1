package main

import (
	"context"
	"fmt"
	"strconv"

	kmf1 "github.com/KohanMathers/KmF1/go"
	"github.com/KohanMathers/KmF1/go/internal/livetiming/feeds"
)

func cmdLaps(ctx context.Context, client *kmf1.Client, jsonOut bool, args []string) error {
	session, extra, err := loadSession(ctx, client, args)
	if err != nil {
		return err
	}
	if len(extra) > 1 {
		return fmt.Errorf("usage: laps <year> <round|event name> <sessionType> [driver]")
	}

	var laps []feeds.LapData
	if len(extra) == 1 {
		driver := session.Driver(extra[0])
		if !driver.Exists() {
			return fmt.Errorf("no driver matching %q in this session's roster", extra[0])
		}
		laps = driver.Laps()
	} else {
		laps = session.Laps()
	}

	headers := []string{"DRIVER", "LAP", "TIME", "S1", "S2", "S3", "POS", "PB", "DELETED"}
	rows := make([][]string, len(laps))
	for i, l := range laps {
		rows[i] = []string{
			l.DriverNumber,
			strconv.Itoa(l.LapNo),
			durOrDash(l.LapTime),
			durOrDash(l.Sector1Time),
			durOrDash(l.Sector2Time),
			durOrDash(l.Sector3Time),
			intOrDash(l.Position),
			boolYesNo(l.IsPersonalBest),
			boolYesNo(l.Deleted),
		}
	}
	return output(jsonOut, laps, headers, rows)
}
