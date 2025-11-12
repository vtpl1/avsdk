package pcm_test

import (
	"encoding/hex"
	"fmt"
	"reflect"
	"testing"

	"github.com/vtpl1/avsdk/av"
	"github.com/vtpl1/avsdk/codec/pcm"
)

func TestTranscode(t *testing.T) {
	tests := []struct {
		name   string
		src    av.AudioCodecData
		dst    av.AudioCodecData
		source string
		expect string
	}{
		{
			name:   "s16be->s16be",
			src:    pcm.NewPCMCodecData(),
			dst:    pcm.NewPCMCodecData(),
			source: "FCCA00130343062808130B510D9E0F7610DA111113EA15BD16F2168215D41561",
			expect: "FCCA00130343062808130B510D9E0F7610DA111113EA15BD16F2168215D41561",
		},
		{
			name:   "s16be->s16le",
			src:    pcm.PCMCodecData{Typ: av.PCM, SmplRate: 8000, ChLayout: av.ChMono},
			dst:    pcm.PCMLCodecData{Typ: av.PCML, SmplRate: 8000, ChLayout: av.ChMono},
			source: "FCCA00130343062808130B510D9E0F7610DA111113EA15BD16F2168215D41561",
			expect: "CAFC1300430328061308510B9E0D760FDA101111EA13BD15F2168216D4156115",
		},
		{
			name:   "s16be->mulaw",
			src:    pcm.NewPCMCodecData(),
			dst:    pcm.NewPCMMulawCodecData(),
			source: "FCCA00130343062808130B510D9E0F7610DA111113EA15BD16F2168215D41561",
			expect: "52FDD1C5BEB8B3B0AEAEABA9A8A8A9AA",
		},
		{
			name:   "s16be->alaw",
			src:    pcm.NewPCMCodecData(),
			dst:    pcm.NewPCMAlawCodecData(),
			source: "FCCA00130343062808130B510D9E0F7610DA111113EA15BD16F2168215D41561",
			expect: "7CD4FFED95939E9B8584868083838080",
		},
		{
			name:   "2ch->1ch",
			src:    pcm.PCMCodecData{Typ: av.PCM, SmplRate: 8000, ChLayout: av.ChStereo},
			dst:    pcm.PCMCodecData{Typ: av.PCM, SmplRate: 8000, ChLayout: av.ChMono},
			source: "FCCAFCCA001300130343034306280628081308130B510B510D9E0D9E0F760F76",
			expect: "FCCA00130343062808130B510D9E0F76",
		},
		{
			name:   "1ch->2ch",
			src:    pcm.PCMCodecData{Typ: av.PCM, SmplRate: 8000, ChLayout: av.ChMono},
			dst:    pcm.PCMCodecData{Typ: av.PCM, SmplRate: 8000, ChLayout: av.ChStereo},
			source: "FCCA00130343062808130B510D9E0F76",
			expect: "FCCAFCCA001300130343034306280628081308130B510B510D9E0D9E0F760F76",
		},
		{
			name:   "16khz->8khz",
			src:    pcm.PCMCodecData{Typ: av.PCM, SmplRate: 16000, ChLayout: av.ChMono},
			dst:    pcm.PCMCodecData{Typ: av.PCM, SmplRate: 8000, ChLayout: av.ChMono},
			source: "FCCAFCCA001300130343034306280628081308130B510B510D9E0D9E0F760F76",
			expect: "FCCA00130343062808130B510D9E0F76",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			f := pcm.Transcode(test.dst, test.src)
			b, _ := hex.DecodeString(test.source)
			b = f(b)
			s := fmt.Sprintf("%X", b)
			if !reflect.DeepEqual(s, test.expect) {
				t.Errorf("Transcode() = %v, want %v", s, test.expect)
			}
			// require.Equal(t, test.expect, s)
		})
	}
}
