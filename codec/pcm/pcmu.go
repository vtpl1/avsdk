// Package pcm
// https://www.codeproject.com/Articles/14237/Using-the-G711-standard
package pcm

import (
	"time"

	"github.com/vtpl1/avsdk/av"
)

const (
	bias    = 0x84 // 132 or 1000 0100
	ulawMax = alawMax - bias
)

func PCMUtoPCM(ulaw byte) int16 {
	ulaw = ^ulaw

	exponent := (ulaw & 0x70) >> 4
	data := (int16((((ulaw&0x0F)|0x10)<<1)+1) << (exponent + 2)) - bias

	// sign
	//nolint:gocritic
	if ulaw&0x80 == 0 {
		return data
	} else if data == 0 {
		return -1
	} else {
		return -data
	}
}

func PCMtoPCMU(pcm int16) byte {
	var ulaw byte

	if pcm < 0 {
		pcm = -pcm
		ulaw = 0x80
	}

	if pcm > ulawMax {
		pcm = ulawMax
	}

	pcm += bias

	exponent := byte(7)
	for expMask := int16(0x4000); (pcm & expMask) == 0; expMask >>= 1 {
		exponent--
	}

	// mantisa
	ulaw |= byte(pcm>>(exponent+3)) & 0x0F

	if exponent > 0 {
		ulaw |= exponent << 4
	}

	return ^ulaw
}

type PCMMulawCodecData struct {
	Typ        av.CodecType
	SmplFormat av.SampleFormat
	SmplRate   int
	ChLayout   av.ChannelLayout
}

// ChannelLayout implements av.AudioCodecData.
func (m PCMMulawCodecData) ChannelLayout() av.ChannelLayout {
	return m.ChLayout
}

// PacketDuration implements av.AudioCodecData.
func (m PCMMulawCodecData) PacketDuration(pkt []byte) (time.Duration, error) {
	return time.Duration(len(pkt)) * time.Second / time.Duration(m.SampleRate()), nil
}

// SampleFormat implements av.AudioCodecData.
func (m PCMMulawCodecData) SampleFormat() av.SampleFormat {
	return m.SmplFormat
}

// SampleRate implements av.AudioCodecData.
func (m PCMMulawCodecData) SampleRate() int {
	return m.SmplRate
}

// Type implements av.AudioCodecData.
func (m PCMMulawCodecData) Type() av.CodecType {
	return m.Typ
}

func NewPCMMulawCodecData() av.AudioCodecData {
	return PCMMulawCodecData{
		Typ:        av.PCM_MULAW,
		SmplFormat: av.S16,
		SmplRate:   8000,
		ChLayout:   av.ChMono,
	}
}
