// Package codec returns codecs from sdp
package codec_test

import (
	"fmt"
	"testing"

	"github.com/vtpl1/avsdk/av"
	"github.com/vtpl1/avsdk/codec"
	"github.com/vtpl1/avsdk/codec/h264parser"
	"github.com/vtpl1/avsdk/codec/h265parser"
	"github.com/vtpl1/avsdk/codec/pcm"
)

const MPEG4UnmarshalSDP = "v=0\r\n" +
	"o=- 1459325504777324 1 IN IP4 192.168.0.123\r\n" +
	"s=RTSP/RTP stream from Network Video Server\r\n" +
	"i=mpeg4cif\r\n" +
	"t=0 0\r\n" +
	"a=tool:LIVE555 Streaming Media v2009.09.28\r\n" +
	"a=type:broadcast\r\n" +
	"a=control:*\r\n" +
	"a=range:npt=0-\r\n" +
	"a=x-qt-text-nam:RTSP/RTP stream from Network Video Server\r\n" +
	"a=x-qt-text-inf:mpeg4cif\r\n" +
	"m=video 0 RTP/AVP 96\r\n" +
	"c=IN IP4 0.0.0.0\r\n" +
	"b=AS:300\r\n" +
	"a=rtpmap:96 H264/90000\r\n" +
	"a=fmtp:96 profile-level-id=420029; packetization-mode=1; sprop-parameter-sets=Z00AHpWoKA9k,aO48gA==\r\n" +
	"a=x-dimensions: 720, 480\r\n" +
	"a=x-framerate: 15\r\n" +
	"a=control:track1\r\n" +
	"m=audio 0 RTP/AVP 96\r\n" +
	"c=IN IP4 0.0.0.0\r\n" +
	"b=AS:256\r\n" +
	"a=rtpmap:96 MPEG4-GENERIC/16000/2\r\n" +
	"a=fmtp:96 streamtype=5;profile-level-id=1;mode=AAC-hbr;sizelength=13;indexlength=3;indexdeltalength=3;config=1408\r\n" +
	"a=control:track2\r\n" +
	"m=audio 0 RTP/AVP 0\r\n" +
	"c=IN IP4 0.0.0.0\r\n" +
	"b=AS:50\r\n" +
	"a=recvonly\r\n" +
	"a=control:rtsp://109.195.127.207:554/mpeg4cif/trackID=2\r\n" +
	"a=rtpmap:0 PCMU/8000\r\n" +
	"a=Media_header:MEDIAINFO=494D4B48010100000400010010710110401F000000FA000000000000000000000000000000000000;\r\n" +
	"a=appversion:1.0\r\n"

const H264UnamarshalSDP = "v=0\r\n" +
	"o=RTSP 1739297693 341 IN IP4 0.0.0.0\r\n" +
	"s=RTSP server\r\n" +
	"c=IN IP4 0.0.0.0\r\n" +
	"t=0 0\r\n" +
	"a=charset:Shift_JIS\r\n" +
	"a=range:npt=0-\r\n" +
	"a=control:*\r\n" +
	"a=etag:1234567890\r\n" +
	"m=video 0 RTP/AVP 98\r\n" +
	"b=AS:0\r\n" +
	"a=rtpmap:98 H264/90000\r\n" +
	"a=control:trackID=2\r\n" +
	"a=x-onvif-track:trackID=2\r\n" +
	"a=fmtp:98 packetization-mode=1; profile-level-id=4d6020; sprop-parameter-sets=J01gII1oBVBhv/AQAA/2yAAAAwAIAAADAFB4oRUA,KO4FSSAAAAAAAAAA\r\n" +
	"m=audio 0 RTP/AVP 0\r\n" +
	"a=control:trackID=8\r\n" +
	"a=x-onvif-track:trackID=8\r\n" +
	"a=rtpmap:0 pcmu/8000\r\n" +
	"m=application 0 RTP/AVP 107\r\n" +
	"a=control:trackID=13\r\n" +
	"a=rtpmap:107 vnd.onvif.metadata/90000\r\n"

const H265UnmarshalSDP = "v=0\r\n" +
	"o=RTSP 1751733122 286 IN IP4 0.0.0.0\r\n" +
	"s=RTSP server\r\n" +
	"c=IN IP4 0.0.0.0\r\n" +
	"t=0 0\r\n" +
	"a=charset:Shift_JIS\r\n" +
	"a=range:npt=0-\r\n" +
	"a=control:*\r\n" +
	"a=etag:1234567890\r\n" +
	"m=video 0 RTP/AVP 99\r\n" +
	"a=rtpmap:99 H265/90000\r\n" +
	"a=control:trackID=2\r\n" +
	"a=x-onvif-track:trackID=2\r\n" +
	"a=fmtp:99 sprop-vps=QAEMAf//AWAAAAMAgAAAAwAAAwCWrAk=; sprop-sps=QgEBAWAAAAMAgAAAAwAAAwCWoAFAIAeB/ja7tTd3JdYC3AQEBBAAAD6AAAJxByHe5R2I; sprop-pps=RAHBcrCcGw3iQA==\r\n" +
	"a=recvonly\r\n" +
	"m=audio 0 RTP/AVP 0\r\n" +
	"a=rtpmap:0 pcmu/8000\r\n" +
	"a=control:trackID=0\r\n" +
	"a=x-onvif-track:trackID=0\r\n" +
	"a=recvonly\r\n" +
	"m=application 0 RTP/AVP 107\r\n" +
	"a=control:trackID=1\r\n" +
	"a=x-onvif-track:trackID=1\r\n" +
	"a=rtpmap:107 vnd.onvif.metadata/90000\r\n" +
	"a=recvonly\r\n"

func tt(t av.AudioCodecData) {
}

func TestAudioCodecs(t *testing.T) {
	s := pcm.SpeexCodecData{}
	tt(s)
}

func TestSdpToCodecs(t *testing.T) {
	tests := []struct {
		name    string
		s       string
		wantRet []av.CodecData
		wantErr bool
	}{
		{
			name: "H264UnamarshalSDP",
			s:    H264UnamarshalSDP,
		},
		{
			name: "MPEG4UnmarshalSDP",
			s:    MPEG4UnmarshalSDP,
		},
		{
			name: "H265UnmarshalSDP",
			s:    H265UnmarshalSDP,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotRet, err := codec.SdpToCodecs(tt.s)
			if (err != nil) != tt.wantErr {
				t.Errorf("SdpToCodecs() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			for _, codec := range gotRet {
				switch v := codec.(type) {
				case h264parser.CodecData:
					fmt.Println(v.Width(), v.Height())
				case h265parser.CodecData:
					fmt.Println(v.Width(), v.Height())
				}
			}
		})
	}
}
