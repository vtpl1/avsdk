// Package h264parser holds Muxer and Demuxer for h264
package h264parser

import (
	"bytes"
	"errors"
	"fmt"
	"math"
	"strconv"
	"time"

	"github.com/vtpl1/avsdk/av"
	"github.com/vtpl1/avsdk/codec/parser"
	"github.com/vtpl1/avsdk/utils/bits"
	"github.com/vtpl1/avsdk/utils/bits/pio"
)

//nolint:dupword
/*
From: http://stackoverflow.com/questions/24884827/possible-locations-for-sequence-picture-parameter-sets-for-h-264-stream

First off, it's important to understand that there is no single standard H.264 elementary bitstream format.
The specification document does contain an Annex, specifically Annex B, that describes one possible format,
but it is not an actual requirement. The standard specifies how video is encoded into individual packets.
How these packets are stored and transmitted is left open to the integrator.

1. Annex B
Network Abstraction Layer Units
The packets are called Network Abstraction Layer Units. Often abbreviated NALU (or sometimes just NAL) each packet can be individually parsed and processed.
The first byte of each NALU contains the NALU type, specifically bits 3 through 7. (bit 0 is always off, and bits 1-2 indicate whether a NALU is referenced
by another NALU).

There are 19 different NALU types defined separated into two categories, VCL and non-VCL:

VCL, or Video Coding Layer packets contain the actual visual information.
Non-VCLs contain metadata that may or may not be required to decode the video.
A single NALU, or even a VCL NALU is NOT the same thing as a frame. A frame can be ‘sliced’ into several NALUs. Just like you can slice a pizza.
One or more slices are then virtually grouped into a Access Units (AU) that contain one frame. Slicing does come at a slight quality cost, so it is not
often used.

Below is a table of all defined NALUs.

0      Unspecified                                                    non-VCL
1      Coded slice of a non-IDR picture                               VCL
2      Coded slice data partition A                                   VCL
3      Coded slice data partition B                                   VCL
4      Coded slice data partition C                                   VCL
5      Coded slice of an IDR picture                                  VCL
6      Supplemental enhancement information (SEI)                     non-VCL
7      Sequence parameter set                                         non-VCL
8      Picture parameter set                                          non-VCL
9      Access unit delimiter                                          non-VCL
10     End of sequence                                                non-VCL
11     End of stream                                                  non-VCL
12     Filler data                                                    non-VCL
13     Sequence parameter set extension                               non-VCL
14     Prefix NAL unit                                                non-VCL
15     Subset sequence parameter set                                  non-VCL
16     Depth parameter set                                            non-VCL
17..18 Reserved                                                       non-VCL
19     Coded slice of an auxiliary coded picture without partitioning non-VCL
20     Coded slice extension                                          non-VCL
21     Coded slice extension for depth view components                non-VCL
22..23 Reserved                                                       non-VCL
24..31 Unspecified                                                    non-VCL
There are a couple of NALU types where having knowledge of may be helpful later.

Sequence Parameter Set (SPS). This non-VCL NALU contains information required to configure the decoder such as profile, level, resolution, frame rate.
Picture Parameter Set (PPS). Similar to the SPS, this non-VCL contains information on entropy coding mode, slice groups, motion prediction and deblocking
filters.
Instantaneous Decoder Refresh (IDR). This VCL NALU is a s contained image slice. That is, an IDR can be decoded and displayed without referencing any other
NALU save SPS and PPS.
Access Unit Delimiter (AUD). An AUD is an optional NALU that can be use to delimit frames in an elementary stream.
It is not required (unless otherwise stated by the container/protocol, like TS), and is often not included in order to save space,
but it can be useful to finds the start of a frame without having to fully parse each NALU.
NALU Start Codes
A NALU does not contain is its size. Therefore simply concatenating the NALUs to create a stream will not work because you will not know where one stops and
the next begins.

The Annex B specification solves this by requiring ‘Start Codes’ to precede each NALU. A start code is 2 or 3 0x00 bytes followed with a 0x01 byte.
e.g. 0x000001 or 0x00000001.

The 4 byte variation is useful for transmission over a serial connection as it is trivial to byte align the stream by looking for 31 zero bits followed by a
one.
If the next bit is 0 (because every NALU starts with a 0 bit), it is the start of a NALU.
The 4 byte variation is usually only used for signalling random access points in the stream such as a SPS PPS AUD and IDR Where as the 3 byte variation is
used everywhere else to save space.

Emulation Prevention Bytes
Start codes work because the four byte sequences 0x000000, 0x000001, 0x000002 and 0x000003 are illegal within a non-RBSP NALU.
So when creating a NALU, care is taken to escape these values that could otherwise be confused with a start code.
This is accomplished by inserting an ‘Emulation Prevention’ byte 0x03, so that 0x000001 becomes 0x00000301.

When decoding, it is important to look for and ignore emulation prevention bytes.
Because emulation prevention bytes can occur almost anywhere within a NALU, it is often more convenient in documentation to assume they have already been
removed.
A representation without emulation prevention bytes is called Raw Byte Sequence Payload (RBSP).

Example
Let's look at a complete example.

0x0000 | 00 00 00 01 67 64 00 0A AC 72 84 44 26 84 00 00
0x0010 | 03 00 04 00 00 03 00 CA 3C 48 96 11 80 00 00 00
0x0020 | 01 68 E8 43 8F 13 21 30 00 00 01 65 88 81 00 05
0x0030 | 4E 7F 87 DF 61 A5 8B 95 EE A4 E9 38 B7 6A 30 6A
0x0040 | 71 B9 55 60 0B 76 2E B5 0E E4 80 59 27 B8 67 A9
0x0050 | 63 37 5E 82 20 55 FB E4 6A E9 37 35 72 E2 22 91
0x0060 | 9E 4D FF 60 86 CE 7E 42 B7 95 CE 2A E1 26 BE 87
0x0070 | 73 84 26 BA 16 36 F4 E6 9F 17 DA D8 64 75 54 B1
0x0080 | F3 45 0C 0B 3C 74 B3 9D BC EB 53 73 87 C3 0E 62
0x0090 | 47 48 62 CA 59 EB 86 3F 3A FA 86 B5 BF A8 6D 06
0x00A0 | 16 50 82 C4 CE 62 9E 4E E6 4C C7 30 3E DE A1 0B
0x00B0 | D8 83 0B B6 B8 28 BC A9 EB 77 43 FC 7A 17 94 85
0x00C0 | 21 CA 37 6B 30 95 B5 46 77 30 60 B7 12 D6 8C C5
0x00D0 | 54 85 29 D8 69 A9 6F 12 4E 71 DF E3 E2 B1 6B 6B
0x00E0 | BF 9F FB 2E 57 30 A9 69 76 C4 46 A2 DF FA 91 D9
0x00F0 | 50 74 55 1D 49 04 5A 1C D6 86 68 7C B6 61 48 6C
0x0100 | 96 E6 12 4C 27 AD BA C7 51 99 8E D0 F0 ED 8E F6
0x0110 | 65 79 79 A6 12 A1 95 DB C8 AE E3 B6 35 E6 8D BC
0x0120 | 48 A3 7F AF 4A 28 8A 53 E2 7E 68 08 9F 67 77 98
0x0130 | 52 DB 50 84 D6 5E 25 E1 4A 99 58 34 C7 11 D6 43
0x0140 | FF C4 FD 9A 44 16 D1 B2 FB 02 DB A1 89 69 34 C2
0x0150 | 32 55 98 F9 9B B2 31 3F 49 59 0C 06 8C DB A5 B2
0x0160 | 9D 7E 12 2F D0 87 94 44 E4 0A 76 EF 99 2D 91 18
0x0170 | 39 50 3B 29 3B F5 2C 97 73 48 91 83 B0 A6 F3 4B
0x0180 | 70 2F 1C 8F 3B 78 23 C6 AA 86 46 43 1D D7 2A 23
0x0190 | 5E 2C D9 48 0A F5 F5 2C D1 FB 3F F0 4B 78 37 E9
0x01A0 | 45 DD 72 CF 80 35 C3 95 07 F3 D9 06 E5 4A 58 76
0x01B0 | 03 6C 81 20 62 45 65 44 73 BC FE C1 9F 31 E5 DB
0x01C0 | 89 5C 6B 79 D8 68 90 D7 26 A8 A1 88 86 81 DC 9A
0x01D0 | 4F 40 A5 23 C7 DE BE 6F 76 AB 79 16 51 21 67 83
0x01E0 | 2E F3 D6 27 1A 42 C2 94 D1 5D 6C DB 4A 7A E2 CB
0x01F0 | 0B B0 68 0B BE 19 59 00 50 FC C0 BD 9D F5 F5 F8
0x0200 | A8 17 19 D6 B3 E9 74 BA 50 E5 2C 45 7B F9 93 EA
0x0210 | 5A F9 A9 30 B1 6F 5B 36 24 1E 8D 55 57 F4 CC 67
0x0220 | B2 65 6A A9 36 26 D0 06 B8 E2 E3 73 8B D1 C0 1C
0x0230 | 52 15 CA B5 AC 60 3E 36 42 F1 2C BD 99 77 AB A8
0x0240 | A9 A4 8E 9C 8B 84 DE 73 F0 91 29 97 AE DB AF D6
0x0250 | F8 5E 9B 86 B3 B3 03 B3 AC 75 6F A6 11 69 2F 3D
0x0260 | 3A CE FA 53 86 60 95 6C BB C5 4E F3

This is a complete AU containing 3 NALUs. As you can see, we begin with a Start code followed by an SPS (SPS starts with 67).
Within the SPS, you will see two Emulation Prevention bytes. Without these bytes the illegal sequence 0x000000 would occur at these positions.
Next you will see a start code followed by a PPS (PPS starts with 68) and one final start code followed by an IDR slice. This is a complete H.264 stream.
If you type these values into a hex editor and save the file with a .264 extension, you will be able to convert it to this image:

Lena

Annex B is commonly used in live and streaming formats such as transport streams, over the air broadcasts, and DVDs.
In these formats it is common to repeat the SPS and PPS periodically, usually preceding every IDR thus creating a random access point for the decoder.
This enables the ability to join a stream already in progress.

2. AVCC
The other common method of storing an H.264 stream is the AVCC format. In this format, each NALU is preceded with its length (in big endian format).
This method is easier to parse, but you lose the byte alignment features of Annex B. Just to complicate things, the length may be encoded using 1, 2 or 4 bytes.
This value is stored in a header object. This header is often called ‘extradata’ or ‘sequence header’. Its basic format is as follows:

bits
8   version ( always 0x01 )
8   avc profile ( sps[0][1] )
8   avc compatibility ( sps[0][2] )
8   avc level ( sps[0][3] )
6   reserved ( all bits on )
2   NALULengthSizeMinusOne
3   reserved ( all bits on )
5   number of SPS NALUs (usually 1)
repeated once per SPS:
  16         SPS size
  variable   SPS NALU data

  8   number of PPS NALUs (usually 1)

repeated once per PPS
  16         PPS size
  variable   PPS NALU data

Using the same example above, the AVCC extradata will look like this:

0x0000 | 01 64 00 0A FF E1 00 19 67 64 00 0A AC 72 84 44
0x0010 | 26 84 00 00 03 00 04 00 00 03 00 CA 3C 48 96 11
0x0020 | 80 01 00 07 68 E8 43 8F 13 21 30

You will notice SPS and PPS is now stored out of band. That is, separate from the elementary stream data.
Storage and transmission of this data is the job of the file container, and beyond the scope of this document.
Notice that even though we are not using start codes, emulation prevention bytes are still inserted.

Additionally, there is a new variable called NALULengthSizeMinusOne. This confusingly named variable tells us how many bytes to use to store the length of
each NALU.
So, if NALULengthSizeMinusOne is set to 0, then each NALU is preceded with a single byte indicating its length. Using a single byte to store the size, the
max size of a NALU is 255 bytes.
That is obviously pretty small. Way too small for an entire key frame. Using 2 bytes gives us 64k per NALU. It would work in our example, but is still a
pretty low limit.
3 bytes would be perfect, but for some reason is not universally supported. Therefore, 4 bytes is by far the most common, and it is what we used here:

0x0000 | 00 00 02 41 65 88 81 00 05 4E 7F 87 DF 61 A5 8B
0x0010 | 95 EE A4 E9 38 B7 6A 30 6A 71 B9 55 60 0B 76 2E
0x0020 | B5 0E E4 80 59 27 B8 67 A9 63 37 5E 82 20 55 FB
0x0030 | E4 6A E9 37 35 72 E2 22 91 9E 4D FF 60 86 CE 7E
0x0040 | 42 B7 95 CE 2A E1 26 BE 87 73 84 26 BA 16 36 F4
0x0050 | E6 9F 17 DA D8 64 75 54 B1 F3 45 0C 0B 3C 74 B3
0x0060 | 9D BC EB 53 73 87 C3 0E 62 47 48 62 CA 59 EB 86
0x0070 | 3F 3A FA 86 B5 BF A8 6D 06 16 50 82 C4 CE 62 9E
0x0080 | 4E E6 4C C7 30 3E DE A1 0B D8 83 0B B6 B8 28 BC
0x0090 | A9 EB 77 43 FC 7A 17 94 85 21 CA 37 6B 30 95 B5
0x00A0 | 46 77 30 60 B7 12 D6 8C C5 54 85 29 D8 69 A9 6F
0x00B0 | 12 4E 71 DF E3 E2 B1 6B 6B BF 9F FB 2E 57 30 A9
0x00C0 | 69 76 C4 46 A2 DF FA 91 D9 50 74 55 1D 49 04 5A
0x00D0 | 1C D6 86 68 7C B6 61 48 6C 96 E6 12 4C 27 AD BA
0x00E0 | C7 51 99 8E D0 F0 ED 8E F6 65 79 79 A6 12 A1 95
0x00F0 | DB C8 AE E3 B6 35 E6 8D BC 48 A3 7F AF 4A 28 8A
0x0100 | 53 E2 7E 68 08 9F 67 77 98 52 DB 50 84 D6 5E 25
0x0110 | E1 4A 99 58 34 C7 11 D6 43 FF C4 FD 9A 44 16 D1
0x0120 | B2 FB 02 DB A1 89 69 34 C2 32 55 98 F9 9B B2 31
0x0130 | 3F 49 59 0C 06 8C DB A5 B2 9D 7E 12 2F D0 87 94
0x0140 | 44 E4 0A 76 EF 99 2D 91 18 39 50 3B 29 3B F5 2C
0x0150 | 97 73 48 91 83 B0 A6 F3 4B 70 2F 1C 8F 3B 78 23
0x0160 | C6 AA 86 46 43 1D D7 2A 23 5E 2C D9 48 0A F5 F5
0x0170 | 2C D1 FB 3F F0 4B 78 37 E9 45 DD 72 CF 80 35 C3
0x0180 | 95 07 F3 D9 06 E5 4A 58 76 03 6C 81 20 62 45 65
0x0190 | 44 73 BC FE C1 9F 31 E5 DB 89 5C 6B 79 D8 68 90
0x01A0 | D7 26 A8 A1 88 86 81 DC 9A 4F 40 A5 23 C7 DE BE
0x01B0 | 6F 76 AB 79 16 51 21 67 83 2E F3 D6 27 1A 42 C2
0x01C0 | 94 D1 5D 6C DB 4A 7A E2 CB 0B B0 68 0B BE 19 59
0x01D0 | 00 50 FC C0 BD 9D F5 F5 F8 A8 17 19 D6 B3 E9 74
0x01E0 | BA 50 E5 2C 45 7B F9 93 EA 5A F9 A9 30 B1 6F 5B
0x01F0 | 36 24 1E 8D 55 57 F4 CC 67 B2 65 6A A9 36 26 D0
0x0200 | 06 B8 E2 E3 73 8B D1 C0 1C 52 15 CA B5 AC 60 3E
0x0210 | 36 42 F1 2C BD 99 77 AB A8 A9 A4 8E 9C 8B 84 DE
0x0220 | 73 F0 91 29 97 AE DB AF D6 F8 5E 9B 86 B3 B3 03
0x0230 | B3 AC 75 6F A6 11 69 2F 3D 3A CE FA 53 86 60 95
0x0240 | 6C BB C5 4E F3

An advantage to this format is the ability to configure the decoder at the start and jump into the middle of a stream.
This is a common use case where the media is available on a random access medium such as a hard drive,
and is therefore used in common container formats such as MP4 and MKV.
*/

