// Package h265parser holds Muxer and Demuxer for h265
package h265parser

import (
	"bytes"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/vtpl1/avsdk/av"
	"github.com/vtpl1/avsdk/utils/bits"
	"github.com/vtpl1/avsdk/utils/bits/pio"
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

type NaluType int

const (
	HEVC_NAL_TRAIL_N        NaluType = iota // = 0
	HEVC_NAL_TRAIL_R                        // = 1,
	HEVC_NAL_TSA_N                          // = 2,
	HEVC_NAL_TSA_R                          // = 3,
	HEVC_NAL_STSA_N                         // = 4,
	HEVC_NAL_STSA_R                         // = 5,
	HEVC_NAL_RADL_N                         // = 6,
	HEVC_NAL_RADL_R                         // = 7,
	HEVC_NAL_RASL_N                         // = 8,
	HEVC_NAL_RASL_R                         // = 9,
	HEVC_NAL_VCL_N10                        // = 10,
	HEVC_NAL_VCL_R11                        // = 11,
	HEVC_NAL_VCL_N12                        // = 12,
	HEVC_NAL_VCL_R13                        // = 13,
	HEVC_NAL_VCL_N14                        // = 14,
	HEVC_NAL_VCL_R15                        // = 15,
	HEVC_NAL_BLA_W_LP                       // = 16,
	HEVC_NAL_BLA_W_RADL                     // = 17,
	HEVC_NAL_BLA_N_LP                       // = 18,
	HEVC_NAL_IDR_W_RADL                     // = 19,
	HEVC_NAL_IDR_N_LP                       // = 20,
	HEVC_NAL_CRA_NUT                        // = 21,
	HEVC_NAL_RSV_IRAP_VCL22                 // = 22,
	HEVC_NAL_RSV_IRAP_VCL23                 // = 23,
	HEVC_NAL_RSV_VCL24                      // = 24,
	HEVC_NAL_RSV_VCL25                      // = 25,
	HEVC_NAL_RSV_VCL26                      // = 26,
	HEVC_NAL_RSV_VCL27                      // = 27,
	HEVC_NAL_RSV_VCL28                      // = 28,
	HEVC_NAL_RSV_VCL29                      // = 29,
	HEVC_NAL_RSV_VCL30                      // = 30,
	HEVC_NAL_RSV_VCL31                      // = 31,
	HEVC_NAL_VPS                            // = 32,
	HEVC_NAL_SPS                            // = 33,
	HEVC_NAL_PPS                            // = 34,
	HEVC_NAL_AUD                            // = 35,
	HEVC_NAL_EOS_NUT                        // = 36,
	HEVC_NAL_EOB_NUT                        // = 37,
	HEVC_NAL_FD_NUT                         // = 38,
	HEVC_NAL_SEI_PREFIX                     // = 39,
	HEVC_NAL_SEI_SUFFIX                     // = 40,
	HEVC_NAL_RSV_NVCL41                     // = 41,
	HEVC_NAL_RSV_NVCL42                     // = 42,
	HEVC_NAL_RSV_NVCL43                     // = 43,
	HEVC_NAL_RSV_NVCL44                     // = 44,
	HEVC_NAL_RSV_NVCL45                     // = 45,
	HEVC_NAL_RSV_NVCL46                     // = 46,
	HEVC_NAL_RSV_NVCL47                     // = 47,
	HEVC_NAL_UNSPEC48                       // = 48,
	HEVC_NAL_UNSPEC49                       // = 49,
	HEVC_NAL_UNSPEC50                       // = 50,
	HEVC_NAL_UNSPEC51                       // = 51,
	HEVC_NAL_UNSPEC52                       // = 52,
	HEVC_NAL_UNSPEC53                       // = 53,
	HEVC_NAL_UNSPEC54                       // = 54,
	HEVC_NAL_UNSPEC55                       // = 55,
	HEVC_NAL_UNSPEC56                       // = 56,
	HEVC_NAL_UNSPEC57                       // = 57,
	HEVC_NAL_UNSPEC58                       // = 58,
	HEVC_NAL_UNSPEC59                       // = 59,
	HEVC_NAL_UNSPEC60                       // = 60,
	HEVC_NAL_UNSPEC61                       // = 61,
	HEVC_NAL_UNSPEC62                       // = 62,
	HEVC_NAL_UNSPEC63                       // = 63,
)

const (
	MaxVpsCount  = 16
	MaxSubLayers = 7
	MaxSpsCount  = 32
)

func IsDataNALU(b []byte) bool {
	typ := b[0] & 0x1f

	return typ >= 1 && typ <= 5
}

var (
	StartCodeBytes = []byte{0, 0, 1}                           //nolint:gochecknoglobals
	AUDBytes       = []byte{0, 0, 0, 1, 0x9, 0xf0, 0, 0, 0, 1} //nolint:gochecknoglobals // AUD
)

func CheckNALUsType(b []byte) NALUAvccOrAnnexb {
	_, typ := SplitNALUs(b)

	return typ
}

type NALUAvccOrAnnexb int

const (
	NALURaw NALUAvccOrAnnexb = iota
	NALUAvcc
	NALUAnnexb
)

//nolint:gocognit
func SplitNALUs(b []byte) ([][]byte, NALUAvccOrAnnexb) {
	var nalus [][]byte

	if len(b) < 4 {
		return [][]byte{b}, NALURaw
	}

	val3 := pio.U24BE(b)

	val4 := pio.U32BE(b)
	if val4 <= uint32(len(b)) {
		_val4 := val4
		_b := b[4:]
		nalus = [][]byte{}

		for {
			nalus = append(nalus, _b[:_val4])

			_b = _b[_val4:]
			if len(_b) < 4 {
				break
			}

			_val4 = pio.U32BE(_b)
			_b = _b[4:]

			if _val4 > uint32(len(_b)) {
				break
			}
		}

		if len(_b) == 0 {
			return nalus, NALUAvcc
		}
	}

	if val3 == 1 || val4 == 1 {
		_val3 := val3
		_val4 := val4
		start := 0
		pos := 0

		for {
			if start != pos {
				nalus = append(nalus, b[start:pos])
			}

			if _val3 == 1 {
				pos += 3
			} else if _val4 == 1 {
				pos += 4
			}

			start = pos
			if start == len(b) {
				break
			}

			_val3 = 0
			_val4 = 0

			for pos < len(b) {
				if pos+2 < len(b) && b[pos] == 0 {
					_val3 = pio.U24BE(b[pos:])
					if _val3 == 0 {
						if pos+3 < len(b) {
							_val4 = uint32(b[pos+3])
							if _val4 == 1 {
								break
							}
						}
					} else if _val3 == 1 {
						break
					}

					pos++
				} else {
					pos++
				}
			}
		}

		return nalus, NALUAnnexb
	}

	return [][]byte{b}, NALURaw
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

func (s CodecData) Resolution() string {
	return fmt.Sprintf("%vx%v", s.Width(), s.Height())
}

func (s CodecData) Tag() string {
	// return fmt.Sprintf("hvc1.%02X%02X%02X", s.RecordInfo.AVCProfileIndication, s.RecordInfo.ProfileCompatibility, s.RecordInfo.AVCLevelIndication)
	return "hev1.1.6.L120.90"
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
