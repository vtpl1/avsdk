// https://github.com/AlexxIT/go2rtc.git
// Package pcm
// https://www.codeproject.com/Articles/14237/Using-the-G711-standard
package pcm

import (
	"time"

	"github.com/vtpl1/avsdk/av"
)

const alawMax = 0x7FFF

func PCMAtoPCM(alaw byte) int16 {
	alaw ^= 0xD5

	data := int16(((alaw & 0x0F) << 4) + 8)
	exponent := (alaw & 0x70) >> 4

	if exponent != 0 {
		data |= 0x100
	}

	if exponent > 1 {
		data <<= exponent - 1
	}

	// sign
	if alaw&0x80 == 0 {
		return data
	} else {
		return -data
	}
}

func PCMtoPCMA(pcm int16) byte {
	var alaw byte

	if pcm < 0 {
		pcm = -pcm
		alaw = 0x80
	}

	if pcm > alawMax {
		pcm = alawMax
	}

	exponent := byte(7)
	for expMask := int16(0x4000); (pcm&expMask) == 0 && exponent > 0; expMask >>= 1 {
		exponent--
	}

	if exponent == 0 {
		alaw |= byte(pcm>>4) & 0x0F
	} else {
		alaw |= (exponent << 4) | (byte(pcm>>(exponent+3)) & 0x0F)
	}

	return alaw ^ 0xD5
}

type PCMAlawCodecData struct {
	Typ        av.CodecType
	SmplFormat av.SampleFormat
	SmplRate   int
	ChLayout   av.ChannelLayout
}

// ChannelLayout implements av.AudioCodecData.
func (m PCMAlawCodecData) ChannelLayout() av.ChannelLayout {
	return m.ChLayout
}

// PacketDuration implements av.AudioCodecData.
func (m PCMAlawCodecData) PacketDuration(pkt []byte) (time.Duration, error) {
	return time.Duration(len(pkt)) * time.Second / time.Duration(m.SampleRate()), nil
}

// SampleFormat implements av.AudioCodecData.
func (m PCMAlawCodecData) SampleFormat() av.SampleFormat {
	return m.SmplFormat
}

// SampleRate implements av.AudioCodecData.
func (m PCMAlawCodecData) SampleRate() int {
	return m.SmplRate
}

// Type implements av.AudioCodecData.
func (m PCMAlawCodecData) Type() av.CodecType {
	return m.Typ
}

func NewPCMAlawCodecData() av.AudioCodecData {
	return PCMAlawCodecData{
		Typ:        av.PCM_ALAW,
		SmplFormat: av.S16,
		SmplRate:   8000,
		ChLayout:   av.ChMono,
	}
}
