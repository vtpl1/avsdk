package parser

import (
	"bytes"
	"encoding/binary"

	"github.com/vtpl1/avsdk/utils/bits/pio"
)

var (
	startCode3 = []byte{0x00, 0x00, 0x01}         //nolint:gochecknoglobals
	startCode4 = []byte{0x00, 0x00, 0x00, 0x01}   //nolint:gochecknoglobals
	startCodes = [][]byte{startCode3, startCode4} //nolint:gochecknoglobals
)

type NALUAvccOrAnnexb int

const (
	NALURaw NALUAvccOrAnnexb = iota
	NALUAvcc
	NALUAnnexb
)

const (
	Last9BbitsNALUMask  = 0x1F
	Last10BbitsNALUMask = 0x3F
	MinimumNALULength   = 4
)

type NaluType byte

const (
	H264_NAL_UNSPECIFIED       NaluType = iota //nolint:stylecheck
	H264_NAL_SLICE                             //nolint:stylecheck
	H264_NAL_DPA                               //nolint:stylecheck
	H264_NAL_DPB                               //nolint:stylecheck
	H264_NAL_DPC                               //nolint:stylecheck
	H264_NAL_IDR_SLICE                         //nolint:stylecheck
	H264_NAL_SEI                               //nolint:stylecheck
	H264_NAL_SPS                               //nolint:stylecheck
	H264_NAL_PPS                               //nolint:stylecheck
	H264_NAL_AUD                               //nolint:stylecheck
	H264_NAL_END_SEQUENCE                      //nolint:stylecheck
	H264_NAL_END_STREAM                        //nolint:stylecheck
	H264_NAL_FILLER_DATA                       //nolint:stylecheck
	H264_NAL_SPS_EXT                           //nolint:stylecheck
	H264_NAL_PREFIX                            //nolint:stylecheck
	H264_NAL_SUB_SPS                           //nolint:stylecheck
	H264_NAL_DPS                               //nolint:stylecheck
	H264_NAL_RESERVED17                        //nolint:stylecheck
	H264_NAL_RESERVED18                        //nolint:stylecheck
	H264_NAL_AUXILIARY_SLICE                   //nolint:stylecheck
	H264_NAL_EXTEN_SLICE                       //nolint:stylecheck
	H264_NAL_DEPTH_EXTEN_SLICE                 //nolint:stylecheck
	H264_NAL_RESERVED22                        //nolint:stylecheck
	H264_NAL_RESERVED23                        //nolint:stylecheck
	H264_NAL_UNSPECIFIED24                     //nolint:stylecheck
	H264_NAL_UNSPECIFIED25                     //nolint:stylecheck
	H264_NAL_UNSPECIFIED26                     //nolint:stylecheck
	H264_NAL_UNSPECIFIED27                     //nolint:stylecheck
	H264_NAL_UNSPECIFIED28                     //nolint:stylecheck
	H264_NAL_UNSPECIFIED29                     //nolint:stylecheck
	H264_NAL_UNSPECIFIED30                     //nolint:stylecheck
	H264_NAL_UNSPECIFIED31                     //nolint:stylecheck
)

