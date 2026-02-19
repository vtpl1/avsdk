package av

import (
	"context"
	"time"
)

// PacketReader defines the interface for reading compressed audio/video packets.
// A returned Packet with non-nil NewCodecs signals a mid-stream codec change for
// the listed streams; receivers must update their per-stream codec state accordingly.
type PacketReader interface {
	ReadPacket(ctx context.Context) (Packet, error)
}

// Demuxer reads compressed audio/video packets from a container (MP4, FLV, MPEG-TS, …).
//
// Lifecycle:
//  1. Call GetCodecs to obtain initial stream configuration.
//  2. Loop on ReadPacket until io.EOF or context cancellation.
//  3. Handle Packet.NewCodecs != nil for mid-stream codec changes (partial updates —
//     only the listed streams changed; all others remain as returned by GetCodecs).
type Demuxer interface {
	// GetCodecs reads the container header and returns the initial Stream list.
	// Each Stream.Idx matches Packet.Idx for that track; indices may be non-contiguous.
	// Do not infer stream identity from slice position — always use Stream.Idx.
	GetCodecs(ctx context.Context) ([]Stream, error)
	PacketReader
}

// DemuxCloser is a Demuxer whose underlying source must be closed when done.
// This is the primary type returned by source-opening functions such as avutil.Open.
// Optional capabilities are accessed by type assertion:
//
//	if p, ok := dmx.(av.Pauser);     ok { p.Pause(ctx) }
//	if s, ok := dmx.(av.TimeSeeker); ok { s.SeekToTime(ctx, 30*time.Second) }
type DemuxCloser interface {
	Demuxer
	Close() error
}

// Pauser is an optional capability a Demuxer may implement to pause and resume delivery.
// Pause stops ReadPacket from returning new packets (it blocks or returns immediately,
// depending on implementation) until Resume is called.
type Pauser interface {
	Pause(ctx context.Context) error
	Resume(ctx context.Context) error
	IsPaused() bool
}

// TimeSeeker is an optional capability a Demuxer may implement to seek within a stream.
// pos is the desired stream position as a duration from the start (matching Packet.DTS).
// The returned duration is the actual position landed on, which may differ from pos
// due to keyframe alignment or container constraints.
// The first packet after a successful seek will have IsDiscontinuity set to true.
type TimeSeeker interface {
	SeekToTime(ctx context.Context, pos time.Duration) (time.Duration, error)
}
