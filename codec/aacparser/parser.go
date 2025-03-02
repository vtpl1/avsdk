// Package aacparser holds Muxer and Demuxer for aac
package aacparser

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/vtpl1/avsdk/av"
	"github.com/vtpl1/avsdk/utils/bits"
)

var (
	ErrAACparserNotAdtsHeader           = errors.New("aacparser: not adts header")
	ErrAACparserAdtsChannelCountInvalid = errors.New("aacparser: adts channel count invalid")
	ErrAACparserAdtsFrameLen            = errors.New("aacparser: adts framelen < hdrlen")
	ErrAACparserMPEG4AudioConfigFailed  = errors.New("aacparser: parse MPEG4AudioConfig failed")
)

// copied from libavcodec/mpeg4audio.h.
const (
	AotAacMain       = 1 + iota  ///< Y                       Main
	AotAacLc                     ///< Y                       Low Complexity
	AotAacSsr                    ///< N (code in SoC repo)    Scalable Sample Rate
	AotAacLtp                    ///< Y                       Long Term Prediction
	AotSbr                       ///< Y                       Spectral Band Replication
	AotAacScalable               ///< N                       Scalable
	AotTwinvq                    ///< N                       Twin Vector Quantizer
	AotCelp                      ///< N                       Code Excited Linear Prediction
	AotHvxc                      ///< N                       Harmonic Vector eXcitation Coding
	AotTtsi          = 12 + iota ///< N                       Text-To-Speech Interface
	AotMainsynth                 ///< N                       Main Synthesis
	AotWavesynth                 ///< N                       Wavetable Synthesis
	AotMidi                      ///< N                       General MIDI
	AotSafx                      ///< N                       Algorithmic Synthesis and Audio Effects
	AotErAacLc                   ///< N                       Error Resilient Low Complexity
	AotErAacLtp      = 19 + iota ///< N                       Error Resilient Long Term Prediction
	AotErAacScalable             ///< N                       Error Resilient Scalable
	AotErTwinvq                  ///< N                       Error Resilient Twin Vector Quantizer
	AotErBsac                    ///< N                       Error Resilient Bit-Sliced Arithmetic Coding
	AotErAacLd                   ///< N                       Error Resilient Low Delay
	AotErCelp                    ///< N                       Error Resilient Code Excited Linear Prediction
	AotErHvxc                    ///< N                       Error Resilient Harmonic Vector eXcitation Coding
	AotErHiln                    ///< N                       Error Resilient Harmonic and Individual Lines plus Noise
	AotErParam                   ///< N                       Error Resilient Parametric
	AotSsc                       ///< N                       SinuSoidal Coding
	AotPs                        ///< N                       Parametric Stereo
	AotSurround                  ///< N                       MPEG Surround
	AotEscape                    ///< Y                       Escape Value
	AotL1                        ///< Y                       Layer 1
	AotL2                        ///< Y                       Layer 2
	AotL3                        ///< Y                       Layer 3
	AotDst                       ///< N                       Direct Stream Transfer
	AotAls                       ///< Y                       Audio LosslesS
	AotSls                       ///< N                       Scalable LosslesS
	AotSlsNonCore                ///< N                       Scalable LosslesS (non core)
	AotErAacEld                  ///< N                       Error Resilient Enhanced Low Delay
	AotSmrSimple                 ///< N                       Symbolic Music Representation Simple
	AotSmrMain                   ///< N                       Symbolic Music Representation Main
	AotUsacNosbr                 ///< N                       Unified Speech and Audio Coding (no SBR)
	AotSaoc                      ///< N                       Spatial Audio Object Coding
	AotLdSurround                ///< N                       Low Delay MPEG Surround
	AotUsac                      ///< N                       Unified Speech and Audio Coding
)

type MPEG4AudioConfig struct {
	SampleRate      int
	ChannelLayout   av.ChannelLayout
	ObjectType      uint
	SampleRateIndex uint
	ChannelConfig   uint
}

//nolint:gochecknoglobals
var sampleRateTable = []int{
	96000, 88200, 64000, 48000, 44100, 32000,
	24000, 22050, 16000, 12000, 11025, 8000, 7350,
}

