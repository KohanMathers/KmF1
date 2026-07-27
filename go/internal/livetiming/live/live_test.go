package live

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/coder/websocket"

	"github.com/KohanMathers/KmF1/go/internal/livetiming/signalr"
)

type fakeF1Server struct {
	negotiate *httptest.Server
	ws        *httptest.Server
	accepts   chan *websocket.Conn
}

func newFakeF1Server(t *testing.T) *fakeF1Server {
	t.Helper()
	fs := &fakeF1Server{accepts: make(chan *websocket.Conn, 8)}

	fs.negotiate = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.SetCookie(w, &http.Cookie{Name: "AWSALBCORS", Value: "test-affinity-token"})
		w.WriteHeader(http.StatusOK)
	}))

	fs.ws = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c, err := websocket.Accept(w, r, nil)
		if err != nil {
			t.Errorf("fake F1 server accept: %v", err)
			return
		}
		ctx := context.Background()

		if _, _, err := c.Read(ctx); err != nil {
			t.Errorf("fake F1 server read handshake: %v", err)
			return
		}
		okResp, _ := json.Marshal(map[string]any{})
		okResp = append(okResp, 0x1e)
		if err := c.Write(ctx, websocket.MessageText, okResp); err != nil {
			t.Errorf("fake F1 server write handshake response: %v", err)
			return
		}

		fs.accepts <- c
	}))
	return fs
}

func (fs *fakeF1Server) close() {
	fs.negotiate.Close()
	fs.ws.CloseClientConnections()
	fs.ws.Close()
}

func (fs *fakeF1Server) wsURL() string {
	return "ws" + fs.ws.URL[len("http"):]
}

func (fs *fakeF1Server) accepted(t *testing.T) *websocket.Conn {
	t.Helper()
	select {
	case c := <-fs.accepts:
		return c
	case <-time.After(2 * time.Second):
		t.Fatal("fake F1 server never accepted a connection")
		return nil
	}
}

func readSubscribeInvocation(t *testing.T, conn *websocket.Conn) signalr.Envelope {
	t.Helper()
	_, raw, err := conn.Read(context.Background())
	if err != nil {
		t.Fatalf("read subscribe invocation: %v", err)
	}
	var env signalr.Envelope
	if err := json.Unmarshal(bytes.TrimRight(raw, "\x1e"), &env); err != nil {
		t.Fatalf("decode subscribe invocation: %v", err)
	}
	if env.Type != signalr.TypeInvocation || env.Target != "Subscribe" {
		t.Fatalf("first client message = %+v, want a Subscribe invocation", env)
	}
	return env
}

func writeEnvelope(t *testing.T, conn *websocket.Conn, env signalr.Envelope) {
	t.Helper()
	b, err := json.Marshal(env)
	if err != nil {
		t.Fatalf("marshal envelope: %v", err)
	}
	b = append(b, 0x1e)
	if err := conn.Write(context.Background(), websocket.MessageText, b); err != nil {
		t.Fatalf("write envelope: %v", err)
	}
}

func rawJSON(v any) json.RawMessage {
	b, _ := json.Marshal(v)
	return b
}

func TestRecord_SubscribesWithConfiguredTopics(t *testing.T) {
	fs := newFakeF1Server(t)
	defer fs.close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	out := make(chan Message, 8)
	done := make(chan error, 1)
	go func() {
		done <- record(ctx, out, Options{Topics: []string{"Heartbeat", "TimingData"}}, fs.wsURL(), fs.negotiate.URL)
	}()

	conn := fs.accepted(t)
	sub := readSubscribeInvocation(t, conn)
	var gotTopics []string
	if err := json.Unmarshal(sub.Arguments[0], &gotTopics); err != nil {
		t.Fatalf("decode subscribed topics: %v", err)
	}
	if len(gotTopics) != 2 || gotTopics[0] != "Heartbeat" || gotTopics[1] != "TimingData" {
		t.Errorf("subscribed topics = %v, want [Heartbeat TimingData]", gotTopics)
	}

	cancel()
	if err := <-done; err != context.Canceled {
		t.Errorf("record returned %v, want context.Canceled", err)
	}
}

