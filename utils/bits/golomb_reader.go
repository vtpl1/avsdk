// Package bits holds bit related functionalities
package bits

import (
	"io"
)

type GolombBitReader struct {
	R    io.Reader
	buf  [1]byte
	left byte
}

func (s *GolombBitReader) ReadBit() (uint, error) {
	if s.left == 0 {
		if _, err := s.R.Read(s.buf[:]); err != nil {
			return 0, err
		}

		s.left = 8
	}

	s.left--
	res := uint(s.buf[0]>>s.left) & 1

	return res, nil
}

func (s *GolombBitReader) ReadBits(n int) (uint, error) {
	var res uint

	for i := range n {
		var bit uint

		var err error
		if bit, err = s.ReadBit(); err != nil {
			return 0, err
		}

		res |= bit << uint(n-i-1)
	}

	return res, nil
}

func (s *GolombBitReader) ReadBits32(n uint) (uint32, error) {
	var t uint

	var r uint32

	for range n {
		var err error

		t, err = s.ReadBit()
		if err != nil {
			return r, err
		}

		r = (r << 1) | uint32(t)
	}

	return r, nil
}

func (s *GolombBitReader) ReadBits64(n uint) (uint64, error) {
	var t uint

	var r uint64

	for range n {
		var err error

		t, err = s.ReadBit()
		if err != nil {
			return r, err
		}

		r = (r << 1) | uint64(t)
	}

	return r, nil
}

func (s *GolombBitReader) ReadExponentialGolombCode() (uint, error) {
	i := 0

	var res uint

	var err error

	for {
		var bit uint

		if bit, err = s.ReadBit(); err != nil {
			return res, err
		}

		if !(bit == 0 && i < 32) {
			break
		}

		i++
	}

	if res, err = s.ReadBits(i); err != nil {
		return res, err
	}

	res += (1 << uint(i)) - 1

	return res, err
}

func (s *GolombBitReader) ReadSE() (uint, error) {
	var res uint

	var err error
	if res, err = s.ReadExponentialGolombCode(); err != nil {
		return res, err
	}

	if res&0x01 != 0 {
		res = (res + 1) / 2
	} else {
		res = -res / 2
	}

	return res, err
}
