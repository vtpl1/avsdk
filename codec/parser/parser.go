package parser

import (
	"encoding/binary"

	"github.com/vtpl1/avsdk/utils/bits/pio"
)

var (
	StartCode3 = []byte{0x00, 0x00, 0x01}       //nolint:gochecknoglobals
	StartCode4 = []byte{0x00, 0x00, 0x00, 0x01} //nolint:gochecknoglobals
	// StartCodes is retained for clarity or potential external use, though not directly used in the optimized lenStartCode.
	StartCodes = [][]byte{StartCode3, StartCode4} //nolint:gochecknoglobals
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

// Optimized lenStartCode to use direct byte checks, avoiding bytes.HasPrefix and loop overhead.
func lenStartCode(data []byte) int {
	if len(data) >= 4 && data[0] == 0x00 && data[1] == 0x00 && data[2] == 0x00 && data[3] == 0x01 {
		return 4
	}
	if len(data) >= 3 && data[0] == 0x00 && data[1] == 0x00 && data[2] == 0x01 {
		return 3
	}
	return 0
}

func hasAnnexBStartCode(data []byte) bool {
	return lenStartCode(data) > 0
}

func IsAnnexBOrAVCC(data []byte) NALUAvccOrAnnexb {
	if len(data) < 4 {
		return NALURaw
	}
	if hasAnnexBStartCode(data) {
		return NALUAnnexb
	}
	// Check if the first 4 bytes represent a valid NALU length for AVCC.
	// The length should be greater than 0 and not exceed the remaining data length.
	naluLen := readNALULength(data[:4])
	if naluLen > 0 && naluLen <= len(data)-4 {
		return NALUAvcc
	}

	return NALURaw
}

func readNALULength(b []byte) int {
	if len(b) < 4 {
		return 0
	}
	// Using binary.BigEndian.Uint32 is generally idiomatic and might be micro-optimized by the Go runtime.
	return int(binary.BigEndian.Uint32(b[:4]))
}

// SplitNALUs optimizes AnnexB parsing by performing direct byte checks for start codes
// within the main loop to avoid repeated slicing and function call overhead.
func SplitNALUs(b []byte) ([][]byte, NALUAvccOrAnnexb) {
	annexBOrAvccOrRaw := IsAnnexBOrAVCC(b)
	if annexBOrAvccOrRaw == NALUAnnexb {
		var nalus [][]byte

		// Optimized loop to find all start code positions by direct byte checking
		naluIndices := []int{}
		i := 0
		for i < len(b) {
			scLen := 0
			// Directly check for 4-byte start code first (most common and longest)
			if i+4 <= len(b) && b[i] == 0x00 && b[i+1] == 0x00 && b[i+2] == 0x00 && b[i+3] == 0x01 {
				scLen = 4
			} else if i+3 <= len(b) && b[i] == 0x00 && b[i+1] == 0x00 && b[i+2] == 0x01 {
				// Directly check for 3-byte start code
				scLen = 3
			}

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

		// Extract NALUs
		for i := range naluIndices {
			start := naluIndices[i]
			end := len(b)
			if next := i + 1; next < len(naluIndices) {
				end = naluIndices[next]
			}
			nalu := b[start:end]

			// Determine offset using the now-optimized lenStartCode
			offset := lenStartCode(nalu)

			if offset >= len(nalu) {
				continue // corrupted NALU or just a start code (e.g., 00 00 01 at end of stream)
			}
			naluNoPrefix := nalu[offset:]
			if len(naluNoPrefix) > 0 {
				nalus = append(nalus, naluNoPrefix)
			}
		}

		return nalus, NALUAnnexb
	} else if annexBOrAvccOrRaw == NALUAvcc {
		_val4 := pio.U32BE(b)
		_b := b[4:]
		nalus := [][]byte{}

		// The AVCC parsing loop is already quite efficient with direct slicing and integer operations.
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

		if len(_b) == 0 { // Check if all data was consumed
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
		// Benefits from the optimized hasAnnexBStartCode (which uses optimized lenStartCode)
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
		// Benefits from the optimized hasAnnexBStartCode (which uses optimized lenStartCode)
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
		output = append(output, StartCode4...) // 4-byte start code
		output = append(output, data[offset:offset+naluLen]...)
		offset += naluLen
	}

	return output, nil
}
