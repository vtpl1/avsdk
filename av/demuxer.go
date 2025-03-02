package av

type PacketReader interface {
	ReadPacket() (Packet, error)
}

// Demuxer can read compressed audio/video packets from container formats like MP4/FLV/MPEG-TS.
type Demuxer interface {
	Streams() ([]CodecData, error) // reads the header, contains video/audio meta infomations
	PacketReader                   // read compressed audio/video packets
}

// DemuxCloser exposes Demuxer with Close() method.
type DemuxCloser interface {
	Demuxer
	Close() error
}
