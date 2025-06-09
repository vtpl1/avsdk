package parser_test

import (
	"encoding/hex"
	"testing"

	"github.com/vtpl1/avsdk/codec/parser"
)

func TestSplitNALUs(t *testing.T) {
	annexbFrame, _ := hex.DecodeString("00000001223322330000000122332233223300000133000001000001")
	annexbFrame2, _ := hex.DecodeString("000000016742c028d900780227e584000003000400000300503c60c920")
	avccFrame, _ := hex.DecodeString("00000008aabbccaabbccaabb00000001aa")

	tests := []struct {
		name      string
		b         []byte
		wantNalus [][]byte
		wantTyp   parser.NALUAvccOrAnnexb
		wantLen   int
	}{
		{
			name:    "annexbFrame",
			b:       annexbFrame,
			wantTyp: parser.NALUAnnexb,
			wantLen: 3,
		},
		{
			name:    "annexbFrame2",
			b:       annexbFrame2,
			wantTyp: parser.NALUAnnexb,
			wantLen: 1,
		},
		{
			name:    "avccFrame",
			b:       avccFrame,
			wantTyp: parser.NALUAvcc,
			wantLen: 2,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotNalus, gotTyp := parser.SplitNALUs(tt.b)
			if gotTyp != tt.wantTyp {
				t.Errorf("SplitNALUs() gotTyp = %v, want %v", gotTyp, tt.wantTyp)
			}
			if len(gotNalus) != tt.wantLen {
				t.Errorf("SplitNALUs() len(gotNalus) = %v, want %v", len(gotNalus), tt.wantLen)
			}
		})
	}
}

func TestIsAnnexBOrAVCC(t *testing.T) {
	if typ := parser.IsAnnexBOrAVCC([]byte{0x00, 0x00, 0x00, 0x01, 0x65}); typ != parser.NALUAnnexb {
		t.Errorf("Expected AnnexB, got %v", typ)
	}
	if typ := parser.IsAnnexBOrAVCC([]byte{0x00, 0x00, 0x01, 0x65}); typ != parser.NALUAnnexb {
		t.Errorf("Expected AnnexB, got %v", typ)
	}
	if typ := parser.IsAnnexBOrAVCC([]byte{0x00, 0x00, 0x00, 0x05, 0x65, 0x88, 0x99, 0xaa, 0xbb}); typ != parser.NALUAvcc {
		t.Errorf("Expected AVCC, got %v", typ)
	}
}

func TestAnnexBToAVCC(t *testing.T) {
	annexb := append([]byte{0x00, 0x00, 0x00, 0x01}, []byte{0x65, 0x88, 0x99}...)
	avcc, err := parser.AnnexBToAVCC(annexb)
	if err != nil {
		t.Fatalf("AnnexBToAVCC() returned error: %v", err)
	}
	expected := []byte{0x00, 0x00, 0x00, 0x03, 0x65, 0x88, 0x99}
	if !equalSlices(avcc, expected) {
		t.Errorf("AnnexBToAVCC() = %v, want %v", avcc, expected)
	}
}

func TestAVCCToAnnexB(t *testing.T) {
	avcc := []byte{0x00, 0x00, 0x00, 0x03, 0x65, 0x88, 0x99}
	annexb, err := parser.AVCCToAnnexB(avcc)
	if err != nil {
		t.Fatalf("AVCCToAnnexB() returned error: %v", err)
	}
	expected := append([]byte{0x00, 0x00, 0x00, 0x01}, avcc[4:]...)
	if !equalSlices(annexb, expected) {
		t.Errorf("AVCCToAnnexB() = %v, want %v", annexb, expected)
	}
}

func TestSplitNALUs_H264(t *testing.T) {
	raw := append([]byte{0x00, 0x00, 0x01, 0x65, 0x44}, []byte{0x00, 0x00, 0x01, 0x61, 0x33}...)
	nalus, format := parser.SplitNALUs(raw)

	if len(nalus) != 2 {
		t.Errorf("Expected 2 NALUs, got %d", len(nalus))
	}
	if format != parser.NALUAnnexb {
		t.Errorf("Expected format Annexb, got %v", format)
	}
}

func TestSplitNALUs_H265(t *testing.T) {
	raw := append([]byte{0x00, 0x00, 0x01, 0x40, 0x01}, []byte{0x00, 0x00, 0x01, 0x26, 0x99}...)
	nalus, format := parser.SplitNALUs(raw)

	if len(nalus) != 2 {
		t.Errorf("Expected 2 NALUs, got %d", len(nalus))
	}
	if format != parser.NALUAnnexb {
		t.Errorf("Expected format Annexb, got %v", format)
	}
}

func TestSplitNALUs_1(t *testing.T) {
	data := []byte{
		0x00, 0x00, 0x00, 0x01, 0x67, 0x64, 0x00, 0x1f, // SPS
		0x00, 0x00, 0x01, 0x68, 0xee, 0x3c, 0x80, // PPS
		0x00, 0x00, 0x01, 0x65, 0x88, 0x84, 0x00, // IDR frame
	}

	nalus, format := parser.SplitNALUs(data)

	if len(nalus) != 3 {
		t.Errorf("Expected 3 NALUs, got %d", len(nalus))
	}
	if format != parser.NALUAnnexb {
		t.Errorf("Expected format Annexb, got %v", format)
	}
	if len(nalus) >= 3 {
		if nalus[0][0] != 0x67 {
			t.Errorf("Expected first NALU to start with 0x67, got 0x%x", nalus[0][0])
		}
		if nalus[1][0] != 0x68 {
			t.Errorf("Expected second NALU to start with 0x68, got 0x%x", nalus[1][0])
		}
		if nalus[2][0] != 0x65 {
			t.Errorf("Expected third NALU to start with 0x65, got 0x%x", nalus[2][0])
		}
	}
}

func equalSlices(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
