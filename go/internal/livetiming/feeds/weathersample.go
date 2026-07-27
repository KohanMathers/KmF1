package feeds

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/KohanMathers/KmF1/go/internal/livetiming/wire"
)

type WeatherSample struct {
	AirTemp       float64
	TrackTemp     float64
	Humidity      float64
	Pressure      float64
	Rainfall      bool
	WindDirection int
	WindSpeed     float64
	Timestamp     time.Duration
}

type rawWeatherSample struct {
	AirTemp       string `json:"AirTemp"`
	TrackTemp     string `json:"TrackTemp"`
	Humidity      string `json:"Humidity"`
	Pressure      string `json:"Pressure"`
	Rainfall      string `json:"Rainfall"`
	WindDirection string `json:"WindDirection"`
	WindSpeed     string `json:"WindSpeed"`
}

func ParseWeatherSamples(data []byte) ([]WeatherSample, error) {
	records, err := wire.SplitFrames(data)
	if err != nil {
		return nil, fmt.Errorf("feeds: WeatherData: %w", err)
	}

	var d WeatherSampleDecoder
	for i, rec := range records {
		if err := d.Feed(rec.Elapsed, rec.Payload); err != nil {
			return nil, fmt.Errorf("feeds: WeatherData: record %d: %w", i, err)
		}
	}
	return d.Snapshot(), nil
}

type WeatherSampleDecoder struct {
	samples []WeatherSample
}

func (d *WeatherSampleDecoder) Feed(elapsed time.Duration, payload string) error {
	decoded, err := wire.Decode(payload, false)
	if err != nil {
		return err
	}

	var raw rawWeatherSample
	if err := json.Unmarshal(decoded, &raw); err != nil {
		return fmt.Errorf("unmarshal: %w", err)
	}

	windDir, _ := strconv.Atoi(raw.WindDirection)
	d.samples = append(d.samples, WeatherSample{
		AirTemp:       parseWeatherFloat(raw.AirTemp),
		TrackTemp:     parseWeatherFloat(raw.TrackTemp),
		Humidity:      parseWeatherFloat(raw.Humidity),
		Pressure:      parseWeatherFloat(raw.Pressure),
		Rainfall:      raw.Rainfall == "1",
		WindDirection: windDir,
		WindSpeed:     parseWeatherFloat(raw.WindSpeed),
		Timestamp:     elapsed,
	})
	return nil
}

func (d *WeatherSampleDecoder) Snapshot() []WeatherSample {
	out := make([]WeatherSample, len(d.samples))
	copy(out, d.samples)
	return out
}

func parseWeatherFloat(s string) float64 {
	v, _ := strconv.ParseFloat(s, 64)
	return v
}
