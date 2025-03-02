package bits

import "io"

type Reader struct {
	R    io.Reader
	n    int
	bits uint64
}

func (s *Reader) ReadBits64(n int) (uint64, error) {
	var bits uint64

	var err error

	if s.n < n {
		var b [8]byte

		var got int

		want := (n - s.n + 7) / 8
		if got, err = s.R.Read(b[:want]); err != nil {
			return bits, err
		}

		if got < want {
			err = io.EOF

			return bits, err
		}

		for i := range got {
			s.bits <<= 8
			s.bits |= uint64(b[i])
		}

		s.n += got * 8
	}

	bits = s.bits >> uint(s.n-n)
	s.bits ^= bits << uint(s.n-n)
	s.n -= n

	return bits, err
}

func (s *Reader) ReadBits(n int) (uint, error) {
	var bits64 uint64

	var bits uint

	var err error
	if bits64, err = s.ReadBits64(n); err != nil {
		return bits, err
	}

	bits = uint(bits64)

	return bits, err
}

func (s *Reader) Read(p []byte) (int, error) {
	n := 0

	var err error

	for n < len(p) {
		want := 8
		if len(p)-n < want {
			want = len(p) - n
		}

		var bits uint64

		if bits, err = s.ReadBits64(want * 8); err != nil {
			break
		}

		for i := range want {
			p[n+i] = byte(bits >> uint((want-i-1)*8))
		}

		n += want
	}

	return n, err
}

type Writer struct {
	W    io.Writer
	n    int
	bits uint64
}

func (s *Writer) WriteBits64(bits uint64, n int) error {
	var err error

	if s.n+n > 64 {
		move := uint(64 - s.n)
		mask := bits >> move
		s.bits = (s.bits << move) | mask
		s.n = 64

		if err = s.FlushBits(); err != nil {
			return err
		}

		n -= int(move)
		bits ^= (mask << move)
	}

	s.bits = (s.bits << uint(n)) | bits
	s.n += n

	return err
}

func (s *Writer) WriteBits(bits uint, n int) error {
	return s.WriteBits64(uint64(bits), n)
}

func (s *Writer) Write(p []byte) (int, error) {
	n := 0

	var err error

	for n < len(p) {
		if err = s.WriteBits64(uint64(p[n]), 8); err != nil {
			return n, err
		}

		n++
	}

	return n, err
}

func (s *Writer) FlushBits() error {
	var err error

	if s.n > 0 {
		var b [8]byte

		bits := s.bits
		if s.n%8 != 0 {
			bits <<= uint(8 - (s.n % 8))
		}

		want := (s.n + 7) / 8
		for i := range want {
			b[i] = byte(bits >> uint((want-i-1)*8))
		}

		if _, err = s.W.Write(b[:want]); err != nil {
			return err
		}

		s.n = 0
	}

	return err
}
