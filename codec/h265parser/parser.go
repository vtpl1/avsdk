// Package h265parser holds Muxer and Demuxer for h265
package h265parser

import (
	"bytes"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/vtpl1/avsdk/av"
	"github.com/vtpl1/avsdk/codec/parser"
	"github.com/vtpl1/avsdk/utils/bits"
	"github.com/vtpl1/avsdk/utils/bits/pio"
)

type NaluType byte

// HEVC NALU types according to ISO/IEC 23008-2 Table 7.1.
const (
	HEVC_NAL_TRAIL_N NaluType = iota // TRAIL_N
	HEVC_NAL_TRAIL_R                 // TRAIL_R
	HEVC_NAL_TSA_N                   // TSA_N
	HEVC_NAL_TSA_R                   // TSA_R
	HEVC_NAL_STSA_N                  // STSA_N
	HEVC_NAL_STSA_R                  // STSA_R
	HEVC_NAL_RADL_N                  // RADL_N
	HEVC_NAL_RADL_R                  // RADL_R
	HEVC_NAL_RASL_N                  // RASL_N
	HEVC_NAL_RASL_R                  // RASL_R
	// Unused.
	HEVC_NAL_VCL_N10 // VCL_N10
	HEVC_NAL_VCL_R11 // VCL_R11
	HEVC_NAL_VCL_N12 // VCL_N12
	HEVC_NAL_VCL_R13 // VCL_R13
	HEVC_NAL_VCL_N14 // VCL_N14
	HEVC_NAL_VCL_R15 // VCL_R15
	// BLA_W_LP and the following types are Random Access.
	HEVC_NAL_BLA_W_LP   // BLA_W_LP
	HEVC_NAL_BLA_W_RADL // BLA_W_RADL
	HEVC_NAL_BLA_N_LP   // BLA_N_LP
	HEVC_NAL_IDR_W_RADL // IDR_W_RADL
	HEVC_NAL_IDR_N_LP   // IDR_N_LP
	HEVC_NAL_CRA_NUT    // CRA_NUT
	// Reserved IRAP VCL NAL unit types.
	HEVC_NAL_RSV_IRAP_VCL22 // RSV_IRAP_VCL22
	HEVC_NAL_RSV_IRAP_VCL23 // RSV_IRAP_VCL23
	// Unused.
	HEVC_NAL_RSV_VCL24 // RSV_VCL24
	HEVC_NAL_RSV_VCL25 // RSV_VCL25
	HEVC_NAL_RSV_VCL26 // RSV_VCL26
	HEVC_NAL_RSV_VCL27 // RSV_VCL27
	HEVC_NAL_RSV_VCL28 // RSV_VCL28
	HEVC_NAL_RSV_VCL29 // RSV_VCL29
	HEVC_NAL_RSV_VCL30 // RSV_VCL30
	HEVC_NAL_RSV_VCL31 // RSV_VCL31
	// NALU_VPS - VideoParameterSet NAL Unit.
	HEVC_NAL_VPS // VPS
	// NALU_SPS - SequenceParameterSet NAL Unit.
	HEVC_NAL_SPS // SPS
	// NALU_PPS - PictureParameterSet NAL Unit.
	HEVC_NAL_PPS // PPS
	// NALU_AUD - AccessUnitDelimiter NAL Unit.
	HEVC_NAL_AUD // AUD
	// NALU_EOS - End of Sequence NAL Unit.
	HEVC_NAL_EOS_NUT // EOS_NUT
	// NALU_EOB - End of Bitstream NAL Unit.
	HEVC_NAL_EOB_NUT // EOB_NUT
	// NALU_FD - Filler data NAL Unit.
	HEVC_NAL_FD_NUT // FD_NUT
	// NALU_SEI_PREFIX - Prefix SEI NAL Unit.
	HEVC_NAL_SEI_PREFIX // SEI_PREFIX
	// NALU_SEI_SUFFIX - Suffix SEI NAL Unit.
	HEVC_NAL_SEI_SUFFIX // SEI_SUFFIX
	// Unused.
	HEVC_NAL_RSV_NVCL41 // RSV_NVCL41
	HEVC_NAL_RSV_NVCL42 // RSV_NVCL42
	HEVC_NAL_RSV_NVCL43 // RSV_NVCL43
	HEVC_NAL_RSV_NVCL44 // RSV_NVCL44
	HEVC_NAL_RSV_NVCL45 // RSV_NVCL45
	HEVC_NAL_RSV_NVCL46 // RSV_NVCL46
	HEVC_NAL_RSV_NVCL47 // RSV_NVCL47
	HEVC_NAL_UNSPEC48   // UNSPEC48
	HEVC_NAL_UNSPEC49   // UNSPEC49
	HEVC_NAL_UNSPEC50   // UNSPEC50
	HEVC_NAL_UNSPEC51   // UNSPEC51
	HEVC_NAL_UNSPEC52   // UNSPEC52
	HEVC_NAL_UNSPEC53   // UNSPEC53
	HEVC_NAL_UNSPEC54   // UNSPEC54
	HEVC_NAL_UNSPEC55   // UNSPEC55
	HEVC_NAL_UNSPEC56   // UNSPEC56
	HEVC_NAL_UNSPEC57   // UNSPEC57
	HEVC_NAL_UNSPEC58   // UNSPEC58
	HEVC_NAL_UNSPEC59   // UNSPEC59
	HEVC_NAL_UNSPEC60   // UNSPEC60
	HEVC_NAL_UNSPEC61   // UNSPEC61
	HEVC_NAL_UNSPEC62   // UNSPEC62
	HEVC_NAL_UNSPEC63   // UNSPEC63
)

