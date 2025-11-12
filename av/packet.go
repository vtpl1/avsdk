package av

import (
	"fmt"
	"reflect"
	"time"
)

// Packet stores compressed audio/video data.
type Packet struct {
	IsKeyFrame      bool          // true if this video packet is a keyframe
	Idx             int8          // stream index in container format
	CompositionTime time.Duration // PTS - DTS (e.g., for H.264 B-frames)
	Time            time.Duration // decode timestamp (DTS)
	Duration        time.Duration // packet duration
	Data            []byte        // raw packet data
	Extra           any           // optional extra metadata
	FrameID         int64         // unique frame identifier
	CodecType       CodecType     // codec type (H.264, H.265, etc.)
	IsParamSetNALU  bool          // true if this packet contains parameter sets (SPS/PPS/VPS)
}

// String returns a compact human-readable description of the packet.
// Suitable for logging.
func (m *Packet) String() string {
	var naluStr string
	if m.CodecType.IsVideo() {
		if len(m.Data) > 0 {
			nalu := NaluType(m.Data[0])
			naluStr = nalu.String(m.CodecType)
		} else {
			naluStr = "EMPTY"
		}
	} else {
		naluStr = "AUDIO"
	}

	return fmt.Sprintf(
		"ID=%d Time=%dms Media=%s NALU=%s Duration=%s DataLen=%d",
		m.FrameID,
		m.Time.Milliseconds(),
		m.CodecType.String(),
		naluStr,
		m.Duration,
		len(m.Data),
	)
}

// GoString returns a detailed developer-friendly representation of the packet.
// Used when printing with %#v.
func (m *Packet) GoString() string {
	var extraType string
	if m.Extra != nil {
		extraType = reflect.TypeOf(m.Extra).String()
	} else {
		extraType = "nil"
	}

	return fmt.Sprintf(
		"&av.Packet{\n"+
			"  FrameID:         %d,\n"+
			"  IsKeyFrame:      %t,\n"+
			"  Idx:             %d,\n"+
			"  CodecType:       %s,\n"+
			"  Time:            %s,\n"+
			"  CompositionTime: %s,\n"+
			"  Duration:        %s,\n"+
			"  DataLen:         %d,\n"+
			"  IsParamSetNALU:  %t,\n"+
			"  Extra:           %s (%s),\n"+
			"}",
		m.FrameID,
		m.IsKeyFrame,
		m.Idx,
		m.CodecType.String(),
		m.Time,
		m.CompositionTime,
		m.Duration,
		len(m.Data),
		m.IsParamSetNALU,
		fmt.Sprintf("%v", m.Extra),
		extraType,
	)
}
