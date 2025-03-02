// Package h264parser holds Muxer and Demuxer for h264
package h264parser

import (
	"encoding/hex"
	"testing"
)

func TestSplitNALUs(t *testing.T) {
	annexbFrame, _ := hex.DecodeString("00000001223322330000000122332233223300000133000001000001")
	avccFrame, _ := hex.DecodeString(
		"00000008aabbccaabbccaabb00000001aa",
	)
	type args struct {
		b []byte
	}
	tests := []struct {
		name      string
		b         []byte
		wantNalus [][]byte
		wantTyp   NaluAvccOrAnnexb
	}{
		{
			name:    "annexbFrame",
			b:       annexbFrame,
			wantTyp: NaluAnnexb,
		},
		{
			name:    "avccFrame",
			b:       avccFrame,
			wantTyp: NaluAvcc,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotNalus, gotTyp := SplitNALUs(tt.b)
			t.Logf("SplitNALUs() gotNalus = %v", gotNalus)
			if gotTyp != tt.wantTyp {
				t.Errorf("SplitNALUs() gotTyp = %v, want %v", gotTyp, tt.wantTyp)
			}
		})
	}
}