type SPSInfo struct {
	ProfileIdc             uint
	LevelIdc               uint
	MbWidth                uint
	MbHeight               uint
	CropLeft               uint
	CropRight              uint
	CropTop                uint
	CropBottom             uint
	Width                  uint
	Height                 uint
	numTemporalLayers      uint
	temporalIDNested       uint
	chromaFormat           uint
	PicWidthInLumaSamples  uint
	PicHeightInLumaSamples uint
	// bitDepthLumaMinus8               uint
	bitDepthChromaMinus8             uint
	generalProfileSpace              uint
	generalTierFlag                  uint
	generalProfileIDC                uint
	generalProfileCompatibilityFlags uint32
	generalConstraintIndicatorFlags  uint64
	generalLevelIDC                  uint
	fps                              uint
}

//nolint:gocyclo,cyclop,funlen
func (data NaluType) String() string {
	naluType := NaluType(data & parser.Last10BbitsNALUMask)
	switch naluType {
	case HEVC_NAL_TRAIL_N:
		return "TRAIL_N"
	case HEVC_NAL_TRAIL_R:
		return "TRAIL_R"
	case HEVC_NAL_TSA_N:
		return "TSA_N"
	case HEVC_NAL_TSA_R:
		return "TSA_R"
	case HEVC_NAL_STSA_N:
		return "STSA_N"
	case HEVC_NAL_STSA_R:
		return "STSA_R"
	case HEVC_NAL_RADL_N:
		return "RADL_N"
	case HEVC_NAL_RADL_R:
		return "RADL_R"
	case HEVC_NAL_RASL_N:
		return "RASL_N"
	case HEVC_NAL_RASL_R:
		return "RASL_R"
	case HEVC_NAL_VCL_N10:
		return "VCL_N10"
	case HEVC_NAL_VCL_R11:
		return "VCL_R11"
	case HEVC_NAL_VCL_N12:
		return "VCL_N12"
	case HEVC_NAL_VCL_R13:
		return "VCL_R13"
	case HEVC_NAL_VCL_N14:
		return "VCL_N14"
	case HEVC_NAL_VCL_R15:
		return "VCL_R15"
	case HEVC_NAL_BLA_W_LP:
		return "BLA_W_LP"
	case HEVC_NAL_BLA_W_RADL:
		return "BLA_W_RADL"
	case HEVC_NAL_BLA_N_LP:
		return "BLA_N_LP"
	case HEVC_NAL_IDR_W_RADL:
		return "IDR_W_RADL"
	case HEVC_NAL_IDR_N_LP:
		return "IDR_N_LP"
	case HEVC_NAL_CRA_NUT:
		return "CRA_NUT"
	case HEVC_NAL_RSV_IRAP_VCL22:
		return "RSV_IRAP_VCL22"
	case HEVC_NAL_RSV_IRAP_VCL23:
		return "RSV_IRAP_VCL23"
	case HEVC_NAL_RSV_VCL24:
		return "RSV_VCL24"
	case HEVC_NAL_RSV_VCL25:
		return "RSV_VCL25"
	case HEVC_NAL_RSV_VCL26:
		return "RSV_VCL26"
	case HEVC_NAL_RSV_VCL27:
		return "RSV_VCL27"
	case HEVC_NAL_RSV_VCL28:
		return "RSV_VCL28"
	case HEVC_NAL_RSV_VCL29:
		return "RSV_VCL29"
	case HEVC_NAL_RSV_VCL30:
		return "RSV_VCL30"
	case HEVC_NAL_RSV_VCL31:
		return "RSV_VCL31"
	case HEVC_NAL_VPS:
		return "VPS"
	case HEVC_NAL_SPS:
		return "SPS"
	case HEVC_NAL_PPS:
		return "PPS"
	case HEVC_NAL_AUD:
		return "AUD"
	case HEVC_NAL_EOS_NUT:
		return "EOS_NUT"
	case HEVC_NAL_EOB_NUT:
		return "EOB_NUT"
	case HEVC_NAL_FD_NUT:
		return "FD_NUT"
	case HEVC_NAL_SEI_PREFIX:
		return "SEI_PREFIX"
	case HEVC_NAL_SEI_SUFFIX:
		return "SEI_SUFFIX"
	case HEVC_NAL_RSV_NVCL41:
		return "RSV_NVCL41"
	case HEVC_NAL_RSV_NVCL42:
		return "RSV_NVCL42"
	case HEVC_NAL_RSV_NVCL43:
		return "RSV_NVCL43"
	case HEVC_NAL_RSV_NVCL44:
		return "RSV_NVCL44"
	case HEVC_NAL_RSV_NVCL45:
		return "RSV_NVCL45"
	case HEVC_NAL_RSV_NVCL46:
		return "RSV_NVCL46"
	case HEVC_NAL_RSV_NVCL47:
		return "RSV_NVCL47"
	case HEVC_NAL_UNSPEC48:
		return "UNSPEC48"
	case HEVC_NAL_UNSPEC49:
		return "UNSPEC49"
	case HEVC_NAL_UNSPEC50:
		return "UNSPEC50"
	case HEVC_NAL_UNSPEC51:
		return "UNSPEC51"
	case HEVC_NAL_UNSPEC52:
		return "UNSPEC52"
	case HEVC_NAL_UNSPEC53:
		return "UNSPEC53"
	case HEVC_NAL_UNSPEC54:
		return "UNSPEC54"
	case HEVC_NAL_UNSPEC55:
		return "UNSPEC55"
	case HEVC_NAL_UNSPEC56:
		return "UNSPEC56"
	case HEVC_NAL_UNSPEC57:
		return "UNSPEC57"
	case HEVC_NAL_UNSPEC58:
		return "UNSPEC58"
	case HEVC_NAL_UNSPEC59:
		return "UNSPEC59"
	case HEVC_NAL_UNSPEC60:
		return "UNSPEC60"
	case HEVC_NAL_UNSPEC61:
		return "UNSPEC61"
	case HEVC_NAL_UNSPEC62:
		return "UNSPEC62"
	case HEVC_NAL_UNSPEC63:
		return "UNSPEC63"
	default:
		return fmt.Sprintf("UNKNOWN(%d)", data)
	}
}

