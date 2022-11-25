package zipbomb

import (
	"bytes"
	"io"
	"time"
)

type fileHeader struct {
	// Name is the name of the file.
	Name string

	Comment string

	CreatorVersion uint16
	ReaderVersion  uint16
	Flags          uint16

	// Method is the compression method.
	Method uint16

	ModifiedTime uint16
	ModifiedDate uint16

	CRC32              uint32
	CompressedSize     uint32
	UncompressedSize   uint32
	CompressedSize64   uint64
	UncompressedSize64 uint64

	Extra         []byte
	ExternalAttrs uint32
}

func newFileHeader(compressedSize, uncompressedSize uint64, crc32 uint32, name string, method uint16) *fileHeader {
	fdate, ftime := timeToMsDosTime(time.Now())

	lfh := &fileHeader{
		CompressedSize64:   compressedSize,
		UncompressedSize64: uncompressedSize,
		CRC32:              crc32,
		Name:               name,
		Method:             Deflate,
		ModifiedTime:       ftime,
		ModifiedDate:       fdate,
	}

	var zipVersion uint16

	switch method {
	case Deflate:
		if lfh.IsZip64() {
			zipVersion = zipVersion45
		} else {
			zipVersion = zipVersion20
		}
	case BZip2:
		zipVersion = zipVersion46
	}

	if lfh.IsZip64() {
		lfh.CompressedSize = uint32max
		lfh.UncompressedSize = uint32max
		lfh.ReaderVersion = zipVersion

		// append a zip64 extra block to Extra
		var buf [20]byte // 2x uint16 + 2x uint64
		eb := writeBuf(buf[:])
		eb.uint16(zip64ExtraID)
		eb.uint16(16) // size = 2x uint64
		eb.uint64(lfh.UncompressedSize64)
		eb.uint64(lfh.CompressedSize64)
		lfh.Extra = append(lfh.Extra, buf[:]...)
	} else {
		lfh.CompressedSize = uint32(lfh.CompressedSize64)
		lfh.UncompressedSize = uint32(lfh.UncompressedSize64)
		lfh.ReaderVersion = zipVersion
	}

	lfh.CreatorVersion = (0 << 8) | lfh.ReaderVersion

	return lfh
}

// IsZip64 reports whether the file size exceeds the 32 bit limit
func (h *fileHeader) IsZip64() bool {
	return h.CompressedSize64 >= uint32max || h.UncompressedSize64 >= uint32max
}

func (h *fileHeader) MarshalBinary() ([]byte, error) {
	buffer := new(bytes.Buffer)

	if len(h.Name) > uint16max {
		return nil, errLongName
	}

	if len(h.Extra) > uint16max {
		return nil, errLongExtra
	}

	var buf [fileHeaderLen]byte
	b := writeBuf(buf[:])
	b.uint32(uint32(fileHeaderSignature))
	b.uint16(h.ReaderVersion)
	b.uint16(h.Flags)
	b.uint16(h.Method)
	b.uint16(h.ModifiedTime)
	b.uint16(h.ModifiedDate)
	b.uint32(h.CRC32)

	b.uint32(uint32(min64(h.CompressedSize64, uint32max)))
	b.uint32(uint32(min64(h.UncompressedSize64, uint32max)))

	b.uint16(uint16(len(h.Name)))
	b.uint16(uint16(len(h.Extra)))

	if _, err := buffer.Write(buf[:]); err != nil {
		return nil, err
	}

	if _, err := io.WriteString(buffer, h.Name); err != nil {
		return nil, err
	}

	if _, err := buffer.Write(h.Extra); err != nil {
		return nil, err
	}

	if _, err := io.WriteString(buffer, h.Comment); err != nil {
		return nil, err
	}

	return buffer.Bytes(), nil
}

// timeToMsDosTime converts a time.Time to an MS-DOS date and time.
// See https://msdn.microsoft.com/en-us/library/ms724274(v=VS.85).aspx
func timeToMsDosTime(t time.Time) (fDate uint16, fTime uint16) {
	fDate = uint16(t.Day() + int(t.Month())<<5 + (t.Year()-1980)<<9)
	fTime = uint16(t.Second()/2 + t.Minute()<<5 + t.Hour()<<11)

	return
}

func min64(x, y uint64) uint64 {
	if x < y {
		return x
	}

	return y
}
