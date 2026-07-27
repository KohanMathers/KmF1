package kmf1

import "github.com/KohanMathers/KmF1/go/internal/livetiming/feeds"

// Broadcast holds the audio and content streams advertised for a session, obtained via Session.Broadcast or LiveSession.Broadcast.
type Broadcast struct {
	audio   []feeds.AudioStream
	content []feeds.ContentStream
}

// Audio returns the available commentary/team-radio audio streams.
func (b *Broadcast) Audio() []feeds.AudioStream { return b.audio }

// Content returns the available video content streams (e.g. driver onboard feeds).
func (b *Broadcast) Content() []feeds.ContentStream { return b.content }
