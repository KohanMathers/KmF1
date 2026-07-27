package signalr

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"

	"github.com/coder/websocket"
)

type MessageType int

const (
	TypeInvocation       MessageType = 1
	TypeStreamItem       MessageType = 2
	TypeCompletion       MessageType = 3
	TypeStreamInvocation MessageType = 4
	TypeCancelInvocation MessageType = 5
	TypePing             MessageType = 6
	TypeClose            MessageType = 7
)

const recordSeparator = 0x1e

type Envelope struct {
	Type         MessageType       `json:"type"`
	Target       string            `json:"target,omitempty"`
	InvocationID string            `json:"invocationId,omitempty"`
	Arguments    []json.RawMessage `json:"arguments,omitempty"`
	Result       json.RawMessage   `json:"result,omitempty"`
	Error        string            `json:"error,omitempty"`
}

type Conn struct {
	ws  *websocket.Conn
	buf []byte
}

func Dial(ctx context.Context, url string, opts *websocket.DialOptions) (*Conn, error) {
	ws, _, err := websocket.Dial(ctx, url, opts)
	if err != nil {
		return nil, fmt.Errorf("signalr: dial: %w", err)
	}

	c := &Conn{ws: ws}
	if err := c.handshake(ctx); err != nil {

		ws.CloseNow()
		return nil, err
	}
	return c, nil
}

func (c *Conn) handshake(ctx context.Context) error {
	req, err := json.Marshal(map[string]any{"protocol": "json", "version": 1})
	if err != nil {
		return fmt.Errorf("signalr: encode handshake request: %w", err)
	}
	if err := c.writeFrame(ctx, req); err != nil {
		return fmt.Errorf("signalr: send handshake request: %w", err)
	}

	frame, err := c.readFrame(ctx)
	if err != nil {
		return fmt.Errorf("signalr: read handshake response: %w", err)
	}
	var resp struct {
		Error string `json:"error"`
	}
	if err := json.Unmarshal(frame, &resp); err != nil {
		return fmt.Errorf("signalr: decode handshake response: %w", err)
	}
	if resp.Error != "" {
		return fmt.Errorf("signalr: handshake rejected: %s", resp.Error)
	}
	return nil
}

func (c *Conn) Invoke(ctx context.Context, target, invocationID string, args ...any) error {
	rawArgs := make([]json.RawMessage, len(args))
	for i, a := range args {
		b, err := json.Marshal(a)
		if err != nil {
			return fmt.Errorf("signalr: encode invocation argument %d: %w", i, err)
		}
		rawArgs[i] = b
	}

	env := Envelope{Type: TypeInvocation, Target: target, InvocationID: invocationID, Arguments: rawArgs}
	b, err := json.Marshal(env)
	if err != nil {
		return fmt.Errorf("signalr: encode invocation: %w", err)
	}
	if err := c.writeFrame(ctx, b); err != nil {
		return fmt.Errorf("signalr: send invocation: %w", err)
	}
	return nil
}

func (c *Conn) NextEnvelope(ctx context.Context) (Envelope, error) {
	frame, err := c.readFrame(ctx)
	if err != nil {
		return Envelope{}, err
	}
	if len(frame) == 0 {

		return c.NextEnvelope(ctx)
	}
	var env Envelope
	if err := json.Unmarshal(frame, &env); err != nil {
		return Envelope{}, fmt.Errorf("signalr: decode envelope: %w", err)
	}
	return env, nil
}

func (c *Conn) SetReadLimit(n int64) {
	c.ws.SetReadLimit(n)
}

func (c *Conn) Close() error {
	return c.ws.Close(websocket.StatusNormalClosure, "")
}

func (c *Conn) CloseNow() error {
	return c.ws.CloseNow()
}

func (c *Conn) writeFrame(ctx context.Context, payload []byte) error {
	framed := make([]byte, 0, len(payload)+1)
	framed = append(framed, payload...)
	framed = append(framed, recordSeparator)
	return c.ws.Write(ctx, websocket.MessageText, framed)
}

func (c *Conn) readFrame(ctx context.Context) ([]byte, error) {
	for {
		if i := bytes.IndexByte(c.buf, recordSeparator); i >= 0 {
			chunk := c.buf[:i]
			c.buf = c.buf[i+1:]
			return chunk, nil
		}

		_, data, err := c.ws.Read(ctx)
		if err != nil {
			return nil, err
		}
		c.buf = append(c.buf, data...)
	}
}
