package kmf1

import (
	"context"
	"testing"
	"time"

	"github.com/KohanMathers/KmF1/go/internal/livetiming/live"
	"github.com/KohanMathers/KmF1/go/internal/livetiming/wire"
	"github.com/KohanMathers/KmF1/go/internal/testfixture"
)

var realBase = time.Date(2024, 6, 23, 14, 0, 0, 0, time.UTC)

func readFixture(t *testing.T, path string) []byte {
	t.Helper()
	return testfixture.Read(t, path)
}

func recordsToMessages(t *testing.T, topic string, data []byte) []live.Message {
	t.Helper()
	records, err := wire.SplitFrames(data)
	if err != nil {
		t.Fatalf("split frames for %s: %v", topic, err)
	}
	out := make([]live.Message, len(records))
	for i, rec := range records {
		out[i] = live.Message{
			Topic:     topic,
			Payload:   rec.Payload,
			Timestamp: realBase.Add(rec.Elapsed).Format(time.RFC3339Nano),
		}
	}
	return out
}

func TestLiveSession_ConsumeRealFixtures_DriverListAndTrackStatus(t *testing.T) {

	messages := recordsToMessages(t, "TrackStatus", readFixture(t, "internal/livetiming/feeds/testdata/trackstatus.jsonstream"))
	messages = append(messages, recordsToMessages(t, "DriverList", readFixture(t, "internal/livetiming/feeds/testdata/driverlist.jsonstream"))...)

	ch := make(chan live.Message, len(messages))
	for _, m := range messages {
		ch <- m
	}
	close(ch)

	session := NewLiveSession()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	session.Consume(ctx, ch)

	drivers := session.Drivers()
	if len(drivers) != 20 {
		t.Fatalf("Drivers() = %d entries, want 20 (matches TestParseDriverList_RealCapture)", len(drivers))
	}
	found := false
	for _, d := range drivers {
		if d.DriverNumber != "1" {
			continue
		}
		found = true
		if d.Abbreviation != "VER" || d.Line != 1 {
			t.Errorf("driver 1 = %+v, want Abbreviation=VER, Line=1 (final merge-patched value)", d)
		}
	}
	if !found {
		t.Fatal("driver 1 (VER) not found in LiveSession.Drivers()")
	}

	status := session.TrackStatus()
	if len(status) != 2 {
		t.Fatalf("TrackStatus() = %d events, want 2 (matches TestParseTrackStatus_RealCapture)", len(status))
	}
	if status[0].Status != "2" || status[0].Message != "Yellow" || status[0].Timestamp != 0 {
		t.Errorf("first TrackStatus event = %+v, want {2 Yellow 0}", status[0])
	}
	wantSecond := 10*time.Minute + 895*time.Millisecond
	if status[1].Timestamp != wantSecond {
		t.Errorf("second TrackStatus event Timestamp = %v, want %v (anchor must exactly reproduce the archive's own elapsed values)", status[1].Timestamp, wantSecond)
	}
}

func TestLiveSession_ConsumeRealFixtures_TimingDataProducesLaps(t *testing.T) {
	messages := recordsToMessages(t, "TimingData", readFixture(t, "internal/livetiming/feeds/testdata/timingdata.jsonstream"))

	ch := make(chan live.Message, len(messages))
	for _, m := range messages {
		ch <- m
	}
	close(ch)

	session := NewLiveSession()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	session.Consume(ctx, ch)

	laps := session.Laps()
	if len(laps) == 0 {
		t.Fatal("Laps() is empty after feeding the real TimingData fixture through LiveSession")
	}

	intervals := session.Intervals()
	if len(intervals) == 0 {
		t.Fatal("Intervals() is empty after feeding the same real TimingData fixture through LiveSession")
	}
}

func TestLiveSession_UnknownTopicIsIgnoredNotAnError(t *testing.T) {
	ch := make(chan live.Message, 1)
	ch <- live.Message{Topic: "Heartbeat", Payload: `{"Utc":"2024-06-23T14:00:00Z"}`, Timestamp: "2024-06-23T14:00:00Z"}
	close(ch)

	session := NewLiveSession()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	session.Consume(ctx, ch)

	if len(session.Drivers()) != 0 {
		t.Error("a Heartbeat message should not have populated any decoder")
	}
}