const (
	MaxVpsCount  = 16
	MaxSubLayers = 7
	MaxSpsCount  = 32
)

func IsKeyFrame(nalHeader []byte) bool {
	typ := (nalHeader[0] >> 1) & parser.Last10BbitsNALUMask

	return typ == byte(HEVC_NAL_IDR_N_LP) || typ == byte(HEVC_NAL_IDR_W_RADL)
}

func IsDataNALU(nalHeader []byte) bool {
	typ := (nalHeader[0] >> 1) & parser.Last10BbitsNALUMask

	return typ >= byte(HEVC_NAL_TRAIL_R) && typ <= byte(HEVC_NAL_IDR_N_LP)
}

func IsSPSNALU(nalHeader []byte) bool {
	typ := (nalHeader[0] >> 1) & parser.Last10BbitsNALUMask

	return typ == byte(HEVC_NAL_SPS)
}

func IsPPSNALU(nalHeader []byte) bool {
	typ := (nalHeader[0] >> 1) & parser.Last10BbitsNALUMask

	return typ == byte(HEVC_NAL_PPS)
}

func IsVPSNALU(nalHeader []byte) bool {
	typ := (nalHeader[0] >> 1) & parser.Last10BbitsNALUMask

	return typ == byte(HEVC_NAL_VPS)
}

