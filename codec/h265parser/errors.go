package h265parser

import "errors"

var (
	ErrSPSNotFound          = errors.New("h265parser SPS not found")
	ErrPPSNotFound          = errors.New("h265parser PPS not found")
	ErrVPSNotFound          = errors.New("h265parser VPS not found")
	ErrSPSParseFailed       = errors.New("h265parser parse SPS failed")
	ErrDecconfInvalid       = errors.New("h265parser AVCDecoderConfRecord invalid")
	ErrPacketTooShort       = errors.New("h265parser packet too short to parse slice header")
	ErrNalHasNoSliceHeader  = errors.New("h265parser nal_unit_type has no slice header")
	ErrInvalidSliceType     = errors.New("h265parser slice_type invalid")
	ErrH265IncorectUnitSize = errors.New("incorrect unit cize")
	ErrH265IncorectUnitType = errors.New("incorrect unit type")
)
