package av

import "time"

type HeaderPacket struct {
	VPS []byte
	SPS []byte
	PPS []byte
}

// Packet stores compressed audio/video data.
type Packet struct {
	IsKeyFrame      bool          // video packet is key frame
	Idx             int8          // stream index in container format
	CompositionTime time.Duration // packet presentation time minus decode time for H264 B-Frame
	Time            time.Duration // packet decode time
	Duration        time.Duration // packet duration
	Data            []byte        // packet data
}