var (
	StartCodeBytes = []byte{0, 0, 1}                           //nolint:gochecknoglobals
	AUDBytes       = []byte{0, 0, 0, 1, 0x9, 0xf0, 0, 0, 0, 1} //nolint:gochecknoglobals // AUD
)

type NaluType byte

const (
	H264_NAL_UNSPECIFIED NaluType = iota
	// NALU_NON_IDR - Non-IDR Slice NAL unit.
	H264_NAL_SLICE
	H264_NAL_DPA
	H264_NAL_DPB
	H264_NAL_DPC
	// NALU_IDR - IDR Random Access Slice NAL Unit.
	H264_NAL_IDR_SLICE
	// NALU_SEI - Supplementary Enhancement Information NAL Unit.
	H264_NAL_SEI
	// NALU_SPS - SequenceParameterSet NAL Unit.
	H264_NAL_SPS
	// NALU_PPS - PictureParameterSet NAL Unit.
	H264_NAL_PPS
	// NALU_AUD - AccessUnitDelimiter NAL Unit.
	H264_NAL_AUD
	// NALU_EO_SEQ - End of Sequence NAL Unit.
	H264_NAL_END_SEQUENCE
	// NALU_EO_STREAM - End of Stream NAL Unit.
	H264_NAL_END_STREAM
	// NALU_FILL - Filler NAL Unit.
	H264_NAL_FILLER_DATA
	H264_NAL_SPS_EXT
	H264_NAL_PREFIX
	H264_NAL_SUB_SPS
	H264_NAL_DPS
	H264_NAL_RESERVED17
	H264_NAL_RESERVED18
	H264_NAL_AUXILIARY_SLICE
	H264_NAL_EXTEN_SLICE
	H264_NAL_DEPTH_EXTEN_SLICE
	H264_NAL_RESERVED22
	H264_NAL_RESERVED23
	H264_NAL_UNSPECIFIED24
	H264_NAL_UNSPECIFIED25
	H264_NAL_UNSPECIFIED26
	H264_NAL_UNSPECIFIED27
	H264_NAL_UNSPECIFIED28
	H264_NAL_UNSPECIFIED29
	H264_NAL_UNSPECIFIED30
	H264_NAL_UNSPECIFIED31
)

