package av

import "context"

type PacketWriter interface {
	WritePacket(ctx context.Context, pkt Packet) error
}

type HandshakeMuxer interface {
	Handshake(ctx context.Context, codecs []CodecData, sdpIn string) (sdp string, err error)
	Muxer
}

type HandshakeMuxCloser interface {
	HandshakeMuxer
	MuxCloser
}

// Muxer describes the steps of writing compressed audio/video packets into container formats like MP4/FLV/MPEG-TS.
//
// Container formats, rtmp.Conn, and transcode.Muxer implements Muxer interface.
type Muxer interface {
	WriteHeader(ctx context.Context, codecs []CodecData) error // write the file header
	PacketWriter                                               // write compressed audio/video packets
	WriteTrailer(ctx context.Context) error                    // finish writing file, this func can be called only once
}

// MuxCloser exposes Muxer with Close() method.
type MuxCloser interface {
	Muxer
	Close() error
}
