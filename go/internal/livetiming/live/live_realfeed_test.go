package live

import (
	"bytes"
	"context"
	"log/slog"
	"os"
	"testing"
	"time"
)

func TestRecord_RealFeed_RequiresAuthToken(t *testing.T) {
	token := os.Getenv("KMF1_LIVE_AUTH_TOKEN")
	if token == "" {
		t.Skip("KMF1_LIVE_AUTH_TOKEN not set — skipping the real live-feed test; set it to your own F1TV subscription token to run this against the real F1 SignalR endpoint")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	var logs bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&logs, nil))

	out := make(chan Message, 64)
	done := make(chan error, 1)
	go func() {
		done <- Record(ctx, out, Options{
			AuthToken:   token,
			IdleTimeout: 15 * time.Second,
			Logger:      logger,
		})
	}()

	var received []Message
collect:
	for {
		select {
		case m := <-out:
			received = append(received, m)
			if len(received) >= 5 {
				cancel()
			}
		case err := <-done:
			if err != nil && err != context.Canceled && err != context.DeadlineExceeded {
				t.Fatalf("Record returned an unexpected error against the real feed: %v\nlog output:\n%s", err, logs.String())
			}
			break collect
		case <-ctx.Done():
			break collect
		}
	}

	if len(received) == 0 {
		t.Logf("connected to the real live feed but received zero messages within the timeout — expected outside an active live session; log output:\n%s", logs.String())
		return
	}

	t.Logf("received %d real live-feed message(s), first topic = %q", len(received), received[0].Topic)
	for _, m := range received {
		if m.Topic == "" {
			t.Errorf("got a message with an empty Topic: %+v", m)
		}
	}
}