const bitsInByte = 8

func IsKeyFrame(nalHeader []byte) bool {
	typ := nalHeader[0] & parser.Last9BbitsNALUMask

	return typ == byte(H264_NAL_IDR_SLICE)
}

func IsDataNALU(nalHeader []byte) bool {
	typ := nalHeader[0] & parser.Last9BbitsNALUMask

	return typ >= byte(H264_NAL_SLICE) && typ <= byte(H264_NAL_IDR_SLICE)
}

func IsSPSNALU(nalHeader []byte) bool {
	typ := nalHeader[0] & parser.Last9BbitsNALUMask

	return typ == byte(H264_NAL_SPS)
}

func IsPPSNALU(nalHeader []byte) bool {
	typ := nalHeader[0] & parser.Last9BbitsNALUMask

	return typ == byte(H264_NAL_PPS)
}

func CheckNALUsType(b []byte) parser.NALUAvccOrAnnexb {
	_, typ := parser.SplitNALUs(b)

	return typ
}

type SPSInfo struct {
	ID                uint
	ProfileIdc        uint
	LevelIdc          uint
	ConstraintSetFlag uint

	MbWidth  uint
	MbHeight uint

	CropLeft   uint
	CropRight  uint
	CropTop    uint
	CropBottom uint

	Width  uint
	Height uint
	FPS    uint
}

