package feeds

import (
	"os"
	"testing"
	"time"
)

func TestParseWeatherSamples_RealCapture(t *testing.T) {
	data, err := os.ReadFile("testdata/weatherdata.jsonstream")
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}

	samples, err := ParseWeatherSamples(data)
	if err != nil {
		t.Fatalf("ParseWeatherSamples: %v", err)
	}
	if len(samples) != 154 {
		t.Fatalf("got %d samples, want 154", len(samples))
	}

	first := samples[0]
	want := WeatherSample{
		AirTemp: 23.9, Humidity: 63.0, Pressure: 1001.8, Rainfall: false,
		TrackTemp: 39.1, WindDirection: 244, WindSpeed: 2.7,
		Timestamp: 20*time.Second + 654*time.Millisecond,
	}
	if first != want {
		t.Errorf("first sample = %+v, want %+v", first, want)
	}

	found := false
	for _, s := range samples {
		if s.Timestamp == time.Hour+20*time.Second+889*time.Millisecond {
			found = true
			if !s.Rainfall {
				t.Error("expected Rainfall = true for the real rain reading")
			}
		}
	}
	if !found {
		t.Fatal("expected sample fixture assumption changed: rain-reading timestamp not found")
	}
}

func TestParseWeatherSamples_RejectsMalformed(t *testing.T) {
	if _, err := ParseWeatherSamples([]byte("00:00:00.000not json\r\n")); err == nil {
		t.Fatal("expected an error for a malformed payload")
	}
}