var AUDBytes = []byte{0, 0, 0, 1, 0x9, 0xf0, 0, 0, 0, 1} //nolint:gochecknoglobals // AUD

func CheckNALUsType(b []byte) parser.NALUAvccOrAnnexb {
	_, typ := parser.SplitNALUs(b)

	return typ
}

//nolint:gocyclo,cyclop,funlen
func ParseSPS(sps []byte) (SPSInfo, error) {
	var spsInfo SPSInfo

	var err error
	if len(sps) < 2 {
		err = ErrH265IncorectUnitSize

		return spsInfo, err
	}

	rbsp := nal2rbsp(sps[2:])

	br := &bits.GolombBitReader{R: bytes.NewReader(rbsp)}
	if _, err = br.ReadBits(4); err != nil {
		return spsInfo, err
	}

	spsMaxSubLayersMinus1, err := br.ReadBits(3)
	if err != nil {
		return spsInfo, err
	}

	if spsMaxSubLayersMinus1+1 > spsInfo.numTemporalLayers {
		spsInfo.numTemporalLayers = spsMaxSubLayersMinus1 + 1
	}

	if spsInfo.temporalIDNested, err = br.ReadBit(); err != nil {
		return spsInfo, err
	}

	if err = parsePTL(br, &spsInfo, spsMaxSubLayersMinus1); err != nil {
		return spsInfo, err
	}

	if _, err = br.ReadExponentialGolombCode(); err != nil {
		return spsInfo, err
	}

	var cf uint

	if cf, err = br.ReadExponentialGolombCode(); err != nil {
		return spsInfo, err
	}

	spsInfo.chromaFormat = cf
	if spsInfo.chromaFormat == 3 {
		if _, err = br.ReadBit(); err != nil {
			return spsInfo, err
		}
	}

	if spsInfo.PicWidthInLumaSamples, err = br.ReadExponentialGolombCode(); err != nil {
		return spsInfo, err
	}

	spsInfo.Width = spsInfo.PicWidthInLumaSamples
	if spsInfo.PicHeightInLumaSamples, err = br.ReadExponentialGolombCode(); err != nil {
		return spsInfo, err
	}

	spsInfo.Height = spsInfo.PicHeightInLumaSamples
	conformanceWindowFlag, err := br.ReadBit()
	if err != nil {
		return spsInfo, err
	}

	if conformanceWindowFlag != 0 {
		if _, err = br.ReadExponentialGolombCode(); err != nil {
			return spsInfo, err
		}

		if _, err = br.ReadExponentialGolombCode(); err != nil {
			return spsInfo, err
		}

		if _, err = br.ReadExponentialGolombCode(); err != nil {
			return spsInfo, err
		}

		if _, err = br.ReadExponentialGolombCode(); err != nil {
			return spsInfo, err
		}
	}

	var bdlm8 uint

	if bdlm8, err = br.ReadExponentialGolombCode(); err != nil {
		return spsInfo, err
	}

	spsInfo.bitDepthChromaMinus8 = bdlm8

	var bdcm8 uint

	if bdcm8, err = br.ReadExponentialGolombCode(); err != nil {
		return spsInfo, err
	}

	spsInfo.bitDepthChromaMinus8 = bdcm8

	_, err = br.ReadExponentialGolombCode()
	if err != nil {
		return spsInfo, err
	}

	spsSubLayerOrderingInfoPresentFlag, err := br.ReadBit()
	if err != nil {
		return spsInfo, err
	}

	var i uint
	if spsSubLayerOrderingInfoPresentFlag != 0 {
		i = 0
	} else {
		i = spsMaxSubLayersMinus1
	}

	for ; i <= spsMaxSubLayersMinus1; i++ {
		if _, err = br.ReadExponentialGolombCode(); err != nil {
			return spsInfo, err
		}

		if _, err = br.ReadExponentialGolombCode(); err != nil {
			return spsInfo, err
		}

		if _, err = br.ReadExponentialGolombCode(); err != nil {
			return spsInfo, err
		}
	}

	if _, err = br.ReadExponentialGolombCode(); err != nil {
		return spsInfo, err
	}

	if _, err = br.ReadExponentialGolombCode(); err != nil {
		return spsInfo, err
	}

	if _, err = br.ReadExponentialGolombCode(); err != nil {
		return spsInfo, err
	}

	if _, err = br.ReadExponentialGolombCode(); err != nil {
		return spsInfo, err
	}

	if _, err = br.ReadExponentialGolombCode(); err != nil {
		return spsInfo, err
	}

	if _, err = br.ReadExponentialGolombCode(); err != nil {
		return spsInfo, err
	}

	return spsInfo, err
}