/*
These are the channel configurations:
0: Defined in AOT Specifc Config
1: 1 channel: front-center
2: 2 channels: front-left, front-right
3: 3 channels: front-center, front-left, front-right
4: 4 channels: front-center, front-left, front-right, back-center
5: 5 channels: front-center, front-left, front-right, back-left, back-right
6: 6 channels: front-center, front-left, front-right, back-left, back-right, LFE-channel
7: 8 channels: front-center, front-left, front-right, side-left, side-right, back-left, back-right, LFE-channel
8-15: Reserved
*/
//nolint:gochecknoglobals
var chanConfigTable = []av.ChannelLayout{
	0,
	av.ChFrontCenter,
	av.ChFrontLeft | av.ChFrontRight,
	av.ChFrontCenter | av.ChFrontLeft | av.ChFrontRight,
	av.ChFrontCenter | av.ChFrontLeft | av.ChFrontRight | av.ChBackCenter,
	av.ChFrontCenter | av.ChFrontLeft | av.ChFrontRight | av.ChBackLeft | av.ChBackRight,
	av.ChFrontCenter | av.ChFrontLeft | av.ChFrontRight | av.ChBackLeft | av.ChBackRight | av.ChLowFreq,
	av.ChFrontCenter | av.ChFrontLeft | av.ChFrontRight | av.ChSideLeft | av.ChSideRight | av.ChBackLeft | av.ChBackRight | av.ChLowFreq,
}

//nolint:nonamedreturns
func ParseADTSHeader(frame []byte) (config MPEG4AudioConfig, hdrlen int, framelen int, samples int, err error) {
	if frame[0] != 0xff || frame[1]&0xf6 != 0xf0 {
		err = ErrAACparserNotAdtsHeader

		return config, hdrlen, framelen, samples, err
	}

	config.ObjectType = uint(frame[2]>>6) + 1
	config.SampleRateIndex = uint(frame[2] >> 2 & 0xf)
	config.ChannelConfig = uint(frame[2]<<2&0x4 | frame[3]>>6&0x3)

	if config.ChannelConfig == uint(0) {
		err = ErrAACparserAdtsChannelCountInvalid

		return config, hdrlen, framelen, samples, err
	}

	(&config).Complete()

	framelen = int(frame[3]&0x3)<<11 | int(frame[4])<<3 | int(frame[5]>>5)
	samples = (int(frame[6]&0x3) + 1) * 1024

	hdrlen = 7
	if frame[1]&0x1 == 0 {
		hdrlen = 9
	}

	if framelen < hdrlen {
		err = ErrAACparserAdtsFrameLen

		return config, hdrlen, framelen, samples, err
	}

	return config, hdrlen, framelen, samples, err
}

const ADTSHeaderLength = 7

func FillADTSHeader(header []byte, config MPEG4AudioConfig, samples int, payloadLength int) {
	payloadLength += 7
	// AAAAAAAA AAAABCCD EEFFFFGH HHIJKLMM MMMMMMMM MMMOOOOO OOOOOOPP (QQQQQQQQ QQQQQQQQ)
	header[0] = 0xff
	header[1] = 0xf1
	header[2] = 0x50
	header[3] = 0x80
	header[4] = 0x43
	header[5] = 0xff
	header[6] = 0xcd
	// config.ObjectType = uint(frames[2]>>6)+1
	// config.SampleRateIndex = uint(frames[2]>>2&0xf)
	// config.ChannelConfig = uint(frames[2]<<2&0x4|frames[3]>>6&0x3)
	header[2] = (byte(config.ObjectType-1)&0x3)<<6 | (byte(config.SampleRateIndex)&0xf)<<2 | byte(config.ChannelConfig>>2)&0x1
	header[3] = header[3]&0x3f | byte(config.ChannelConfig&0x3)<<6
	header[3] = header[3]&0xfc | byte(payloadLength>>11)&0x3
	header[4] = byte(payloadLength >> 3)
	header[5] = header[5]&0x1f | (byte(payloadLength)&0x7)<<5
	header[6] = header[6]&0xfc | byte(samples/1024-1)
}

func readObjectType(r *bits.Reader) (uint, error) {
	var objectType uint

	var err error
	if objectType, err = r.ReadBits(5); err != nil {
		return objectType, err
	}

	if objectType == AotEscape {
		var i uint

		if i, err = r.ReadBits(6); err != nil {
			return objectType, err
		}

		objectType = 32 + i
	}

	return objectType, err
}

func writeObjectType(w *bits.Writer, objectType uint) error {
	if objectType >= 32 {
		if err := w.WriteBits(AotEscape, 5); err != nil {
			return err
		}

		if err := w.WriteBits(objectType-32, 6); err != nil {
			return err
		}
	} else {
		if err := w.WriteBits(objectType, 5); err != nil {
			return err
		}
	}

	return nil
}

