package av

import "context"

// PacketWriter is the write half of a media pipeline stage.
// Implementations must be safe to call from a single goroutine.
type PacketWriter interface {
	WritePacket(ctx context.Context, pkt Packet) error
}

// Muxer writes compressed audio/video packets into a container or transport.
//
// Lifecycle — callers must follow this order exactly:
//  1. WriteHeader — declares all streams; must be called exactly once before any WritePacket.
//  2. WritePacket — called repeatedly for each compressed packet.
//  3. WriteTrailer — finalises the container; must be called exactly once.
//     A second call must return an error.
//
// Optional capabilities are accessed by type assertion:
//
//	if cc, ok := mux.(av.CodecChanger); ok { cc.WriteCodecChange(ctx, changed) }
type Muxer interface {
	WriteHeader(ctx context.Context, streams []Stream) error
	PacketWriter
	WriteTrailer(ctx context.Context) error
}

// MuxCloser is a Muxer whose underlying resource must be released when done.
// This is the primary type returned by sink-opening functions.
type MuxCloser interface {
	Muxer
	Close() error
}

// CodecChanger is an optional capability a Muxer may implement to handle mid-stream
// codec changes. When a Packet carries non-nil NewCodecs, CopyPackets forwards only
// the changed Stream entries here. Unchanged streams are unaffected.
// Muxers that do not implement CodecChanger silently ignore codec-change packets.
type CodecChanger interface {
	WriteCodecChange(ctx context.Context, changed []Stream) error
}