func parsePTL(br *bits.GolombBitReader, ctx *SPSInfo, maxSubLayersMinus1 uint) error {
	var err error

	var ptl SPSInfo

	if ptl.generalProfileSpace, err = br.ReadBits(2); err != nil {
		return err
	}

	if ptl.generalTierFlag, err = br.ReadBit(); err != nil {
		return err
	}

	if ptl.generalProfileIDC, err = br.ReadBits(5); err != nil {
		return err
	}

	if ptl.generalProfileCompatibilityFlags, err = br.ReadBits32(32); err != nil {
		return err
	}

	if ptl.generalConstraintIndicatorFlags, err = br.ReadBits64(48); err != nil {
		return err
	}

	if ptl.generalLevelIDC, err = br.ReadBits(8); err != nil {
		return err
	}

	updatePTL(ctx, &ptl)

	if maxSubLayersMinus1 == 0 {
		return nil
	}

	subLayerProfilePresentFlag := make([]uint, maxSubLayersMinus1)
	subLayerLevelPresentFlag := make([]uint, maxSubLayersMinus1)

	for i := range maxSubLayersMinus1 {
		if subLayerProfilePresentFlag[i], err = br.ReadBit(); err != nil {
			return err
		}

		if subLayerLevelPresentFlag[i], err = br.ReadBit(); err != nil {
			return err
		}
	}

	if maxSubLayersMinus1 > 0 {
		for i := maxSubLayersMinus1; i < 8; i++ {
			if _, err = br.ReadBits(2); err != nil {
				return err
			}
		}
	}

	for i := range maxSubLayersMinus1 {
		if subLayerProfilePresentFlag[i] != 0 {
			if _, err = br.ReadBits32(32); err != nil {
				return err
			}

			if _, err = br.ReadBits32(32); err != nil {
				return err
			}

			if _, err = br.ReadBits32(24); err != nil {
				return err
			}
		}

		if subLayerLevelPresentFlag[i] != 0 {
			if _, err = br.ReadBits(8); err != nil {
				return err
			}
		}
	}

	return nil
}

func updatePTL(ctx, ptl *SPSInfo) {
	ctx.generalProfileSpace = ptl.generalProfileSpace

	if ptl.generalTierFlag > ctx.generalTierFlag {
		ctx.generalLevelIDC = ptl.generalLevelIDC

		ctx.generalTierFlag = ptl.generalTierFlag
	} else if ptl.generalLevelIDC > ctx.generalLevelIDC {
		ctx.generalLevelIDC = ptl.generalLevelIDC
	}

	if ptl.generalProfileIDC > ctx.generalProfileIDC {
		ctx.generalProfileIDC = ptl.generalProfileIDC
	}

	ctx.generalProfileCompatibilityFlags &= ptl.generalProfileCompatibilityFlags

	ctx.generalConstraintIndicatorFlags &= ptl.generalConstraintIndicatorFlags
}

func nal2rbsp(nal []byte) []byte {
	return bytes.ReplaceAll(nal, []byte{0x0, 0x0, 0x3}, []byte{0x0, 0x0})
}

type CodecData struct {
	Record     []byte
	RecordInfo AVCDecoderConfRecord
	SPSInfo    SPSInfo
	ControlURL string
}

