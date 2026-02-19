package av

import (
	"fmt"
	"reflect"
	"strings"
	"time"
)

// Packet stores compressed audio/video data.
type Packet struct {
	KeyFrame        bool          // true if this video packet is a keyframe
	IsDiscontinuity bool          // DTS does not follow from the previous packet; receivers must reinitialise timing
	IsParamSetNALU  bool          // true if this packet contains parameter sets (SPS/PPS/VPS)
	Idx             uint16        // stream index in container format (matches fMP4 uint32 track IDs; uint16 covers 65535 tracks)
	DTS             time.Duration // decode timestamp; the time at which the decoder should process this packet
	PTSOffset       time.Duration // presentation offset: PTS = DTS + PTSOffset; non-zero only for B-frames (H.264/H.265)
	Duration        time.Duration // packet duration
	WallClockTime   time.Time     // wall-clock capture/arrival time (e.g. NTP for live streams); zero means unset
	Data            []byte        // raw packet data; empty for pure codec-change notification packets
	Extra           any           // optional extra metadata
	FrameID         int64         // unique frame identifier
	CodecType       CodecType     // codec type (H.264, H.265, etc.)
	NewCodecs       []Stream      // non-nil signals a mid-stream codec change; contains only the changed streams
}

// PTS returns the presentation timestamp (DTS + PTSOffset).
// For streams without B-frames, PTS == DTS and PTSOffset is zero.
func (m *Packet) PTS() time.Duration {
	return m.DTS + m.PTSOffset
}

// HasWallClockTime reports whether a wall-clock capture/arrival time has been set on this packet.
func (m *Packet) HasWallClockTime() bool {
	return !m.WallClockTime.IsZero()
}

// String returns a compact human-readable description of the packet.
// Suitable for logging.
//
// Format: #<id> <codec>:<idx> dts=<ms>ms [pts=<ms>ms] dur=<ms>ms <size> [K] [DISC] [PS] [<nalu>]
//
// Examples:
//
//	#42 H264:0 dts=1234ms dur=33ms 12.3KB K IDR_SLICE
//	#42 H264:0 dts=1234ms pts=1267ms dur=33ms 12.3KB NON_IDR_SLICE
//	#43 AAC:1 dts=1234ms dur=21ms 1.2KB
//	#44 H264:0 dts=1267ms dur=33ms 8.1KB K DISC IDR_SLICE
func (m *Packet) String() string {
	var b strings.Builder

	fmt.Fprintf(&b, "#%d %s:%d dts=%dms", m.FrameID, m.CodecType, m.Idx, m.DTS.Milliseconds())

	if m.PTSOffset != 0 {
		fmt.Fprintf(&b, " pts=%dms", m.PTS().Milliseconds())
	}

	fmt.Fprintf(&b, " dur=%dms %s", m.Duration.Milliseconds(), packetSizeString(len(m.Data)))

	if m.KeyFrame {
		b.WriteString(" K")
	}

	if m.IsDiscontinuity {
		b.WriteString(" DISC")
	}

	if m.IsParamSetNALU {
		b.WriteString(" PS")
	}

	if m.CodecType.IsVideo() && len(m.Data) > 0 {
		switch m.CodecType {
		case H264:
			fmt.Fprintf(&b, " %s", H264NaluType(m.Data[0])&H264NALTypeMask)
		case H265:
			fmt.Fprintf(&b, " %s", H265NaluType(m.Data[0]>>1)&H265NALTypeMask)
		}
	}

	return b.String()
}

// packetSizeString formats a byte count as a human-readable size string.
func packetSizeString(n int) string {
	switch {
	case n >= 1024*1024:
		return fmt.Sprintf("%.1fMB", float64(n)/(1024*1024))
	case n >= 1024:
		return fmt.Sprintf("%.1fKB", float64(n)/1024)
	default:
		return fmt.Sprintf("%dB", n)
	}
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

	wallStr := "not set"
	if m.HasWallClockTime() {
		wallStr = m.WallClockTime.String()
	}

	return fmt.Sprintf(
		"&av.Packet{\n"+
			"  FrameID:         %d,\n"+
			"  IsKeyFrame:      %t,\n"+
			"  IsDiscontinuity: %t,\n"+
			"  IsParamSetNALU:  %t,\n"+
			"  Idx:             %d,\n"+
			"  CodecType:       %s,\n"+
			"  DTS:             %s,\n"+
			"  PTS:             %s,\n"+
			"  PTSOffset:       %s,\n"+
			"  Duration:        %s,\n"+
			"  WallClockTime:   %s,\n"+
			"  DataLen:         %d,\n"+
			"  Extra:           %s (%s),\n"+
			"}",
		m.FrameID,
		m.KeyFrame,
		m.IsDiscontinuity,
		m.IsParamSetNALU,
		m.Idx,
		m.CodecType.String(),
		m.DTS,
		m.PTS(),
		m.PTSOffset,
		m.Duration,
		wallStr,
		len(m.Data),
		fmt.Sprintf("%v", m.Extra),
		extraType,
	)
}

func (m *Packet) IsKeyFrame() bool {
	return m.KeyFrame
}

func (m *Packet) IsAudio() bool {
	return m.CodecType.IsAudio()
}

func (m *Packet) IsVideo() bool {
	return m.CodecType.IsVideo()
}
