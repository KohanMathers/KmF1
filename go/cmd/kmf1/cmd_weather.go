package main

import (
	"context"
	"fmt"
	"strconv"

	kmf1 "github.com/KohanMathers/KmF1/go"
)

func cmdWeather(ctx context.Context, client *kmf1.Client, jsonOut bool, args []string) error {
	session, extra, err := loadSession(ctx, client, args)
	if err != nil {
		return err
	}
	if len(extra) != 0 {
		return fmt.Errorf("usage: weather <year> <round|event name> <sessionType>")
	}

	samples := session.Weather()
	headers := []string{"TIME", "AIR °C", "TRACK °C", "HUMIDITY", "RAIN", "WIND km/h"}
	rows := make([][]string, len(samples))
	for i, w := range samples {
		rows[i] = []string{
			w.Timestamp.String(),
			strconv.FormatFloat(w.AirTemp, 'f', 1, 64),
			strconv.FormatFloat(w.TrackTemp, 'f', 1, 64),
			strconv.FormatFloat(w.Humidity, 'f', 1, 64),
			boolYesNo(w.Rainfall),
			strconv.FormatFloat(w.WindSpeed, 'f', 1, 64),
		}
	}
	return output(jsonOut, samples, headers, rows)
}