func RemoveH264orH265EmulationBytes(b []byte) []byte {
	j := 0
	r := make([]byte, len(b))

	for i := 0; (i < len(b)) && (j < len(b)); {
		if i+2 < len(b) &&
			b[i] == 0 && b[i+1] == 0 && b[i+2] == 3 {
			r[j] = 0
			r[j+1] = 0
			j += 2
			i += 3
		} else {
			r[j] = b[i]
			j++
			i++
		}
	}

	return r[:j]
}

//nolint:gocyclo,cyclop,funlen,gocognit,maintidx
func ParseSPS(data []byte) (SPSInfo, error) {
	data = RemoveH264orH265EmulationBytes(data)
	r := &bits.GolombBitReader{R: bytes.NewReader(data)}

	var s SPSInfo

	var err error
	if _, err = r.ReadBits(bitsInByte); err != nil {
		return s, err
	}

	if s.ProfileIdc, err = r.ReadBits(bitsInByte); err != nil {
		return s, err
	}

	// constraint_set0_flag-constraint_set6_flag,reserved_zero_2bits
	if s.ConstraintSetFlag, err = r.ReadBits(bitsInByte); err != nil {
		return s, err
	}

	s.ConstraintSetFlag >>= 2

	// level_idc
	if s.LevelIdc, err = r.ReadBits(bitsInByte); err != nil {
		return s, err
	}

	// seq_parameter_set_id
	if s.ID, err = r.ReadExponentialGolombCode(); err != nil {
		return s, err
	}

	if s.ProfileIdc == 100 || s.ProfileIdc == 110 ||
		s.ProfileIdc == 122 || s.ProfileIdc == 244 ||
		s.ProfileIdc == 44 || s.ProfileIdc == 83 ||
		s.ProfileIdc == 86 || s.ProfileIdc == 118 {
		var chromaFormatIdc uint

		if chromaFormatIdc, err = r.ReadExponentialGolombCode(); err != nil {
			return s, err
		}

		if chromaFormatIdc == 3 {
			// residual_colour_transform_flag
			if _, err = r.ReadBit(); err != nil {
				return s, err
			}
		}

		// bit_depth_luma_minus8
		if _, err = r.ReadExponentialGolombCode(); err != nil {
			return s, err
		}
		// bit_depth_chroma_minus8
		if _, err = r.ReadExponentialGolombCode(); err != nil {
			return s, err
		}
		// qpprime_y_zero_transform_bypass_flag
		if _, err = r.ReadBit(); err != nil {
			return s, err
		}

		var seqScalingMatrixPresentFlag uint

		if seqScalingMatrixPresentFlag, err = r.ReadBit(); err != nil {
			return s, err
		}

		if seqScalingMatrixPresentFlag != 0 {
			for i := range 8 {
				var seqScalingListPresentFlag uint

				if seqScalingListPresentFlag, err = r.ReadBit(); err != nil {
					return s, err
				}

				if seqScalingListPresentFlag != 0 {
					var sizeOfScalingList uint
					if i < 6 {
						sizeOfScalingList = 16
					} else {
						sizeOfScalingList = 64
					}

					lastScale := uint(8)

					nextScale := uint(8)
					for range sizeOfScalingList {
						if nextScale != 0 {
							var deltaScale uint

							if deltaScale, err = r.ReadSE(); err != nil {
								return s, err
							}

							nextScale = (lastScale + deltaScale + 256) % 256
						}

						if nextScale != 0 {
							lastScale = nextScale
						}
					}
				}
			}
		}
	}

	// log2_max_frame_num_minus4
	if _, err = r.ReadExponentialGolombCode(); err != nil {
		return s, err
	}

	var picOrderCntType uint

	if picOrderCntType, err = r.ReadExponentialGolombCode(); err != nil {
		return s, err
	}

	switch picOrderCntType {
	case 0:
		// log2_max_pic_order_cnt_lsb_minus4
		if _, err = r.ReadExponentialGolombCode(); err != nil {
			return s, err
		}
	case 1:
		// delta_pic_order_always_zero_flag
		if _, err = r.ReadBit(); err != nil {
			return s, err
		}
		// offset_for_non_ref_pic
		if _, err = r.ReadSE(); err != nil {
			return s, err
		}
		// offset_for_top_to_bottom_field
		if _, err = r.ReadSE(); err != nil {
			return s, err
		}

		var numRefFramesInPicOrderCntCycle uint

		if numRefFramesInPicOrderCntCycle, err = r.ReadExponentialGolombCode(); err != nil {
			return s, err
		}

		for range numRefFramesInPicOrderCntCycle {
			if _, err = r.ReadSE(); err != nil {
				return s, err
			}
		}
	}

	// max_num_ref_frames
	if _, err = r.ReadExponentialGolombCode(); err != nil {
		return s, err
	}

	// gaps_in_frame_num_value_allowed_flag
	if _, err = r.ReadBit(); err != nil {
		return s, err
	}

	if s.MbWidth, err = r.ReadExponentialGolombCode(); err != nil {
		return s, err
	}

	s.MbWidth++

	if s.MbHeight, err = r.ReadExponentialGolombCode(); err != nil {
		return s, err
	}

	s.MbHeight++

	var frameMbsOnlyFlag uint

	if frameMbsOnlyFlag, err = r.ReadBit(); err != nil {
		return s, err
	}

	if frameMbsOnlyFlag == 0 {
		// mb_adaptive_frame_field_flag
		if _, err = r.ReadBit(); err != nil {
			return s, err
		}
	}

	// direct_8x8_inference_flag
	if _, err = r.ReadBit(); err != nil {
		return s, err
	}

	var frameCroppingFlag uint

	if frameCroppingFlag, err = r.ReadBit(); err != nil {
		return s, err
	}

	if frameCroppingFlag != 0 {
		if s.CropLeft, err = r.ReadExponentialGolombCode(); err != nil {
			return s, err
		}

		if s.CropRight, err = r.ReadExponentialGolombCode(); err != nil {
			return s, err
		}

		if s.CropTop, err = r.ReadExponentialGolombCode(); err != nil {
			return s, err
		}

		if s.CropBottom, err = r.ReadExponentialGolombCode(); err != nil {
			return s, err
		}
	}

	s.Width = (s.MbWidth * 16) - s.CropLeft*2 - s.CropRight*2
	s.Height = ((2 - frameMbsOnlyFlag) * s.MbHeight * 16) - s.CropTop*2 - s.CropBottom*2

	var vuiParameterPresentFlag uint

	vuiParameterPresentFlag, err = r.ReadBit()
	if err != nil {
		return s, err
	}

	if vuiParameterPresentFlag != 0 {
		var aspectRatioInfoPresentFlag uint

		aspectRatioInfoPresentFlag, err = r.ReadBit()
		if err != nil {
			return s, err
		}

		if aspectRatioInfoPresentFlag != 0 {
			var aspectRatioIdc uint

			aspectRatioIdc, err = r.ReadBits(8)
			if err != nil {
				return s, err
			}

			if aspectRatioIdc == 255 {
				var sarWidth uint

				sarWidth, err = r.ReadBits(16)
				if err != nil {
					return s, err
				}

				var sarHeight uint

				sarHeight, err = r.ReadBits(16)
				if err != nil {
					return s, err
				}

				_, _ = sarWidth, sarHeight
			}
		}

		var overscanInfoPresentFlag uint

		overscanInfoPresentFlag, err = r.ReadBit()
		if err != nil {
			return s, err
		}

		if overscanInfoPresentFlag != 0 {
			var overscanAppropriateFlagu uint

			overscanAppropriateFlagu, err = r.ReadBit()
			if err != nil {
				return s, err
			}

			_ = overscanAppropriateFlagu
		}

		var videoSignalTypePresentFlag uint

		videoSignalTypePresentFlag, err = r.ReadBit()
		if err != nil {
			return s, err
		}

		if videoSignalTypePresentFlag != 0 {
			var videoFormat uint

			videoFormat, err = r.ReadBits(3)
			if err != nil {
				return s, err
			}

			_ = videoFormat

			var videoFullRangeFlag uint

			videoFullRangeFlag, err = r.ReadBit()
			if err != nil {
				return s, err
			}

			_ = videoFullRangeFlag

			var colourDescriptionPresentFlag uint

			colourDescriptionPresentFlag, err = r.ReadBit()
			if err != nil {
				return s, err
			}

			if colourDescriptionPresentFlag != 0 {
				var colourPrimaries uint

				colourPrimaries, err = r.ReadBits(8)
				if err != nil {
					return s, err
				}

				_ = colourPrimaries

				var transferCharacteristics uint

				transferCharacteristics, err = r.ReadBits(8)
				if err != nil {
					return s, err
				}

				_ = transferCharacteristics

				var matrixCoefficients uint

				matrixCoefficients, err = r.ReadBits(8)
				if err != nil {
					return s, err
				}

				_ = matrixCoefficients
			}
		}

		var chromaLocInfoPresentFlag uint

		chromaLocInfoPresentFlag, err = r.ReadBit()
		if err != nil {
			return s, err
		}

		if chromaLocInfoPresentFlag != 0 {
			var chromaSampleLocTypeTopField uint

			chromaSampleLocTypeTopField, err = r.ReadSE()
			if err != nil {
				return s, err
			}

			_ = chromaSampleLocTypeTopField

			var chromaSampleLocTypeBottomField uint

			chromaSampleLocTypeBottomField, err = r.ReadSE()
			if err != nil {
				return s, err
			}

			_ = chromaSampleLocTypeBottomField
		}

		var timingInfoPresentFlag uint

		timingInfoPresentFlag, err = r.ReadBit()
		if err != nil {
			return s, err
		}

		if timingInfoPresentFlag != 0 {
			var numUnitsInTick uint

			numUnitsInTick, err = r.ReadBits(32)
			if err != nil {
				return s, err
			}

			var timeScale uint

			timeScale, err = r.ReadBits(32)
			if err != nil {
				return s, err
			}

			s.FPS = uint(math.Floor(float64(timeScale) / float64(numUnitsInTick) / 2.0))

			var fixedFrameRateFlag uint

			fixedFrameRateFlag, err = r.ReadBit()
			if err != nil {
				return s, err
			}

			_ = fixedFrameRateFlag
			// if fixedFrameRateFlag != 0 {
			// 	// utils.L.InfoLn("fixed_frame_rate_flag", fixed_frame_rate_flag)
			// 	// have been divide 2
			// 	// s.FPS = s.FPS / 2
			// }
		}
	}

	return s, err
}