const (
	HEVC_NAL_TRAIL_N        NaluType = iota //nolint:stylecheck
	HEVC_NAL_TRAIL_R                        //nolint:stylecheck
	HEVC_NAL_TSA_N                          //nolint:stylecheck
	HEVC_NAL_TSA_R                          //nolint:stylecheck
	HEVC_NAL_STSA_N                         //nolint:stylecheck
	HEVC_NAL_STSA_R                         //nolint:stylecheck
	HEVC_NAL_RADL_N                         //nolint:stylecheck
	HEVC_NAL_RADL_R                         //nolint:stylecheck
	HEVC_NAL_RASL_N                         //nolint:stylecheck
	HEVC_NAL_RASL_R                         //nolint:stylecheck
	HEVC_NAL_VCL_N10                        //nolint:stylecheck
	HEVC_NAL_VCL_R11                        //nolint:stylecheck
	HEVC_NAL_VCL_N12                        //nolint:stylecheck
	HEVC_NAL_VCL_R13                        //nolint:stylecheck
	HEVC_NAL_VCL_N14                        //nolint:stylecheck
	HEVC_NAL_VCL_R15                        //nolint:stylecheck
	HEVC_NAL_BLA_W_LP                       //nolint:stylecheck
	HEVC_NAL_BLA_W_RADL                     //nolint:stylecheck
	HEVC_NAL_BLA_N_LP                       //nolint:stylecheck
	HEVC_NAL_IDR_W_RADL                     //nolint:stylecheck
	HEVC_NAL_IDR_N_LP                       //nolint:stylecheck
	HEVC_NAL_CRA_NUT                        //nolint:stylecheck
	HEVC_NAL_RSV_IRAP_VCL22                 //nolint:stylecheck
	HEVC_NAL_RSV_IRAP_VCL23                 //nolint:stylecheck
	HEVC_NAL_RSV_VCL24                      //nolint:stylecheck
	HEVC_NAL_RSV_VCL25                      //nolint:stylecheck
	HEVC_NAL_RSV_VCL26                      //nolint:stylecheck
	HEVC_NAL_RSV_VCL27                      //nolint:stylecheck
	HEVC_NAL_RSV_VCL28                      //nolint:stylecheck
	HEVC_NAL_RSV_VCL29                      //nolint:stylecheck
	HEVC_NAL_RSV_VCL30                      //nolint:stylecheck
	HEVC_NAL_RSV_VCL31                      //nolint:stylecheck
	HEVC_NAL_VPS                            //nolint:stylecheck
	HEVC_NAL_SPS                            //nolint:stylecheck
	HEVC_NAL_PPS                            //nolint:stylecheck
	HEVC_NAL_AUD                            //nolint:stylecheck
	HEVC_NAL_EOS_NUT                        //nolint:stylecheck
	HEVC_NAL_EOB_NUT                        //nolint:stylecheck
	HEVC_NAL_FD_NUT                         //nolint:stylecheck
	HEVC_NAL_SEI_PREFIX                     //nolint:stylecheck
	HEVC_NAL_SEI_SUFFIX                     //nolint:stylecheck
	HEVC_NAL_RSV_NVCL41                     //nolint:stylecheck
	HEVC_NAL_RSV_NVCL42                     //nolint:stylecheck
	HEVC_NAL_RSV_NVCL43                     //nolint:stylecheck
	HEVC_NAL_RSV_NVCL44                     //nolint:stylecheck
	HEVC_NAL_RSV_NVCL45                     //nolint:stylecheck
	HEVC_NAL_RSV_NVCL46                     //nolint:stylecheck
	HEVC_NAL_RSV_NVCL47                     //nolint:stylecheck
	HEVC_NAL_UNSPEC48                       //nolint:stylecheck
	HEVC_NAL_UNSPEC49                       //nolint:stylecheck
	HEVC_NAL_UNSPEC50                       //nolint:stylecheck
	HEVC_NAL_UNSPEC51                       //nolint:stylecheck
	HEVC_NAL_UNSPEC52                       //nolint:stylecheck
	HEVC_NAL_UNSPEC53                       //nolint:stylecheck
	HEVC_NAL_UNSPEC54                       //nolint:stylecheck
	HEVC_NAL_UNSPEC55                       //nolint:stylecheck
	HEVC_NAL_UNSPEC56                       //nolint:stylecheck
	HEVC_NAL_UNSPEC57                       //nolint:stylecheck
	HEVC_NAL_UNSPEC58                       //nolint:stylecheck
	HEVC_NAL_UNSPEC59                       //nolint:stylecheck
	HEVC_NAL_UNSPEC60                       //nolint:stylecheck
	HEVC_NAL_UNSPEC61                       //nolint:stylecheck
	HEVC_NAL_UNSPEC62                       //nolint:stylecheck
	HEVC_NAL_UNSPEC63                       //nolint:stylecheck
)

// -----------------------------
// NALU Format Detection
// -----------------------------

