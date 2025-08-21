// Package h264parser holds Muxer and Demuxer for h264
package h264parser

import (
	"encoding/hex"
	"reflect"
	"testing"
)

func TestParseSPS(t *testing.T) {
	type args struct {
		data []byte
	}
	sps1nalu, _ := hex.DecodeString("67640020accac05005bb0169e0000003002000000c9c4c000432380008647c12401cb1c31380")
	sps2nalu, _ := hex.DecodeString("6764000dacd941419f9e10000003001000000303c0f1429960")
	sps3nalu, _ := hex.DecodeString("27640020ac2ec05005bb011000000300100000078e840016e300005b8d8bdef83b438627")
	sps4nalu, _ := hex.DecodeString("674d00329a64015005fff8037010101400000fa000013883a1800fee0003fb52ef2e343001fdc0007f6a5de5c280")
	tests := []struct {
		name    string
		args    args
		want    SPSInfo
		wantErr bool
	}{
		{
			name: "sps1nalu",
			args: args{
				data: sps1nalu,
			},
			want: SPSInfo{
				ID:                0,
				ProfileIdc:        100,
				LevelIdc:          32,
				ConstraintSetFlag: 0,
				MbWidth:           80,
				MbHeight:          45,
				CropLeft:          0,
				CropRight:         0,
				CropTop:           0,
				CropBottom:        0,
				Width:             1280,
				Height:            720,
				FPS:               50,
			},
			wantErr: false,
		},
		{
			name: "sps2nalu",
			args: args{
				data: sps2nalu,
			},
			want: SPSInfo{
				ID:                0,
				ProfileIdc:        100,
				LevelIdc:          13,
				ConstraintSetFlag: 0,
				MbWidth:           20,
				MbHeight:          12,
				CropLeft:          0,
				CropRight:         0,
				CropTop:           0,
				CropBottom:        6,
				Width:             320,
				Height:            180,
				FPS:               30,
			},
			wantErr: false,
		},
		{
			name: "sps3nalu",
			args: args{
				data: sps3nalu,
			},
			want: SPSInfo{
				ID:                0,
				ProfileIdc:        100,
				LevelIdc:          32,
				ConstraintSetFlag: 0,
				MbWidth:           80,
				MbHeight:          45,
				CropLeft:          0,
				CropRight:         0,
				CropTop:           0,
				CropBottom:        0,
				Width:             1280,
				Height:            720,
				FPS:               60,
			},
			wantErr: false,
		},
		{
			name: "getsaridc",
			args: args{
				data: sps4nalu,
			},
			want: SPSInfo{
				ID:                0,
				ProfileIdc:        77,
				LevelIdc:          50,
				ConstraintSetFlag: 0,
				MbWidth:           168,
				MbHeight:          95,
				CropLeft:          0,
				CropRight:         0,
				CropTop:           0,
				CropBottom:        0,
				Width:             2688,
				Height:            1520,
				FPS:               10,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseSPS(tt.args.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseSPS() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseSPS() = %v, want %v", got, tt.want)
			}
		})
	}
}
