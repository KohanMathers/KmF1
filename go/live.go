package kmf1

import (
	"context"

	"github.com/KohanMathers/KmF1/go/internal/livetiming/live"
)

type (
	// LiveMessage is a single decoded message from the live SignalR timing feed, tagged with its topic and timestamp.
	LiveMessage = live.Message
	// LiveOptions configures a Live recording session (e.g. logger, reconnect behavior).
	LiveOptions = live.Options
)

// Live connects to F1's live SignalR timing feed and streams decoded messages to out until ctx is cancelled or an unrecoverable error occurs. Pair it with a LiveSession's Consume method to build up live session state from the stream.
func (c *Client) Live(ctx context.Context, out chan<- LiveMessage, opts LiveOptions) error {
	if opts.Logger == nil {
		opts.Logger = c.log()
	}
	return live.Record(ctx, out, opts)
}
