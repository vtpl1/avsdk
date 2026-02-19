package av

import "fmt"

// H264NALTypeMask extracts the 5-bit nal_unit_type from the first byte of an H.264 NAL header
// (ITU-T H.264 §7.4.1: nal_unit_type occupies bits [4:0]).
const H264NALTypeMask H264NaluType = 0x1F

// H265NALTypeMask extracts the 6-bit nal_unit_type after a >>1 shift of the first byte of an
// H.265 NAL header (ITU-T H.265 §7.4.2.2: nal_unit_type occupies bits [14:9] of the 2-byte header;
// bits [6:1] of the first byte after the forbidden_zero_bit).
const H265NALTypeMask H265NaluType = 0x3F

// Extract with: H264NaluType(header[0]) & H264NALTypeMask.
type H264NaluType byte

// Extract with: H265NaluType(header[0]>>1) & H265NALTypeMask.
type H265NaluType byte

//nolint:stylecheck // ALL_CAPS names match the ITU-T H.264 specification identifiers
const (
	H264_NAL_UNSPECIFIED       H264NaluType = iota // 0
	H264_NAL_SLICE                                 // 1  non-IDR slice
	H264_NAL_DPA                                   // 2  data partition A
	H264_NAL_DPB                                   // 3  data partition B
	H264_NAL_DPC                                   // 4  data partition C
	H264_NAL_IDR_SLICE                             // 5  IDR slice (keyframe)
	H264_NAL_SEI                                   // 6  supplemental enhancement information
	H264_NAL_SPS                                   // 7  sequence parameter set
	H264_NAL_PPS                                   // 8  picture parameter set
	H264_NAL_AUD                                   // 9  access unit delimiter
	H264_NAL_END_SEQUENCE                          // 10
	H264_NAL_END_STREAM                            // 11
	H264_NAL_FILLER_DATA                           // 12
	H264_NAL_SPS_EXT                               // 13
	H264_NAL_PREFIX                                // 14
	H264_NAL_SUB_SPS                               // 15
	H264_NAL_DPS                                   // 16
	H264_NAL_RESERVED17                            // 17
	H264_NAL_RESERVED18                            // 18
	H264_NAL_AUXILIARY_SLICE                       // 19
	H264_NAL_EXTEN_SLICE                           // 20
	H264_NAL_DEPTH_EXTEN_SLICE                     // 21
	H264_NAL_RESERVED22                            // 22
	H264_NAL_RESERVED23                            // 23
	H264_NAL_UNSPECIFIED24                         // 24
	H264_NAL_UNSPECIFIED25                         // 25
	H264_NAL_UNSPECIFIED26                         // 26
	H264_NAL_UNSPECIFIED27                         // 27
	H264_NAL_UNSPECIFIED28                         // 28
	H264_NAL_UNSPECIFIED29                         // 29
	H264_NAL_UNSPECIFIED30                         // 30
	H264_NAL_UNSPECIFIED31                         // 31
)

//nolint:stylecheck // ALL_CAPS names match the ITU-T H.265 specification identifiers
const (
	HEVC_NAL_TRAIL_N        H265NaluType = iota // 0
	HEVC_NAL_TRAIL_R                            // 1
	HEVC_NAL_TSA_N                              // 2
	HEVC_NAL_TSA_R                              // 3
	HEVC_NAL_STSA_N                             // 4
	HEVC_NAL_STSA_R                             // 5
	HEVC_NAL_RADL_N                             // 6
	HEVC_NAL_RADL_R                             // 7
	HEVC_NAL_RASL_N                             // 8
	HEVC_NAL_RASL_R                             // 9
	HEVC_NAL_VCL_N10                            // 10
	HEVC_NAL_VCL_R11                            // 11
	HEVC_NAL_VCL_N12                            // 12
	HEVC_NAL_VCL_R13                            // 13
	HEVC_NAL_VCL_N14                            // 14
	HEVC_NAL_VCL_R15                            // 15
	HEVC_NAL_BLA_W_LP                           // 16
	HEVC_NAL_BLA_W_RADL                         // 17
	HEVC_NAL_BLA_N_LP                           // 18
	HEVC_NAL_IDR_W_RADL                         // 19
	HEVC_NAL_IDR_N_LP                           // 20
	HEVC_NAL_CRA_NUT                            // 21
	HEVC_NAL_RSV_IRAP_VCL22                     // 22
	HEVC_NAL_RSV_IRAP_VCL23                     // 23
	HEVC_NAL_RSV_VCL24                          // 24
	HEVC_NAL_RSV_VCL25                          // 25
	HEVC_NAL_RSV_VCL26                          // 26
	HEVC_NAL_RSV_VCL27                          // 27
	HEVC_NAL_RSV_VCL28                          // 28
	HEVC_NAL_RSV_VCL29                          // 29
	HEVC_NAL_RSV_VCL30                          // 30
	HEVC_NAL_RSV_VCL31                          // 31
	HEVC_NAL_VPS                                // 32
	HEVC_NAL_SPS                                // 33
	HEVC_NAL_PPS                                // 34
	HEVC_NAL_AUD                                // 35
	HEVC_NAL_EOS_NUT                            // 36
	HEVC_NAL_EOB_NUT                            // 37
	HEVC_NAL_FD_NUT                             // 38
	HEVC_NAL_SEI_PREFIX                         // 39
	HEVC_NAL_SEI_SUFFIX                         // 40
	HEVC_NAL_RSV_NVCL41                         // 41
	HEVC_NAL_RSV_NVCL42                         // 42
	HEVC_NAL_RSV_NVCL43                         // 43
	HEVC_NAL_RSV_NVCL44                         // 44
	HEVC_NAL_RSV_NVCL45                         // 45
	HEVC_NAL_RSV_NVCL46                         // 46
	HEVC_NAL_RSV_NVCL47                         // 47
	HEVC_NAL_UNSPEC48                           // 48
	HEVC_NAL_UNSPEC49                           // 49
	HEVC_NAL_UNSPEC50                           // 50
	HEVC_NAL_UNSPEC51                           // 51
	HEVC_NAL_UNSPEC52                           // 52
	HEVC_NAL_UNSPEC53                           // 53
	HEVC_NAL_UNSPEC54                           // 54
	HEVC_NAL_UNSPEC55                           // 55
	HEVC_NAL_UNSPEC56                           // 56
	HEVC_NAL_UNSPEC57                           // 57
	HEVC_NAL_UNSPEC58                           // 58
	HEVC_NAL_UNSPEC59                           // 59
	HEVC_NAL_UNSPEC60                           // 60
	HEVC_NAL_UNSPEC61                           // 61
	HEVC_NAL_UNSPEC62                           // 62
	HEVC_NAL_UNSPEC63                           // 63
)

// String implements fmt.Stringer for H.264 NALU types.
//
//nolint:exhaustive
func (t H264NaluType) String() string {
	switch t {
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
	default:
		return fmt.Sprintf("UNSPECIFIED(%d)", t)
	}
}

// String implements fmt.Stringer for H.265 NALU types.
//
//nolint:cyclop,funlen,gocyclo,exhaustive
func (t H265NaluType) String() string {
	switch t {
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
	default:
		return fmt.Sprintf("UNSPECIFIED(%d)", t)
	}
}
