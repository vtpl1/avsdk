package pcm

import (
	"time"

	"github.com/vtpl1/avsdk/av"
)

type SpeexCodecData struct {
	Typ           av.CodecType
	SmplFormat    av.SampleFormat
	SmplRate      int
	channelLayout av.ChannelLayout
}

func NewSpeexCodecData(sr int, cl av.ChannelLayout) av.AudioCodecData {
	return SpeexCodecData{
		Typ:           av.SPEEX,
		SmplFormat:    av.S16,
		SmplRate:      sr,
		channelLayout: cl,
	}
}

// ChannelLayout implements av.AudioCodecData.
func (m SpeexCodecData) ChannelLayout() av.ChannelLayout {
	return m.channelLayout
}

// PacketDuration implements av.AudioCodecData.
func (m SpeexCodecData) PacketDuration(_ []byte) (time.Duration, error) {
	return time.Millisecond * 20, nil
}

// SampleFormat implements av.AudioCodecData.
func (m SpeexCodecData) SampleFormat() av.SampleFormat {
	return m.SmplFormat
}

// SampleRate implements av.AudioCodecData.
func (m SpeexCodecData) SampleRate() int {
	return m.SmplRate
}

// Type implements av.AudioCodecData.
func (m SpeexCodecData) Type() av.CodecType {
	return m.Typ
}