func IsAnnexBOrAVCC(data []byte) NALUAvccOrAnnexb {
	if len(data) < 4 {
		return NALURaw
	}
	if hasAnnexBStartCode(data) {
		return NALUAnnexb
	}
	if naluLen := readNALULength(data[:4]); naluLen > 0 && naluLen <= len(data)-4 {
		return NALUAvcc
	}

	return NALURaw
}

func hasAnnexBStartCode(data []byte) bool {
	return lenStartCode(data) > 0
}

func readNALULength(b []byte) int {
	if len(b) < 4 {
		return 0
	}

	return int(b[0])<<24 | int(b[1])<<16 | int(b[2])<<8 | int(b[3])
}

func lenStartCode(data []byte) int {
	for _, sc := range startCodes {
		if bytes.HasPrefix(data, sc) {
			return len(sc)
		}
	}

	return 0
}

func SplitNALUs(b []byte) ([][]byte, NALUAvccOrAnnexb) {
	if len(b) < 4 {
		return [][]byte{b}, NALURaw
	}
	if IsAnnexBOrAVCC(b) == NALUAnnexb {
		var nalus [][]byte

		// Find all start code positions
		naluIndices := []int{}
		for i := 0; i < len(b)-2; {
			scLen := lenStartCode(b[i:])
			if scLen > 0 {
				naluIndices = append(naluIndices, i)
				i += scLen
			} else {
				i++
			}
		}

		// If no start codes found, fall back to single raw NALU
		if len(naluIndices) == 0 {
			return [][]byte{b}, NALURaw
		}

		// Extract NALUs and detect codec
		for i := range naluIndices {
			start := naluIndices[i]
			end := len(b)
			if next := i + 1; next < len(naluIndices) {
				end = naluIndices[next]
			}
			nalu := b[start:end]
			offset := lenStartCode(nalu)

			if offset >= len(nalu) {
				continue // corrupted NALU
			}
			naluNoPrefix := nalu[offset:]
			if len(naluNoPrefix) > 0 {
				nalus = append(nalus, naluNoPrefix)
			}
		}

		return nalus, NALUAnnexb
	}

	val4 := pio.U32BE(b)
	// maybe AVCC
	if val4 <= uint32(len(b)) {
		_val4 := val4
		_b := b[4:]
		nalus := [][]byte{}

		for {
			if _val4 > uint32(len(_b)) {
				break
			}

			nalus = append(nalus, _b[:_val4])

			_b = _b[_val4:]
			if len(_b) < MinimumNALULength {
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

	return [][]byte{b}, NALURaw
}

//nolint:nonamedreturns
func FindNextAnnexBNALUnit(data []byte, start int) (nalStart int, nalEnd int) {
	nalStart = -1

	// Find start code
	for i := start; i+3 < len(data); i++ {
		if hasAnnexBStartCode(data[i:]) {
			nalStart = i + lenStartCode(data[i:])

			break
		}
	}
	if nalStart == -1 {
		return -1, -1
	}

	// Find next start code
	for i := nalStart; i+3 < len(data); i++ {
		if hasAnnexBStartCode(data[i:]) {
			nalEnd = i

			return
		}
	}
	nalEnd = len(data)

	return
}

func AnnexBToAVCC(data []byte) ([]byte, error) {
	var output []byte
	offset := 0

	for offset < len(data) {
		start, end := FindNextAnnexBNALUnit(data, offset)
		if start < 0 || end < 0 {
			break
		}
		nalu := data[start:end]
		naluLen := uint32(len(nalu))

		var lengthBuf [4]byte
		binary.BigEndian.PutUint32(lengthBuf[:], naluLen)
		output = append(output, lengthBuf[:]...)
		output = append(output, nalu...)

		offset = end
	}

	return output, nil
}

func AVCCToAnnexB(data []byte) ([]byte, error) {
	var output []byte
	offset := 0

	for offset+4 <= len(data) {
		naluLen := readNALULength(data[offset : offset+4])
		offset += 4

		if offset+naluLen > len(data) {
			return nil, errInvalidNALULength
		}
		output = append(output, startCode4...) // 4-byte start code
		output = append(output, data[offset:offset+naluLen]...)
		offset += naluLen
	}

	return output, nil
}
