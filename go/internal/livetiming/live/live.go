package live

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"time"

	"github.com/coder/websocket"

	"github.com/KohanMathers/KmF1/go/internal/livetiming/signalr"
)

const (
	connectionURL = "wss://livetiming.formula1.com/signalrcore"
	negotiateURL  = "https://livetiming.formula1.com/signalrcore/negotiate"
)

var Topics = []string{
	"Heartbeat", "AudioStreams", "DriverList", "ExtrapolatedClock",
	"RaceControlMessages", "SessionInfo", "SessionStatus", "TeamRadio",
	"TimingAppData", "TimingStats", "TrackStatus", "WeatherData",
	"Position.z", "CarData.z", "ContentStreams", "SessionData",
	"TimingData", "TopThree", "RcmSeries", "LapCount",
}

type Message struct {
	Topic     string
	Payload   string
	Timestamp string
}

type Options struct {
	AuthToken string

	Topics []string

	IdleTimeout time.Duration

	Logger *slog.Logger

	MaxMessageSize int64
}

func (o Options) logger() *slog.Logger {
	if o.Logger != nil {
		return o.Logger
	}
	return slog.New(slog.DiscardHandler)
}

func (o Options) topics() []string {
	if o.Topics != nil {
		return o.Topics
	}
	return Topics
}

func (o Options) idleTimeout() time.Duration {
	if o.IdleTimeout == 0 {
		return 60 * time.Second
	}
	if o.IdleTimeout < 0 {
		return 0
	}
	return o.IdleTimeout
}

const defaultMaxMessageSize = 4 << 20

func (o Options) maxMessageSize() int64 {
	if o.MaxMessageSize == 0 {
		return defaultMaxMessageSize
	}
	if o.MaxMessageSize < 0 {
		return -1
	}
	return o.MaxMessageSize
}

func Record(ctx context.Context, out chan<- Message, opts Options) error {
	return record(ctx, out, opts, connectionURL, negotiateURL)
}

func record(ctx context.Context, out chan<- Message, opts Options, connectionURL, negotiateURL string) error {
	logger := opts.logger()
	topics := opts.topics()
	idleTimeout := opts.idleTimeout()
	maxMessageSize := opts.maxMessageSize()

	const (
		initialBackoff = time.Second
		maxBackoff     = 30 * time.Second
	)
	backoff := initialBackoff

	for {
		logger.InfoContext(ctx, "live: connecting", "url", connectionURL, "topics", len(topics))
		err := connectAndStream(ctx, out, topics, opts.AuthToken, idleTimeout, maxMessageSize, logger, connectionURL, negotiateURL)
		if ctx.Err() != nil {
			return ctx.Err()
		}
		logger.WarnContext(ctx, "live: connection ended, reconnecting", "backoff", backoff, "error", err)

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(backoff):
		}
		backoff *= 2
		if backoff > maxBackoff {
			backoff = maxBackoff
		}
	}
}

func connectAndStream(ctx context.Context, out chan<- Message, topics []string, authToken string, idleTimeout time.Duration, maxMessageSize int64, logger *slog.Logger, connectionURL, negotiateURL string) error {
	header, err := negotiateHeader(ctx, negotiateURL)
	if err != nil {
		return fmt.Errorf("live: negotiate: %w", err)
	}

	dialURL := connectionURL
	if authToken != "" {
		dialURL += "?access_token=" + url.QueryEscape(authToken)
	}

	conn, err := signalr.Dial(ctx, dialURL, &websocket.DialOptions{HTTPHeader: header})
	if err != nil {
		return fmt.Errorf("live: connect: %w", err)
	}
	defer conn.CloseNow()
	conn.SetReadLimit(maxMessageSize)

	if err := conn.Invoke(ctx, "Subscribe", "0", topics); err != nil {
		return fmt.Errorf("live: subscribe: %w", err)
	}
	logger.InfoContext(ctx, "live: subscribed", "topics", topics)

	for {
		readCtx := ctx
		var cancel context.CancelFunc
		if idleTimeout > 0 {
			readCtx, cancel = context.WithTimeout(ctx, idleTimeout)
		}
		env, err := conn.NextEnvelope(readCtx)
		if cancel != nil {
			cancel()
		}
		if err != nil {
			if idleTimeout > 0 && ctx.Err() == nil && readCtx.Err() == context.DeadlineExceeded {
				return fmt.Errorf("live: idle timeout after %s with no messages", idleTimeout)
			}
			return err
		}

		if err := dispatch(ctx, out, env); err != nil {
			return err
		}
	}
}

func dispatch(ctx context.Context, out chan<- Message, env signalr.Envelope) error {
	switch env.Type {
	case signalr.TypeClose:
		return fmt.Errorf("live: server closed the connection: %s", env.Error)

	case signalr.TypeCompletion:

		if env.InvocationID != "0" || len(env.Result) == 0 {
			return nil
		}
		var snapshot map[string]json.RawMessage
		if err := json.Unmarshal(env.Result, &snapshot); err != nil {
			return nil
		}
		for topic, raw := range snapshot {
			select {
			case out <- Message{Topic: topic, Payload: string(raw)}:
			case <-ctx.Done():
				return ctx.Err()
			}
		}
		return nil

	case signalr.TypeInvocation:
		if env.Target != "feed" || len(env.Arguments) < 2 {
			return nil
		}
		var topic, payload string
		if err := json.Unmarshal(env.Arguments[0], &topic); err != nil {
			return nil
		}
		if err := json.Unmarshal(env.Arguments[1], &payload); err != nil {
			return nil
		}
		var timestamp string
		if len(env.Arguments) >= 3 {
			_ = json.Unmarshal(env.Arguments[2], &timestamp)
		}
		select {
		case out <- Message{Topic: topic, Payload: payload, Timestamp: timestamp}:
		case <-ctx.Done():
			return ctx.Err()
		}
		return nil

	default:
		return nil
	}
}

func negotiateHeader(ctx context.Context, negotiateURL string) (http.Header, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodOptions, negotiateURL, nil)
	if err != nil {
		return nil, err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	header := http.Header{}
	for _, c := range resp.Cookies() {
		if c.Name == "AWSALBCORS" {
			header.Set("Cookie", c.Name+"="+c.Value)
			break
		}
	}
	return header, nil
}