type CodecData struct {
	Record     []byte
	RecordInfo AVCDecoderConfRecord
	SPSInfo    SPSInfo
	ControlURL string
}

func (s CodecData) Type() av.CodecType {
	return av.H264
}

func (s CodecData) AVCDecoderConfRecordBytes() []byte {
	return s.Record
}

func (s CodecData) SPS() []byte {
	if len(s.RecordInfo.SPS) > 0 {
		return s.RecordInfo.SPS[0]
	}

	return []byte{0}
}

func (s CodecData) PPS() []byte {
	if len(s.RecordInfo.PPS) > 0 {
		return s.RecordInfo.PPS[0]
	}

	return []byte{0}
}

func (s CodecData) Width() int {
	return int(s.SPSInfo.Width)
}

func (s CodecData) Height() int {
	return int(s.SPSInfo.Height)
}

func (s CodecData) FPS() int {
	return int(s.SPSInfo.FPS)
}

func (s CodecData) TrackID() string {
	return s.ControlURL
}

func (s CodecData) Resolution() string {
	return fmt.Sprintf("%vx%v", s.Width(), s.Height())
}

func (s CodecData) Tag() string {
	return fmt.Sprintf("avc1.%02X%02X%02X", s.RecordInfo.AVCProfileIndication, s.RecordInfo.ProfileCompatibility, s.RecordInfo.AVCLevelIndication)
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

	if s.SPSInfo, err = ParseSPS(s.RecordInfo.SPS[0]); err != nil {
		err = errors.Join(ErrSPSNotFound, err)

		return s, err
	}

	return s, err
}

