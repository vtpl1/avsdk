package av

import (
	"context"
)

// DemuxerFactory opens and returns a DemuxCloser for the given stream.
// streamID identifies the source (e.g. an RTSP URL, a camera ID, or a file path).
// The caller is responsible for calling Close() on the returned DemuxCloser when done.
type DemuxerFactory func(ctx context.Context, streamID string) (DemuxCloser, error)

// DemuxerRemover tears down a previously created demuxer and deregisters it from any
// internal registry. It must be called after Close() has been called on the DemuxCloser.
// streamID must match the value used when the demuxer was created.
type DemuxerRemover func(ctx context.Context, streamID string) error

// MuxerFactory opens and returns a MuxCloser for the given stream and consumer.
// streamID identifies the source stream; consumerID identifies the downstream sink
// (e.g. a recording session, a subscriber connection, or an output URL).
// The caller is responsible for calling Close() on the returned MuxCloser when done.
type MuxerFactory func(ctx context.Context, streamID, consumerID string) (MuxCloser, error)

// MuxerRemover tears down a previously created muxer and deregisters it from any
// internal registry. It must be called after Close() has been called on the MuxCloser.
// streamID and consumerID must match the values used when the muxer was created.
type MuxerRemover func(ctx context.Context, streamID, consumerID string) error
