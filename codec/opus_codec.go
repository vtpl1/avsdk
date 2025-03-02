// Package codec returns codecs from sdp
package codec

import (
	"time"

	"github.com/vtpl1/avsdk/av"
)

type OpusCodecData struct {
	typ      av.CodecType
	SampleRt int
	ChLayout av.ChannelLayout
}

func NewOpusCodecData(sr int, cc av.ChannelLayout) av.AudioCodecData {
	return OpusCodecData{
		typ:      av.OPUS,
		SampleRt: sr,
		ChLayout: cc,
	}
}

// ChannelLayout implements av.AudioCodecData.
func (s OpusCodecData) ChannelLayout() av.ChannelLayout {
	var a av.ChannelLayout

	return a
}

// PacketDuration implements av.AudioCodecData.
func (s OpusCodecData) PacketDuration(_ []byte) (time.Duration, error) {
	return time.Duration(20) * time.Millisecond, nil
}

// SampleFormat implements av.AudioCodecData.
func (s OpusCodecData) SampleFormat() av.SampleFormat {
	return av.FLT
}

// SampleRate implements av.AudioCodecData.
func (s OpusCodecData) SampleRate() int {
	var a int

	return a
}

func (s OpusCodecData) Type() av.CodecType {
	return av.OPUS
}
