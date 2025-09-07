package av

import (
	"fmt"
	"time"
)

// SampleFormat represents Audio sample format.
type SampleFormat uint8

const (
	U8   = SampleFormat(iota + 1) // 8-bit unsigned integer
	S16                           // signed 16-bit integer
	S32                           // signed 32-bit integer
	FLT                           // 32-bit float
	DBL                           // 64-bit float
	U8P                           // 8-bit unsigned integer in planar
	S16P                          // signed 16-bit integer in planar
	S32P                          // signed 32-bit integer in planar
	FLTP                          // 32-bit float in planar
	DBLP                          // 64-bit float in planar
	U32                           // unsigned 32-bit integer
)

func (s SampleFormat) BytesPerSample() int {
	switch s {
	case U8, U8P:
		return 1
	case S16, S16P:
		return 2
	case FLT, FLTP, S32, S32P, U32:
		return 4
	case DBL, DBLP:
		return 8
	default:
		return 0
	}
}

func (s SampleFormat) String() string {
	switch s {
	case U8:
		return "U8"
	case S16:
		return "S16"
	case S32:
		return "S32"
	case FLT:
		return "FLT"
	case DBL:
		return "DBL"
	case U8P:
		return "U8P"
	case S16P:
		return "S16P"
	case S32P:
		return "S32P"
	case FLTP:
		return "FLTP"
	case DBLP:
		return "DBLP"
	case U32:
		return "U32"
	default:
		return "?"
	}
}

// IsPlanar Check if this sample format is in planar.
func (s SampleFormat) IsPlanar() bool {
	switch s { //nolint:exhaustive
	case S16P, S32P, FLTP, DBLP:
		return true
	}

	return false
}

// ChannelLayout represents Audio channel layout.
type ChannelLayout uint16

func (s ChannelLayout) String() string {
	return fmt.Sprintf("%dch", s.Count())
}

func (s ChannelLayout) Count() int {
	var n int
	for s != 0 {
		n++
		s = (s - 1) & s
	}

	return n
}

const (
	ChFrontCenter = ChannelLayout(1 << iota)
	ChFrontLeft
	ChFrontRight
	ChBackCenter
	ChBackLeft
	ChBackRight
	ChSideLeft
	ChSideRight
	ChLowFreq
	ChNr

	ChMono     = ChFrontCenter
	ChStereo   = ChFrontLeft | ChFrontRight
	Ch2_1      = ChStereo | ChBackCenter
	Ch2Point1  = ChStereo | ChLowFreq
	ChSurround = ChStereo | ChFrontCenter
	Ch3Point1  = ChSurround | ChLowFreq
)

// Raw audio frame.
type AudioFrame struct {
	SampleFormat  SampleFormat  // audio sample format, e.g: S16,FLTP,...
	ChannelLayout ChannelLayout // audio channel layout, e.g: CH_MONO,CH_STEREO,...
	SampleCount   int           // sample count in this frame
	SampleRate    int           // sample rate
	Data          [][]byte      // data array for planar format len(Data) > 1
}

func (s AudioFrame) Duration() time.Duration {
	return time.Second * time.Duration(s.SampleCount) / time.Duration(s.SampleRate)
}

// Check this audio frame has same format as other audio frame.
func (s AudioFrame) HasSameFormat(other AudioFrame) bool {
	if s.SampleRate != other.SampleRate {
		return false
	}
	if s.ChannelLayout != other.ChannelLayout {
		return false
	}
	if s.SampleFormat != other.SampleFormat {
		return false
	}

	return true
}

// Split sample audio sample from this frame.
func (s AudioFrame) Slice(start int, end int) AudioFrame {
	if start > end {
		panic(fmt.Sprintf("av: AudioFrame split failed start=%d end=%d invalid", start, end))
	}
	out := s
	out.Data = append([][]byte(nil), out.Data...)
	out.SampleCount = end - start
	size := s.SampleFormat.BytesPerSample()
	for i := range out.Data {
		out.Data[i] = out.Data[i][start*size : end*size]
	}

	return out
}

// Concat two audio frames.
func (s AudioFrame) Concat(in AudioFrame) AudioFrame {
	out := s
	out.Data = append([][]byte(nil), out.Data...)
	out.SampleCount += in.SampleCount
	for i := range out.Data {
		out.Data[i] = append(out.Data[i], in.Data[i]...)
	}

	return out
}

// AudioEncoder can encode raw audio frame into compressed audio packets.
// cgo/ffmpeg inplements AudioEncoder, using ffmpeg.NewAudioEncoder to create it.
type AudioEncoder interface {
	CodecData() (AudioCodecData, error)                 // encoder's codec data can put into container
	Encode(frame AudioFrame) ([][]byte, error)          // encode raw audio frame into compressed pakcet(s)
	Close()                                             // close encoder, free cgo contexts
	SetSampleRate(sampleRate int) error                 // set encoder sample rate
	SetChannelLayout(channelLayout ChannelLayout) error // set encoder channel layout
	SetSampleFormat(sampleFormat SampleFormat) error    // set encoder sample format
	SetBitrate(bitrate int) error                       // set encoder bitrate
	SetOption(key string, option interface{}) error     // encoder setopt, in ffmpeg is av_opt_set_dict()
	GetOption(key string, option interface{}) error     // encoder getopt
}

// AudioDecoder can decode compressed audio packets into raw audio frame.
// use ffmpeg.NewAudioDecoder to create it.
type AudioDecoder interface {
	Decode(data []byte) (bool, AudioFrame, error) // decode one compressed audio packet
	Close()                                       // close decode, free cgo contexts
}

// AudioResampler can convert raw audio frames in different sample rate/format/channel layout.
type AudioResampler interface {
	Resample(frame AudioFrame) (AudioFrame, error) // convert raw audio frames
}