func TestRecord_ForwardsSnapshotFromCompletionReply(t *testing.T) {
	fs := newFakeF1Server(t)
	defer fs.close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	out := make(chan Message, 8)
	go func() { _ = record(ctx, out, Options{}, fs.wsURL(), fs.negotiate.URL) }()

	conn := fs.accepted(t)
	sub := readSubscribeInvocation(t, conn)

	writeEnvelope(t, conn, signalr.Envelope{
		Type:         signalr.TypeCompletion,
		InvocationID: sub.InvocationID,
		Result:       rawJSON(map[string]json.RawMessage{"DriverList": rawJSON(map[string]string{"1": "VER"})}),
	})

	select {
	case m := <-out:
		if m.Topic != "DriverList" || m.Timestamp != "" {
			t.Errorf("got %+v, want topic DriverList with an empty Timestamp", m)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("no message forwarded from the Completion snapshot")
	}
}

func TestRecord_ForwardsFeedInvocations(t *testing.T) {
	fs := newFakeF1Server(t)
	defer fs.close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	out := make(chan Message, 8)
	go func() { _ = record(ctx, out, Options{}, fs.wsURL(), fs.negotiate.URL) }()

	conn := fs.accepted(t)
	readSubscribeInvocation(t, conn)

	writeEnvelope(t, conn, signalr.Envelope{
		Type:      signalr.TypeInvocation,
		Target:    "feed",
		Arguments: []json.RawMessage{rawJSON("TimingData"), rawJSON(`{"Lines":{}}`), rawJSON("2024-06-23T14:00:00Z")},
	})

	select {
	case m := <-out:
		if m.Topic != "TimingData" || m.Payload != `{"Lines":{}}` || m.Timestamp != "2024-06-23T14:00:00Z" {
			t.Errorf("got %+v, want the decoded feed message", m)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("no feed message forwarded")
	}
}

func TestRecord_HandlesMessagesLargerThanDefaultWebsocketLimit(t *testing.T) {
	fs := newFakeF1Server(t)
	defer fs.close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	out := make(chan Message, 8)
	go func() { _ = record(ctx, out, Options{}, fs.wsURL(), fs.negotiate.URL) }()

	conn := fs.accepted(t)
	readSubscribeInvocation(t, conn)

	bigPayload := strings.Repeat("x", 40*1024)
	writeEnvelope(t, conn, signalr.Envelope{
		Type:      signalr.TypeInvocation,
		Target:    "feed",
		Arguments: []json.RawMessage{rawJSON("TimingData"), rawJSON(bigPayload)},
	})

	select {
	case m := <-out:
		if len(m.Payload) != len(bigPayload) {
			t.Errorf("got payload of length %d, want %d (message was likely truncated or rejected)", len(m.Payload), len(bigPayload))
		}
	case <-time.After(2 * time.Second):
		t.Fatal("no message received — likely rejected as too big (regression: check Options.MaxMessageSize/SetReadLimit)")
	}
}

func TestRecord_ReconnectsAfterServerCloses(t *testing.T) {
	fs := newFakeF1Server(t)
	defer fs.close()

	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()

	out := make(chan Message, 8)
	go func() { _ = record(ctx, out, Options{}, fs.wsURL(), fs.negotiate.URL) }()

	first := fs.accepted(t)
	readSubscribeInvocation(t, first)
	first.Close(websocket.StatusNormalClosure, "test-triggered close")

	second := fs.accepted(t)
	readSubscribeInvocation(t, second)
}

func TestRecord_IdleTimeoutTriggersReconnect(t *testing.T) {
	fs := newFakeF1Server(t)
	defer fs.close()

	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()

	out := make(chan Message, 8)
	go func() {
		_ = record(ctx, out, Options{IdleTimeout: 200 * time.Millisecond}, fs.wsURL(), fs.negotiate.URL)
	}()

	first := fs.accepted(t)
	readSubscribeInvocation(t, first)

	second := fs.accepted(t)
	readSubscribeInvocation(t, second)
}

func TestNegotiateHeader_ForwardsAWSALBCORSCookie(t *testing.T) {
	fs := newFakeF1Server(t)
	defer fs.close()

	header, err := negotiateHeader(context.Background(), fs.negotiate.URL)
	if err != nil {
		t.Fatalf("negotiateHeader: %v", err)
	}
	cookie := header.Get("Cookie")
	if !strings.Contains(cookie, "AWSALBCORS=test-affinity-token") {
		t.Errorf("Cookie header = %q, want it to contain the AWSALBCORS cookie", cookie)
	}
}

func TestRecord_LogsConnectAndReconnectEvents(t *testing.T) {
	fs := newFakeF1Server(t)
	defer fs.close()

	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()

	var logs bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&logs, nil))

	out := make(chan Message, 8)
	go func() {
		_ = record(ctx, out, Options{Logger: logger, IdleTimeout: 200 * time.Millisecond}, fs.wsURL(), fs.negotiate.URL)
	}()

	first := fs.accepted(t)
	readSubscribeInvocation(t, first)
	fs.accepted(t)
	cancel()

	out2 := logs.String()
	if !strings.Contains(out2, "live: connecting") {
		t.Errorf("expected a connecting log line, got:\n%s", out2)
	}
	if !strings.Contains(out2, "live: subscribed") {
		t.Errorf("expected a subscribed log line, got:\n%s", out2)
	}
	if !strings.Contains(out2, "live: connection ended, reconnecting") {
		t.Errorf("expected a reconnect log line, got:\n%s", out2)
	}
}
