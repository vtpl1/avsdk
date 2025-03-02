package h264parser

import "errors"

var (
	ErrSPSNotFound         = errors.New("h264parser SPS not found")
	ErrPPSNotFound         = errors.New("h264parser PPS not found")
	ErrDecconfInvalid      = errors.New("h264parser AVCDecoderConfRecord invalid")
	ErrPacketTooShort      = errors.New("h264parser packet too short to parse slice header")
	ErrNalHasNoSliceHeader = errors.New("h264parser nal_unit_type has no slice header")
	ErrInvalidSliceType    = errors.New("h264parser slice_type invalid")
)
