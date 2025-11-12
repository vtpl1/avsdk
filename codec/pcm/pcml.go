package pcm

import (
	"time"

	"github.com/vtpl1/avsdk/av"
)

type PCMLCodecData struct {
	Typ        av.CodecType
	SmplFormat av.SampleFormat
	SmplRate   int
	ChLayout   av.ChannelLayout
}

// ChannelLayout implements av.AudioCodecData.
func (m PCMLCodecData) ChannelLayout() av.ChannelLayout {
	return m.ChLayout
}

// PacketDuration implements av.AudioCodecData.
func (m PCMLCodecData) PacketDuration(pkt []byte) (time.Duration, error) {
	return time.Duration(len(pkt)) * time.Second / time.Duration(m.SampleRate()), nil
}

// SampleFormat implements av.AudioCodecData.
func (m PCMLCodecData) SampleFormat() av.SampleFormat {
	return m.SmplFormat
}

// SampleRate implements av.AudioCodecData.
func (m PCMLCodecData) SampleRate() int {
	return m.SmplRate
}

// Type implements av.AudioCodecData.
func (m PCMLCodecData) Type() av.CodecType {
	return m.Typ
}

func NewPCMLCodecData() av.AudioCodecData {
	return PCMLCodecData{
		Typ:        av.PCML,
		SmplFormat: av.S16,
		SmplRate:   8000,
		ChLayout:   av.ChMono,
	}
}