func NewCodecDataFromSPSAndPPS(sps, pps []byte) (CodecData, error) {
	var s CodecData

	var err error

	recordinfo := AVCDecoderConfRecord{}
	recordinfo.AVCProfileIndication = sps[1]
	recordinfo.ProfileCompatibility = sps[2]
	recordinfo.AVCLevelIndication = sps[3]
	recordinfo.SPS = [][]byte{sps}
	recordinfo.PPS = [][]byte{pps}
	recordinfo.LengthSizeMinusOne = 3

	buf := make([]byte, recordinfo.Len())
	recordinfo.Marshal(buf)

	s.RecordInfo = recordinfo
	s.Record = buf

	if s.SPSInfo, err = ParseSPS(sps); err != nil {
		return s, err
	}

	return s, err
}

type AVCDecoderConfRecord struct {
	AVCProfileIndication uint8
	ProfileCompatibility uint8
	AVCLevelIndication   uint8
	LengthSizeMinusOne   uint8
	SPS                  [][]byte
	PPS                  [][]byte
}

func (s *AVCDecoderConfRecord) Unmarshal(b []byte) (int, error) {
	var n int

	if len(b) < 7 {
		return n, ErrDecconfInvalid
	}

	s.AVCProfileIndication = b[1]
	s.ProfileCompatibility = b[2]
	s.AVCLevelIndication = b[3]
	s.LengthSizeMinusOne = b[4] & 0x03
	spscount := int(b[5] & 0x1f)
	n += 6

	for range spscount {
		if len(b) < n+2 {
			return n, ErrDecconfInvalid
		}

		spslen := int(pio.U16BE(b[n:]))
		n += 2

		if len(b) < n+spslen {
			return n, ErrDecconfInvalid
		}

		s.SPS = append(s.SPS, b[n:n+spslen])
		n += spslen
	}

	if len(b) < n+1 {
		return n, ErrDecconfInvalid
	}

	ppscount := int(b[n])
	n++

	for range ppscount {
		if len(b) < n+2 {
			return n, ErrDecconfInvalid
		}

		ppslen := int(pio.U16BE(b[n:]))
		n += 2

		if len(b) < n+ppslen {
			return n, ErrDecconfInvalid
		}

		s.PPS = append(s.PPS, b[n:n+ppslen])
		n += ppslen
	}

	return n, nil
}