func (s CodecData) Type() av.CodecType {
	return av.H265
}

func (s CodecData) AVCDecoderConfRecordBytes() []byte {
	return s.Record
}

func (s CodecData) SPS() []byte {
	return s.RecordInfo.SPS[0]
}

func (s CodecData) PPS() []byte {
	return s.RecordInfo.PPS[0]
}

func (s CodecData) VPS() []byte {
	return s.RecordInfo.VPS[0]
}

func (s CodecData) Width() int {
	return int(s.SPSInfo.Width)
}

func (s CodecData) Height() int {
	return int(s.SPSInfo.Height)
}

func (s CodecData) FPS() int {
	return int(s.SPSInfo.fps)
}

func (s CodecData) TrackID() string {
	return s.ControlURL
}

func (s CodecData) Resolution() string {
	return fmt.Sprintf("%vx%v", s.Width(), s.Height())
}

func (s CodecData) Tag() string {
	return fmt.Sprintf("hvc1.%02X%02X%02X", s.RecordInfo.AVCProfileIndication, s.RecordInfo.ProfileCompatibility, s.RecordInfo.AVCLevelIndication)
	// return "hev1.1.6.L120.90"
}

func (s CodecData) Bandwidth() string {
	return strconv.Itoa((int(float64(s.Width()) * (float64(1.71) * (30 / float64(s.FPS()))))) * 1000)
}

func (s CodecData) PacketDuration(_ []byte) time.Duration {
	return time.Duration(1000./float64(s.FPS())) * time.Millisecond
}

func NewCodecDataFromAVCDecoderConfRecord(record []byte) (CodecData, error) {
	var s CodecData

	var err error

	s.Record = record
	if _, err = (&s.RecordInfo).Unmarshal(record); err != nil {
		return s, err
	}

	if len(s.RecordInfo.SPS) == 0 {
		return s, ErrSPSNotFound
	}

	if len(s.RecordInfo.PPS) == 0 {
		return s, ErrPPSNotFound
	}

	if len(s.RecordInfo.VPS) == 0 {
		return s, ErrVPSNotFound
	}

	if s.SPSInfo, err = ParseSPS(s.RecordInfo.SPS[0]); err != nil {
		return s, errors.Join(ErrSPSParseFailed, err)
	}

	return s, nil
}

func NewCodecDataFromVPSAndSPSAndPPS(vps, sps, pps []byte) (CodecData, error) {
	var s CodecData

	var err error

	recordinfo := AVCDecoderConfRecord{}
	recordinfo.AVCProfileIndication = sps[3]
	recordinfo.ProfileCompatibility = sps[4]
	recordinfo.AVCLevelIndication = sps[5]
	recordinfo.SPS = [][]byte{sps}
	recordinfo.PPS = [][]byte{pps}
	recordinfo.VPS = [][]byte{vps}
	recordinfo.LengthSizeMinusOne = 3

	if s.SPSInfo, err = ParseSPS(sps); err != nil {
		return s, err
	}

	buf := make([]byte, recordinfo.Len())
	recordinfo.Marshal(buf, s.SPSInfo)
	s.RecordInfo = recordinfo
	s.Record = buf

	return s, err
}

type AVCDecoderConfRecord struct {
	AVCProfileIndication uint8
	ProfileCompatibility uint8
	AVCLevelIndication   uint8
	LengthSizeMinusOne   uint8
	VPS                  [][]byte
	SPS                  [][]byte
	PPS                  [][]byte
}

