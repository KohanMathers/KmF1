package signalr

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/coder/websocket"
)

type fakeServer struct {
	t    *testing.T
	srv  *httptest.Server
	conn chan *websocket.Conn
}

func newFakeServer(t *testing.T, handshakeErr string) *fakeServer {
	t.Helper()
	fs := &fakeServer{t: t, conn: make(chan *websocket.Conn, 1)}

	fs.srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c, err := websocket.Accept(w, r, nil)
		if err != nil {
			t.Errorf("fake server accept: %v", err)
			return
		}

		ctx := context.Background()
		_, data, err := c.Read(ctx)
		if err != nil {
			t.Errorf("fake server read handshake request: %v", err)
			return
		}
		if len(data) == 0 || data[len(data)-1] != recordSeparator {
			t.Errorf("fake server: handshake request not record-separator-terminated: %q", data)
		}

		var resp []byte
		if handshakeErr != "" {
			resp, _ = json.Marshal(map[string]string{"error": handshakeErr})
		} else {
			resp, _ = json.Marshal(map[string]any{})
		}
		resp = append(resp, recordSeparator)
		if err := c.Write(ctx, websocket.MessageText, resp); err != nil {
			t.Errorf("fake server write handshake response: %v", err)
			return
		}

		fs.conn <- c
	}))
	return fs
}

func (fs *fakeServer) url() string {
	return "ws" + fs.srv.URL[len("http"):]
}

func (fs *fakeServer) accepted(t *testing.T) *websocket.Conn {
	t.Helper()
	select {
	case c := <-fs.conn:
		return c
	case <-time.After(2 * time.Second):
		t.Fatal("fake server never accepted a connection")
		return nil
	}
}

func (fs *fakeServer) close() {
	fs.srv.CloseClientConnections()
	fs.srv.Close()
}

func TestDial_SuccessfulHandshake(t *testing.T) {
	fs := newFakeServer(t, "")
	defer fs.close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, err := Dial(ctx, fs.url(), nil)
	if err != nil {
		t.Fatalf("Dial: %v", err)
	}
	defer conn.CloseNow()
}

func TestDial_RejectedHandshake(t *testing.T) {
	fs := newFakeServer(t, "protocol not supported")
	defer fs.close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := Dial(ctx, fs.url(), nil)
	if err == nil {
		t.Fatal("Dial succeeded despite a rejected handshake, want an error")
	}
}

func TestInvokeAndNextEnvelope_RoundTrip(t *testing.T) {
	fs := newFakeServer(t, "")
	defer fs.close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	clientConn, err := Dial(ctx, fs.url(), nil)
	if err != nil {
		t.Fatalf("Dial: %v", err)
	}
	defer clientConn.CloseNow()

	serverConn := fs.accepted(t)

	if err := clientConn.Invoke(ctx, "Subscribe", "0", []string{"Heartbeat", "TimingData"}); err != nil {
		t.Fatalf("Invoke: %v", err)
	}

	_, raw, err := serverConn.Read(ctx)
	if err != nil {
		t.Fatalf("server read invocation: %v", err)
	}
	var got Envelope
	if err := json.Unmarshal(raw[:len(raw)-1], &got); err != nil {
		t.Fatalf("server decode invocation: %v", err)
	}
	if got.Type != TypeInvocation || got.Target != "Subscribe" || got.InvocationID != "0" {
		t.Fatalf("server saw envelope %+v, want a Subscribe invocation with id 0", got)
	}
	var topics []string
	if err := json.Unmarshal(got.Arguments[0], &topics); err != nil {
		t.Fatalf("decode topics argument: %v", err)
	}
	if len(topics) != 2 || topics[0] != "Heartbeat" || topics[1] != "TimingData" {
		t.Errorf("topics = %v, want [Heartbeat TimingData]", topics)
	}

	reply, _ := json.Marshal(Envelope{
		Type:      TypeInvocation,
		Target:    "feed",
		Arguments: []json.RawMessage{[]byte(`"TimingData"`), []byte(`"{\"x\":1}"`), []byte(`"2024-01-01T00:00:00Z"`)},
	})
	reply = append(reply, recordSeparator)
	if err := serverConn.Write(ctx, websocket.MessageText, reply); err != nil {
		t.Fatalf("server write feed message: %v", err)
	}

	env, err := clientConn.NextEnvelope(ctx)
	if err != nil {
		t.Fatalf("NextEnvelope: %v", err)
	}
	if env.Type != TypeInvocation || env.Target != "feed" {
		t.Fatalf("NextEnvelope = %+v, want a feed invocation", env)
	}
}

func TestNextEnvelope_MultipleMessagesInOneFrame(t *testing.T) {
	fs := newFakeServer(t, "")
	defer fs.close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	clientConn, err := Dial(ctx, fs.url(), nil)
	if err != nil {
		t.Fatalf("Dial: %v", err)
	}
	defer clientConn.CloseNow()
	serverConn := fs.accepted(t)

	one, _ := json.Marshal(Envelope{Type: TypePing})
	two, _ := json.Marshal(Envelope{Type: TypeInvocation, Target: "feed"})
	packed := append(append(append(one, recordSeparator), two...), recordSeparator)
	if err := serverConn.Write(ctx, websocket.MessageText, packed); err != nil {
		t.Fatalf("server write packed frame: %v", err)
	}

	first, err := clientConn.NextEnvelope(ctx)
	if err != nil {
		t.Fatalf("NextEnvelope (1st): %v", err)
	}
	if first.Type != TypePing {
		t.Errorf("first envelope type = %d, want TypePing (%d)", first.Type, TypePing)
	}

	second, err := clientConn.NextEnvelope(ctx)
	if err != nil {
		t.Fatalf("NextEnvelope (2nd): %v", err)
	}
	if second.Type != TypeInvocation || second.Target != "feed" {
		t.Errorf("second envelope = %+v, want a feed invocation", second)
	}
}

func TestNextEnvelope_ContextCanceled(t *testing.T) {
	fs := newFakeServer(t, "")
	defer fs.close()

	dialCtx, dialCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer dialCancel()
	clientConn, err := Dial(dialCtx, fs.url(), nil)
	if err != nil {
		t.Fatalf("Dial: %v", err)
	}
	defer clientConn.CloseNow()
	fs.accepted(t)

	readCtx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	if _, err := clientConn.NextEnvelope(readCtx); err == nil {
		t.Error("NextEnvelope returned no error on a canceled/timed-out context, want one")
	}
}
