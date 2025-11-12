package av

import "time"

// CodecType represents Video/Audio codec type. can be H264/AAC/SPEEX/...
type CodecType uint32

const (
	codecTypeAudioBit  = 0x1
	codecTypeOtherBits = 1
	avCodecTypeMagic   = 233333
)

// MakeVideoCodecType makes a new video codec type.
func MakeVideoCodecType(base uint32) CodecType {
	c := CodecType(base) << codecTypeOtherBits

	return c
}

// MakeAudioCodecType makes a new audio codec type.
func MakeAudioCodecType(base uint32) CodecType {
	c := CodecType(base)<<codecTypeOtherBits | CodecType(codecTypeAudioBit)

	return c
}

var (
	UNKNOWN    = MakeVideoCodecType(avCodecTypeMagic + 0)  //nolint:gochecknoglobals
	H264       = MakeVideoCodecType(avCodecTypeMagic + 1)  //nolint:gochecknoglobals	// payloadType: 96
	H265       = MakeVideoCodecType(avCodecTypeMagic + 2)  //nolint:gochecknoglobals
	JPEG       = MakeVideoCodecType(avCodecTypeMagic + 3)  //nolint:gochecknoglobals	// payloadType: 26
	VP8        = MakeVideoCodecType(avCodecTypeMagic + 4)  //nolint:gochecknoglobals
	VP9        = MakeVideoCodecType(avCodecTypeMagic + 5)  //nolint:gochecknoglobals
	AV1        = MakeVideoCodecType(avCodecTypeMagic + 6)  //nolint:gochecknoglobals
	MJPEG      = MakeVideoCodecType(avCodecTypeMagic + 7)  //nolint:gochecknoglobals
	AAC        = MakeAudioCodecType(avCodecTypeMagic + 1)  //nolint:gochecknoglobals	// MPEG4-GENERIC
	PCM_MULAW  = MakeAudioCodecType(avCodecTypeMagic + 2)  //nolint:gochecknoglobals,revive,stylecheck	// payloadType: 0
	PCM_ALAW   = MakeAudioCodecType(avCodecTypeMagic + 3)  //nolint:gochecknoglobals,revive,stylecheck	// payloadType: 8
	SPEEX      = MakeAudioCodecType(avCodecTypeMagic + 4)  //nolint:gochecknoglobals	// L16 Linear PCM (big endian)
	NELLYMOSER = MakeAudioCodecType(avCodecTypeMagic + 5)  //nolint:gochecknoglobals
	PCM        = MakeAudioCodecType(avCodecTypeMagic + 6)  //nolint:gochecknoglobals
	OPUS       = MakeAudioCodecType(avCodecTypeMagic + 7)  //nolint:gochecknoglobals	// payloadType: 111
	MP3        = MakeAudioCodecType(avCodecTypeMagic + 8)  //nolint:gochecknoglobals	// MPA payload: 14, aka MPEG-1 Layer III
	PCML       = MakeAudioCodecType(avCodecTypeMagic + 9)  //nolint:gochecknoglobals	// Linear PCM (little endian)
	ELD        = MakeAudioCodecType(avCodecTypeMagic + 10) //nolint:gochecknoglobals	// AAC-ELD
	FLAC       = MakeAudioCodecType(avCodecTypeMagic + 11) //nolint:gochecknoglobals
)

func (s CodecType) String() string {
	switch s {
	case H264:
		return "H264"
	case H265:
		return "H265"
	case JPEG:
		return "JPEG"
	case VP8:
		return "VP8"
	case VP9:
		return "VP9"
	case AV1:
		return "AV1"
	case AAC:
		return "AAC"
	case PCM_MULAW:
		return "PCM_MULAW"
	case PCM_ALAW:
		return "PCM_ALAW"
	case SPEEX:
		return "SPEEX"
	case NELLYMOSER:
		return "NELLYMOSER"
	case PCM:
		return "PCM"
	case OPUS:
		return "OPUS"
	case MP3:
		return "MPA"
	case PCML:
		return "PCML"
	case ELD:
		return "AAC_ELD"
	case FLAC:
		return "FLAC"
	}

	return ""
}

func (s CodecType) IsAudio() bool {
	return s&codecTypeAudioBit != 0
}

func (s CodecType) IsVideo() bool {
	return s&codecTypeAudioBit == 0
}

// CodecData is some important bytes for initialising audio/video decoder,
// can be converted to VideoCodecData or AudioCodecData using:
//
//	codecdata.(AudioCodecData) or codecdata.(VideoCodecData)
//
// for H264, CodecData is AVCDecoderConfigure bytes, includes SPS/PPS.
// for H265, CodecData is AVCDecoderConfigure bytes, includes VPS/SPS/PPS.
type CodecData interface {
	Type() CodecType // Video/Audio codec type
}

type VideoCodecData interface {
	CodecData
	Width() int  // Video width
	Height() int // Video height
}

type AudioCodecData interface {
	CodecData
	SampleFormat() SampleFormat                       // audio sample format
	SampleRate() int                                  // audio sample rate
	ChannelLayout() ChannelLayout                     // audio channel layout
	PacketDuration(pkt []byte) (time.Duration, error) // get audio compressed packet duration
}