func (s *AVCDecoderConfRecord) Unmarshal(b []byte) (int, error) {
	var n int

	var err error
	if len(b) < 30 {
		err = ErrDecconfInvalid

		return n, err
	}

	s.AVCProfileIndication = b[1]
	s.ProfileCompatibility = b[2]
	s.AVCLevelIndication = b[3]
	s.LengthSizeMinusOne = b[4] & 0x03

	vpscount := int(b[25] & 0x1f)
	n += 26

	for range vpscount {
		if len(b) < n+2 {
			err = ErrDecconfInvalid

			return n, err
		}

		vpslen := int(pio.U16BE(b[n:]))
		n += 2

		if len(b) < n+vpslen {
			err = ErrDecconfInvalid

			return n, err
		}

		s.VPS = append(s.VPS, b[n:n+vpslen])
		n += vpslen
	}

	if len(b) < n+1 {
		err = ErrDecconfInvalid

		return n, err
	}

	n++
	n++

	spscount := int(b[n])
	n++

	for range spscount {
		if len(b) < n+2 {
			err = ErrDecconfInvalid

			return n, err
		}

		spslen := int(pio.U16BE(b[n:]))
		n += 2

		if len(b) < n+spslen {
			err = ErrDecconfInvalid

			return n, err
		}

		s.SPS = append(s.SPS, b[n:n+spslen])
		n += spslen
	}

	n++
	n++

	ppscount := int(b[n])
	n++

	for range ppscount {
		if len(b) < n+2 {
			err = ErrDecconfInvalid

			return n, err
		}

		ppslen := int(pio.U16BE(b[n:]))
		n += 2

		if len(b) < n+ppslen {
			err = ErrDecconfInvalid

			return n, err
		}

		s.PPS = append(s.PPS, b[n:n+ppslen])
		n += ppslen
	}

	return n, err
}

func (s *AVCDecoderConfRecord) Len() int {
	n := 23
	for _, sps := range s.SPS {
		n += 5 + len(sps)
	}

	for _, pps := range s.PPS {
		n += 5 + len(pps)
	}

	for _, vps := range s.VPS {
		n += 5 + len(vps)
	}

	return n
}

func (s *AVCDecoderConfRecord) Marshal(b []byte, _ SPSInfo) int {
	var n int

	b[0] = 1
	b[1] = s.AVCProfileIndication
	b[2] = s.ProfileCompatibility
	b[3] = s.AVCLevelIndication
	b[21] = 3
	b[22] = 3
	n += 23
	b[n] = (s.VPS[0][0] >> 1) & 0x3f
	n++
	b[n] = byte(len(s.VPS) >> 8)
	n++
	b[n] = byte(len(s.VPS))
	n++

	for _, vps := range s.VPS {
		pio.PutU16BE(b[n:], uint16(len(vps)))
		n += 2
		copy(b[n:], vps)
		n += len(vps)
	}

	b[n] = (s.SPS[0][0] >> 1) & 0x3f
	n++
	b[n] = byte(len(s.SPS) >> 8)
	n++
	b[n] = byte(len(s.SPS))
	n++

	for _, sps := range s.SPS {
		pio.PutU16BE(b[n:], uint16(len(sps)))
		n += 2
		copy(b[n:], sps)
		n += len(sps)
	}

	b[n] = (s.PPS[0][0] >> 1) & 0x3f
	n++
	b[n] = byte(len(s.PPS) >> 8)
	n++
	b[n] = byte(len(s.PPS))
	n++

	for _, pps := range s.PPS {
		pio.PutU16BE(b[n:], uint16(len(pps)))
		n += 2
		copy(b[n:], pps)
		n += len(pps)
	}

	return n
}

type SliceType uint

func (s SliceType) String() string {
	switch s {
	case SliceP:
		return "P"
	case SliceB:
		return "B"
	case SliceI:
		return "I"
	}

	return ""
}

const (
	SliceP SliceType = iota + 1
	SliceB
	SliceI
)

func ParseSliceHeaderFromNALU(pkt []byte) (SliceType, error) {
	var sliceType SliceType

	var err error
	if len(pkt) <= 1 {
		err = ErrPacketTooShort

		return sliceType, err
	}

	nalUnitType := pkt[0] & 0x1f
	switch nalUnitType {
	case 1, 2, 5, 19:

	default:
		err = ErrNalHasNoSliceHeader

		return sliceType, err
	}

	r := &bits.GolombBitReader{R: bytes.NewReader(pkt[1:])}
	if _, err = r.ReadExponentialGolombCode(); err != nil {
		return sliceType, err
	}

	var u uint

	if u, err = r.ReadExponentialGolombCode(); err != nil {
		return sliceType, err
	}

	switch u {
	case 0, 3, 5, 8:
		sliceType = SliceP
	case 1, 6:
		sliceType = SliceB
	case 2, 4, 7, 9:
		sliceType = SliceI
	default:
		err = ErrInvalidSliceType

		return sliceType, err
	}

	return sliceType, err
}
