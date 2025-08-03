// Package codec returns codecs from sdp
package codec

import (
	"encoding/base64"
	"strings"

	"github.com/pion/sdp/v3"
	"github.com/vtpl1/avsdk/av"
	"github.com/vtpl1/avsdk/codec/h264parser"
	"github.com/vtpl1/avsdk/codec/h265parser"
)

//nolint:gocognit,funlen, gocyclo, cyclop
func SdpToCodecs(s string) ([]av.CodecData, error) {
	sd := sdp.SessionDescription{}
	err := sd.UnmarshalString(s)

	var ret []av.CodecData

	if err != nil {
		return ret, err
	}

	for _, media := range sd.MediaDescriptions {
		mediaStr := media.MediaName.Media
		if mediaStr != "video" {
			continue
		}

		mediaTypeStr := ""

		field, ok := media.Attribute("rtpmap")
		if !ok {
			continue
		}

		keyval := strings.Split(field, " ")
		if len(keyval) >= 2 {
			field := strings.Split(keyval[1], "/")
			mediaTypeStr = field[0]
		}

		controlURL := ""
		field, ok = media.Attribute("control")
		if ok {
			controlURL = field
		}

		field, ok = media.Attribute("fmtp")
		if !ok {
			continue
		}

		var SpropVPS, SpropSPS, SpropPPS []byte

		keyval = strings.FieldsFunc(field, func(r rune) bool {
			return r == ' ' || r == ';'
		})

		for _, field := range keyval {
			keyval := strings.SplitN(field, "=", 2)
			if len(keyval) == 2 {
				key := strings.TrimSpace(keyval[0])
				val := keyval[1]
				switch key {
				case "sprop-vps":
					var valb []byte

					valb, err = base64.StdEncoding.DecodeString(val)
					if err != nil {
						continue
					}

					SpropVPS = valb
				case "sprop-sps":
					var valb []byte

					valb, err = base64.StdEncoding.DecodeString(val)
					if err != nil {
						continue
					}

					SpropSPS = valb
				case "sprop-pps":
					var valb []byte

					valb, err = base64.StdEncoding.DecodeString(val)
					if err != nil {
						continue
					}

					SpropPPS = valb
				case "sprop-parameter-sets":
					fields := strings.Split(val, ",")
					idx := 0

					for _, field := range fields {
						var valb []byte

						valb, err = base64.StdEncoding.DecodeString(field)
						if err != nil {
							continue
						}

						if mediaTypeStr == av.H264.String() {
							switch idx {
							case 0:
								SpropSPS = valb
							case 1:
								SpropPPS = valb
							}
						} else if mediaTypeStr == av.H265.String() {
							switch idx {
							case 0:
								SpropVPS = valb
							case 1:
								SpropSPS = valb
							case 2:
								SpropPPS = valb
							}
						}

						idx++
					}
				}
			}
		}

		if mediaTypeStr == av.H264.String() {
			var codecData h264parser.CodecData

			codecData, err = h264parser.NewCodecDataFromSPSAndPPS(SpropSPS, SpropPPS)
			if err != nil {
				continue
			}
			codecData.ControlURL = controlURL
			ret = append(ret, codecData)
		} else if mediaTypeStr == av.H265.String() {
			var codecData h265parser.CodecData
			codecData, err = h265parser.NewCodecDataFromVPSAndSPSAndPPS(SpropVPS, SpropSPS, SpropPPS)
			if err != nil {
				continue
			}
			codecData.ControlURL = controlURL

			ret = append(ret, codecData)
		}
	}

	return ret, err
}