func readSampleRateIndex(r *bits.Reader) (uint, error) {
	var index uint

	var err error
	if index, err = r.ReadBits(4); err != nil {
		return index, err
	}

	if index == 0xf {
		if index, err = r.ReadBits(24); err != nil {
			return index, err
		}
	}

	return index, err
}

func writeSampleRateIndex(w *bits.Writer, index uint) error {
	if index >= 0xf {
		if err := w.WriteBits(0xf, 4); err != nil {
			return err
		}

		if err := w.WriteBits(index, 24); err != nil {
			return err
		}
	} else {
		if err := w.WriteBits(index, 4); err != nil {
			return err
		}
	}

	return nil
}

func (s *MPEG4AudioConfig) IsValid() bool {
	return s.ObjectType > 0
}

func (s *MPEG4AudioConfig) Complete() {
	if int(s.SampleRateIndex) < len(sampleRateTable) {
		s.SampleRate = sampleRateTable[s.SampleRateIndex]
	}

	if int(s.ChannelConfig) < len(chanConfigTable) {
		s.ChannelLayout = chanConfigTable[s.ChannelConfig]
	}
}

func ParseMPEG4AudioConfigBytes(data []byte) (MPEG4AudioConfig, error) {
	var config MPEG4AudioConfig

	var err error
	// copied from libavcodec/mpeg4audio.c avpriv_mpeg4audio_get_config()
	r := bytes.NewReader(data)

	br := &bits.Reader{R: r}
	if config.ObjectType, err = readObjectType(br); err != nil {
		return config, err
	}

	if config.SampleRateIndex, err = readSampleRateIndex(br); err != nil {
		return config, err
	}

	if config.ChannelConfig, err = br.ReadBits(4); err != nil {
		return config, err
	}

	(&config).Complete()

	return config, err
}

func WriteMPEG4AudioConfig(w io.Writer, config MPEG4AudioConfig) error {
	bw := &bits.Writer{W: w}
	if err := writeObjectType(bw, config.ObjectType); err != nil {
		return err
	}

	if config.SampleRateIndex == 0 {
		for i, rate := range sampleRateTable {
			if rate == config.SampleRate {
				config.SampleRateIndex = uint(i)
			}
		}
	}

	if err := writeSampleRateIndex(bw, config.SampleRateIndex); err != nil {
		return err
	}

	if config.ChannelConfig == 0 {
		for i, layout := range chanConfigTable {
			if layout == config.ChannelLayout {
				config.ChannelConfig = uint(i)
			}
		}
	}

	if err := bw.WriteBits(config.ChannelConfig, 4); err != nil {
		return err
	}

	if err := bw.FlushBits(); err != nil {
		return err
	}

	return nil
}

type CodecData struct {
	ConfigBytes []byte
	Config      MPEG4AudioConfig
}

func (s CodecData) Type() av.CodecType {
	return av.AAC
}

func (s CodecData) MPEG4AudioConfigBytes() []byte {
	return s.ConfigBytes
}

func (s CodecData) ChannelLayout() av.ChannelLayout {
	return s.Config.ChannelLayout
}

func (s CodecData) SampleRate() int {
	return s.Config.SampleRate
}

func (s CodecData) SampleFormat() av.SampleFormat {
	return av.FLTP
}

func (s CodecData) Tag() string {
	return fmt.Sprintf("mp4a.40.%d", s.Config.ObjectType)
}

func (s CodecData) PacketDuration(_ []byte) (time.Duration, error) {
	return time.Duration(1024) * time.Second / time.Duration(s.Config.SampleRate), nil
}

func NewCodecDataFromMPEG4AudioConfig(config MPEG4AudioConfig) (CodecData, error) {
	b := &bytes.Buffer{}
	_ = WriteMPEG4AudioConfig(b, config)

	return NewCodecDataFromMPEG4AudioConfigBytes(b.Bytes())
}

func NewCodecDataFromMPEG4AudioConfigBytes(config []byte) (CodecData, error) {
	var s CodecData

	var err error

	s.ConfigBytes = config

	if s.Config, err = ParseMPEG4AudioConfigBytes(config); err != nil {
		err = ErrAACparserMPEG4AudioConfigFailed

		return s, err
	}

	return s, err
}
