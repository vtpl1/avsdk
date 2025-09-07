package av

import (
	"fmt"
)

const (
	Last9BbitsNALUMask  = 0x1F
	Last10BbitsNALUMask = 0x3F
	MinimumNALULength   = 4
)

type NaluType byte

// -----------------------------------------------------------------------------.
//
//nolint:stylecheck // allow ALL_CAPS naming for H.264 NALU types
const (
	H264_NAL_UNSPECIFIED       NaluType = iota // UNSPECIFIED
	H264_NAL_SLICE                             // NON_IDR_SLICE
	H264_NAL_DPA                               // DPA
	H264_NAL_DPB                               // DPB
	H264_NAL_DPC                               // DPC
	H264_NAL_IDR_SLICE                         // IDR_SLICE
	H264_NAL_SEI                               // SEI
	H264_NAL_SPS                               // SPS
	H264_NAL_PPS                               // PPS
	H264_NAL_AUD                               // AUD
	H264_NAL_END_SEQUENCE                      // END_SEQUENCE
	H264_NAL_END_STREAM                        // END_STREAM
	H264_NAL_FILLER_DATA                       // FILLER_DATA
	H264_NAL_SPS_EXT                           // SPS_EXT
	H264_NAL_PREFIX                            // PREFIX
	H264_NAL_SUB_SPS                           // SUB_SPS
	H264_NAL_DPS                               // DPS
	H264_NAL_RESERVED17                        // RESERVED17
	H264_NAL_RESERVED18                        // RESERVED18
	H264_NAL_AUXILIARY_SLICE                   // AUXILIARY_SLICE
	H264_NAL_EXTEN_SLICE                       // EXTEN_SLICE
	H264_NAL_DEPTH_EXTEN_SLICE                 // DEPTH_EXTEN_SLICE
	H264_NAL_RESERVED22                        // RESERVED22
	H264_NAL_RESERVED23                        // RESERVED23
	H264_NAL_UNSPECIFIED24                     // UNSPECIFIED24
	H264_NAL_UNSPECIFIED25                     // UNSPECIFIED25
	H264_NAL_UNSPECIFIED26                     // UNSPECIFIED26
	H264_NAL_UNSPECIFIED27                     // UNSPECIFIED27
	H264_NAL_UNSPECIFIED28                     // UNSPECIFIED28
	H264_NAL_UNSPECIFIED29                     // UNSPECIFIED29
	H264_NAL_UNSPECIFIED30                     // UNSPECIFIED30
	H264_NAL_UNSPECIFIED31                     // UNSPECIFIED31
)

// -----------------------------------------------------------------------------.
//
//nolint:stylecheck // allow ALL_CAPS naming for H.265 NALU types
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
	// VCL types not used directly.
	HEVC_NAL_VCL_N10
	HEVC_NAL_VCL_R11
	HEVC_NAL_VCL_N12
	HEVC_NAL_VCL_R13
	HEVC_NAL_VCL_N14
	HEVC_NAL_VCL_R15
	// Random Access.
	HEVC_NAL_BLA_W_LP
	HEVC_NAL_BLA_W_RADL
	HEVC_NAL_BLA_N_LP
	HEVC_NAL_IDR_W_RADL
	HEVC_NAL_IDR_N_LP
	HEVC_NAL_CRA_NUT
	// Reserved IRAP.
	HEVC_NAL_RSV_IRAP_VCL22
	HEVC_NAL_RSV_IRAP_VCL23
	// Reserved VCL.
	HEVC_NAL_RSV_VCL24
	HEVC_NAL_RSV_VCL25
	HEVC_NAL_RSV_VCL26
	HEVC_NAL_RSV_VCL27
	HEVC_NAL_RSV_VCL28
	HEVC_NAL_RSV_VCL29
	HEVC_NAL_RSV_VCL30
	HEVC_NAL_RSV_VCL31
	// Parameter sets.
	HEVC_NAL_VPS
	HEVC_NAL_SPS
	HEVC_NAL_PPS
	// Others.
	HEVC_NAL_AUD
	HEVC_NAL_EOS_NUT
	HEVC_NAL_EOB_NUT
	HEVC_NAL_FD_NUT
	HEVC_NAL_SEI_PREFIX
	HEVC_NAL_SEI_SUFFIX
	// Reserved NVCL.
	HEVC_NAL_RSV_NVCL41
	HEVC_NAL_RSV_NVCL42
	HEVC_NAL_RSV_NVCL43
	HEVC_NAL_RSV_NVCL44
	HEVC_NAL_RSV_NVCL45
	HEVC_NAL_RSV_NVCL46
	HEVC_NAL_RSV_NVCL47
	// Unspecified.
	HEVC_NAL_UNSPEC48
	HEVC_NAL_UNSPEC49
	HEVC_NAL_UNSPEC50
	HEVC_NAL_UNSPEC51
	HEVC_NAL_UNSPEC52
	HEVC_NAL_UNSPEC53
	HEVC_NAL_UNSPEC54
	HEVC_NAL_UNSPEC55
	HEVC_NAL_UNSPEC56
	HEVC_NAL_UNSPEC57
	HEVC_NAL_UNSPEC58
	HEVC_NAL_UNSPEC59
	HEVC_NAL_UNSPEC60
	HEVC_NAL_UNSPEC61
	HEVC_NAL_UNSPEC62
	HEVC_NAL_UNSPEC63
)