func (s *AVCDecoderConfRecord) Len() int {
	n := 7
	for _, sps := range s.SPS {
		n += 2 + len(sps)
	}

	for _, pps := range s.PPS {
		n += 2 + len(pps)
	}

	return n
}

func (s *AVCDecoderConfRecord) Marshal(b []byte) int {
	n := 0
	b[0] = 1
	b[1] = s.AVCProfileIndication
	b[2] = s.ProfileCompatibility
	b[3] = s.AVCLevelIndication
	b[4] = s.LengthSizeMinusOne | 0xfc
	b[5] = uint8(len(s.SPS)) | 0xe0
	n += 6

	for _, sps := range s.SPS {
		pio.PutU16BE(b[n:], uint16(len(sps)))
		n += 2
		copy(b[n:], sps)
		n += len(sps)
	}

	b[n] = uint8(len(s.PPS))
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
	SliceP = iota + 1
	SliceB
	SliceI
)

func ParseSliceHeaderFromNALU(pkt []byte) (SliceType, error) {
	var sliceType SliceType

	var err error

	if len(pkt) <= 1 {
		return sliceType, ErrPacketTooShort
	}

	nalUnitType := pkt[0] & 0x1f
	switch nalUnitType {
	case 1, 2, 5, 19:
		// slice_layer_without_partitioning_rbsp
		// slice_data_partition_a_layer_rbsp

	default:
		return sliceType, ErrNalHasNoSliceHeader
	}

	r := &bits.GolombBitReader{R: bytes.NewReader(pkt[1:])}

	// first_mb_in_slice
	if _, err = r.ReadExponentialGolombCode(); err != nil {
		return sliceType, err
	}

	// slice_type
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
		return sliceType, ErrInvalidSliceType
	}

	return sliceType, err
}
