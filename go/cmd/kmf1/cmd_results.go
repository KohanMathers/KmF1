package main

import (
	"context"
	"fmt"
	"strconv"

	kmf1 "github.com/KohanMathers/KmF1/go"
)

func cmdResults(ctx context.Context, client *kmf1.Client, jsonOut bool, args []string) error {
	session, extra, err := loadSession(ctx, client, args)
	if err != nil {
		return err
	}
	if len(extra) != 0 {
		return fmt.Errorf("usage: results <year> <round|event name> <sessionType>")
	}

	results := session.Results()
	headers := []string{"POS", "NUM", "DRIVER", "TEAM", "STATUS", "POINTS", "TIME/GAP"}
	rows := make([][]string, len(results))
	for i, r := range results {
		rows[i] = []string{
			intOrDash(r.Position),
			r.DriverNumber,
			strOrDash(r.Abbreviation),
			strOrDash(r.TeamName),
			strOrDash(r.Status),
			strconv.FormatFloat(r.Points, 'f', -1, 64),
			durOrDash(r.Time),
		}
	}
	return output(jsonOut, results, headers, rows)
}
