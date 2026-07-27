package main

import (
	"context"
	"fmt"

	kmf1 "github.com/KohanMathers/KmF1/go"
)

func cmdDrivers(ctx context.Context, client *kmf1.Client, jsonOut bool, args []string) error {
	session, extra, err := loadSession(ctx, client, args)
	if err != nil {
		return err
	}
	if len(extra) != 0 {
		return fmt.Errorf("usage: drivers <year> <round|event name> <sessionType>")
	}

	drivers := session.Drivers()
	headers := []string{"NUM", "ABBR", "NAME", "TEAM", "COLOR", "COUNTRY"}
	rows := make([][]string, len(drivers))
	for i, d := range drivers {
		rows[i] = []string{
			d.DriverNumber,
			d.Abbreviation,
			strOrDash(d.FullName),
			d.TeamName,
			strOrDash(d.TeamColour),
			strOrDash(d.CountryCode),
		}
	}
	return output(jsonOut, drivers, headers, rows)
}