// String returns a human-readable name for the NALU type.
//
//nolint:cyclop,funlen,gocyclo,maintidx
func (data NaluType) String(codecType CodecType) string {
	switch codecType {
	case H264:
		naluType := data & Last9BbitsNALUMask
		switch naluType { //nolint:exhaustive
		case H264_NAL_UNSPECIFIED:
			return "UNSPECIFIED"
		case H264_NAL_SLICE:
			return "NON_IDR_SLICE"
		case H264_NAL_DPA:
			return "DPA"
		case H264_NAL_DPB:
			return "DPB"
		case H264_NAL_DPC:
			return "DPC"
		case H264_NAL_IDR_SLICE:
			return "IDR_SLICE"
		case H264_NAL_SEI:
			return "SEI"
		case H264_NAL_SPS:
			return "SPS"
		case H264_NAL_PPS:
			return "PPS"
		case H264_NAL_AUD:
			return "AUD"
		case H264_NAL_END_SEQUENCE:
			return "END_SEQUENCE"
		case H264_NAL_END_STREAM:
			return "END_STREAM"
		case H264_NAL_FILLER_DATA:
			return "FILLER_DATA"
		case H264_NAL_SPS_EXT:
			return "SPS_EXT"
		case H264_NAL_PREFIX:
			return "PREFIX"
		case H264_NAL_SUB_SPS:
			return "SUB_SPS"
		case H264_NAL_DPS:
			return "DPS"
		case H264_NAL_RESERVED17:
			return "RESERVED17"
		case H264_NAL_RESERVED18:
			return "RESERVED18"
		case H264_NAL_AUXILIARY_SLICE:
			return "AUXILIARY_SLICE"
		case H264_NAL_EXTEN_SLICE:
			return "EXTEN_SLICE"
		case H264_NAL_DEPTH_EXTEN_SLICE:
			return "DEPTH_EXTEN_SLICE"
		case H264_NAL_RESERVED22:
			return "RESERVED22"
		case H264_NAL_RESERVED23:
			return "RESERVED23"
		case H264_NAL_UNSPECIFIED24:
			return "UNSPECIFIED24"
		case H264_NAL_UNSPECIFIED25:
			return "UNSPECIFIED25"
		case H264_NAL_UNSPECIFIED26:
			return "UNSPECIFIED26"
		case H264_NAL_UNSPECIFIED27:
			return "UNSPECIFIED27"
		case H264_NAL_UNSPECIFIED28:
			return "UNSPECIFIED28"
		case H264_NAL_UNSPECIFIED29:
			return "UNSPECIFIED29"
		case H264_NAL_UNSPECIFIED30:
			return "UNSPECIFIED30"
		case H264_NAL_UNSPECIFIED31:
			return "UNSPECIFIED31"
		default:
			return fmt.Sprintf("UNKNOWN(%d)", data)
		}

	case H265:
		naluType := (data >> 1) & Last10BbitsNALUMask
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

	default:
		return fmt.Sprintf("UNKNOWN(%d)", data)
	}
}
