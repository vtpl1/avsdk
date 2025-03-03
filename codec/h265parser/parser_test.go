package h265parser_test

import (
	"encoding/hex"
	"testing"

	"github.com/vtpl1/avsdk/codec/h265parser"
)

func TestSplitNALUs(t *testing.T) {
	annexbFrame, _ := hex.DecodeString("00000001223322330000000122332233223300000133000001000001")
	avccFrame, _ := hex.DecodeString(
		"00000008aabbccaabbccaabb00000001aa",
	)

	tests := []struct {
		name      string
		b         []byte
		wantNalus [][]byte
		wantTyp   h265parser.NALUAvccOrAnnexb
	}{
		{
			name:    "annexbFrame",
			b:       annexbFrame,
			wantTyp: h265parser.NALUAnnexb,
		},
		{
			name:    "avccFrame",
			b:       avccFrame,
			wantTyp: h265parser.NALUAvcc,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotNalus, gotTyp := h265parser.SplitNALUs(tt.b)
			t.Logf("SplitNALUs() gotNalus = %v", gotNalus)
			if gotTyp != tt.wantTyp {
				t.Errorf("SplitNALUs() gotTyp = %v, want %v", gotTyp, tt.wantTyp)
			}
		})
	}
}
