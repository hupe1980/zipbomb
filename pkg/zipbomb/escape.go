package zipbomb

import "hash"

type escape struct {
	lfh  *fileHeader
	data []byte
	name string
}

func newEscape(name string, header *fileHeader, numEscaped uint16, crc32 hash.Hash32) *escape {
	var buf [5]byte
	b := writeBuf(buf[:])
	b.uint8(0x00)                 // BTYPE=00 => no compression
	b.uint16(numEscaped)          // LEN
	b.uint16(numEscaped ^ 0xffff) // NLEN => one's complement of LEN

	lfh := newFileHeader(
		uint64(uint32(len(buf))+uint32(numEscaped)+header.CompressedSize),
		uint64(uint32(numEscaped)+header.UncompressedSize),
		crc32.Sum32(),
		name,
		Deflate,
	)

	return &escape{
		lfh:  lfh,
		data: buf[:],
		name: name,
	}
}

func (e *escape) Name() string {
	return e.name
}

func (e *escape) LocalFileHeader() *fileHeader {
	return e.lfh
}

func (e *escape) Data() []byte {
	return e.data
}
