package pio

func U8(b []byte) uint8 {
	return b[0]
}

func U16BE(b []byte) uint16 {
	i := uint16(b[0])
	i <<= 8
	i |= uint16(b[1])

	return i
}

func I16BE(b []byte) int16 {
	i := int16(b[0])
	i <<= 8
	i |= int16(b[1])

	return i
}

func I24BE(b []byte) int32 {
	i := int32(int8(b[0]))
	i <<= 8
	i |= int32(b[1])
	i <<= 8
	i |= int32(b[2])

	return i
}

func U24BE(b []byte) uint32 {
	i := uint32(b[0])
	i <<= 8
	i |= uint32(b[1])
	i <<= 8
	i |= uint32(b[2])

	return i
}

func I32BE(b []byte) int32 {
	i := int32(int8(b[0]))
	i <<= 8
	i |= int32(b[1])
	i <<= 8
	i |= int32(b[2])
	i <<= 8
	i |= int32(b[3])

	return i
}

func U32LE(b []byte) uint32 {
	i := uint32(b[3])
	i <<= 8
	i |= uint32(b[2])
	i <<= 8
	i |= uint32(b[1])
	i <<= 8
	i |= uint32(b[0])

	return i
}

func U32BE(b []byte) uint32 {
	i := uint32(b[0])
	i <<= 8
	i |= uint32(b[1])
	i <<= 8
	i |= uint32(b[2])
	i <<= 8
	i |= uint32(b[3])

	return i
}

func U40BE(b []byte) uint64 {
	i := uint64(b[0])
	i <<= 8
	i |= uint64(b[1])
	i <<= 8
	i |= uint64(b[2])
	i <<= 8
	i |= uint64(b[3])
	i <<= 8
	i |= uint64(b[4])

	return i
}

func U64BE(b []byte) uint64 {
	i := uint64(b[0])
	i <<= 8
	i |= uint64(b[1])
	i <<= 8
	i |= uint64(b[2])
	i <<= 8
	i |= uint64(b[3])
	i <<= 8
	i |= uint64(b[4])
	i <<= 8
	i |= uint64(b[5])
	i <<= 8
	i |= uint64(b[6])
	i <<= 8
	i |= uint64(b[7])

	return i
}

func I64BE(b []byte) int64 {
	i := int64(int8(b[0]))
	i <<= 8
	i |= int64(b[1])
	i <<= 8
	i |= int64(b[2])
	i <<= 8
	i |= int64(b[3])
	i <<= 8
	i |= int64(b[4])
	i <<= 8
	i |= int64(b[5])
	i <<= 8
	i |= int64(b[6])
	i <<= 8
	i |= int64(b[7])

	return i
}
