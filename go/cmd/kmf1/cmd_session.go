package main

import (
	"context"
	"fmt"

	kmf1 "github.com/KohanMathers/KmF1/go"
	"github.com/KohanMathers/KmF1/go/internal/livetiming/feeds"
)

type sessionSummary struct {
	Info        *feeds.SessionInfo
	DriverCount int
	LapCount    int
	ResultCount int
}

func cmdSession(ctx context.Context, client *kmf1.Client, jsonOut bool, args []string) error {
	session, extra, err := loadSession(ctx, client, args)
	if err != nil {
		return err
	}
	if len(extra) != 0 {
		return fmt.Errorf("usage: session <year> <round|event name> <sessionType>")
	}

	summary := sessionSummary{
		Info:        session.Info(),
		DriverCount: len(session.Drivers()),
		LapCount:    len(session.Laps()),
		ResultCount: len(session.Results()),
	}

	if jsonOut {
		return output(jsonOut, summary, nil, nil)
	}

	info := summary.Info
	if info == nil {
		fmt.Println("(no SessionInfo)")
		return nil
	}
	fmt.Printf("%s — %s\n", info.Meeting.Name, info.Name)
	fmt.Printf("  Location:   %s, %s\n", info.Meeting.Location, info.Meeting.Country.Name)
	fmt.Printf("  Start:      %s\n", info.StartDate.Format("2006-01-02 15:04 MST"))
	fmt.Printf("  Drivers:    %d\n", summary.DriverCount)
	fmt.Printf("  Laps:       %d\n", summary.LapCount)
	fmt.Printf("  Results:    %d\n", summary.ResultCount)
	return nil
}
