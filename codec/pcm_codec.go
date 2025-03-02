package codec

import (
	"time"

	"github.com/vtpl1/avsdk/av"
)

type PCMUCodecData struct {
	typ av.CodecType
}

func NewPCMMulawCodecData() av.AudioCodecData {
	return PCMUCodecData{
		typ: av.PCM_MULAW,
	}
}

func NewPCMCodecData() av.AudioCodecData {
	return PCMUCodecData{
		typ: av.PCM,
	}
}

func NewPCMAlawCodecData() av.AudioCodecData {
	return PCMUCodecData{
		typ: av.PCM_ALAW,
	}
}

// ChannelLayout implements av.AudioCodecData.
func (p PCMUCodecData) ChannelLayout() av.ChannelLayout {
	return av.ChMono
}

// PacketDuration implements av.AudioCodecData.
func (p PCMUCodecData) PacketDuration(pkt []byte) (time.Duration, error) {
	return time.Duration(len(pkt)) * time.Second / time.Duration(p.SampleRate()), nil
}

// SampleFormat implements av.AudioCodecData.
func (p PCMUCodecData) SampleFormat() av.SampleFormat {
	return av.S16
}

// SampleRate implements av.AudioCodecData.
func (p PCMUCodecData) SampleRate() int {
	return 8000
}

// Type implements av.AudioCodecData.
func (p PCMUCodecData) Type() av.CodecType {
	return av.PCM_MULAW
}

type SpeexCodecData struct {
	typ           av.CodecType
	sampleFormat  av.SampleFormat
	sampleRate    int
	channelLayout av.ChannelLayout
}

func NewSpeexCodecData(sr int, cl av.ChannelLayout) av.AudioCodecData {
	return SpeexCodecData{
		typ:           av.SPEEX,
		sampleFormat:  av.S16,
		sampleRate:    sr,
		channelLayout: cl,
	}
}

// ChannelLayout implements av.AudioCodecData.
func (s SpeexCodecData) ChannelLayout() av.ChannelLayout {
	return s.channelLayout
}

// PacketDuration implements av.AudioCodecData.
func (s SpeexCodecData) PacketDuration(_ []byte) (time.Duration, error) {
	return time.Millisecond * 20, nil
}

// SampleFormat implements av.AudioCodecData.
func (s SpeexCodecData) SampleFormat() av.SampleFormat {
	return s.sampleFormat
}

// SampleRate implements av.AudioCodecData.
func (s SpeexCodecData) SampleRate() int {
	return s.sampleRate
}

// Type implements av.AudioCodecData.
func (s SpeexCodecData) Type() av.CodecType {
	return s.typ
}
