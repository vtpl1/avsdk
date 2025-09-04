package av

import (
	"context"
	"time"
)

// PacketReader defines the interface for reading compressed audio/video packets.
type PacketReader interface {
	ReadPacket(ctx context.Context) (Packet, error)
}

// Demuxer can read compressed audio/video packets from container formats like MP4/FLV/MPEG-TS.
type Demuxer interface {
	Streams(ctx context.Context) ([]CodecData, error) // Reads the header and returns video/audio stream info
	PacketReader
}

// DemuxCloser is a Demuxer that also supports closing the underlying source.
type DemuxCloser interface {
	Demuxer
	Close() error
}

// Pauser allows pausing/resuming demuxing.
type Pauser interface {
	Pause(pause bool)
	IsPaused() bool
}

// TimeSeeker allows seeking to a specific timestamp.
type TimeSeeker interface {
	TimeSeek(ctx context.Context, seekTime time.Time) (time.Time, error)
}

// DemuxPauser is a Demuxer with pause functionality.
type DemuxPauser interface {
	Demuxer
	Pauser
}

// DemuxPauseCloser is a Demuxer with pause and close functionality.
type DemuxPauseCloser interface {
	DemuxCloser
	Pauser
}

// DemuxPauseTimeSeeker is a Demuxer with pause and seek functionality.
type DemuxPauseTimeSeeker interface {
	DemuxPauser
	TimeSeeker
}

// DemuxPauseTimeSeekCloser is a full-featured demuxer supporting pause, seek, and close.
type DemuxPauseTimeSeekCloser interface {
	DemuxPauseCloser
	TimeSeeker
}
