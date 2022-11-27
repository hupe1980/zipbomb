package zipbomb

import (
	"bytes"
	"io"
	"io/fs"
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

	extraLengthExcess   uint16
	extraFieldEscapeTag uint16
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

func (h *fileHeader) SetExtraLengthExcess(v uint16) {
	h.extraLengthExcess = v
}

func (h *fileHeader) ExtraLengthExcess() uint16 {
	return h.extraLengthExcess
}

func (h *fileHeader) SetFieldEscapeTag(v uint16) {
	h.extraFieldEscapeTag = v
}

func (h *fileHeader) ExtraFieldEscapeTag() uint16 {
	return h.extraFieldEscapeTag
}

// func (h *fileHeader) UncompressedSize() uint64 {
// 	return h.UncompressedSize64
// }

// func (h *fileHeader) CompressedSize() uint64 {
// 	return h.CompressedSize64
// }

// IsZip64 reports whether the file size exceeds the 32 bit limit
func (h *fileHeader) IsZip64() bool {
	return h.CompressedSize64 >= uint32max || h.UncompressedSize64 >= uint32max
}

// SetMode changes the permission and mode bits for the FileHeader.
func (h *fileHeader) SetMode(mode fs.FileMode) {
	h.CreatorVersion = h.CreatorVersion&0xff | creatorUnix<<8
	h.ExternalAttrs = fileModeToUnixMode(mode) << 16

	// set MSDOS attributes too, as the original zip does.
	if mode&fs.ModeDir != 0 {
		h.ExternalAttrs |= msdosDir
	}

	if mode&0200 == 0 {
		h.ExternalAttrs |= msdosReadOnly
	}
}

func (h *fileHeader) MarshalBinary() ([]byte, error) {
	buffer := new(bytes.Buffer)

	if len(h.Name) > uint16max {
		return nil, errLongName
	}

	if len(h.Extra) > uint16max {
		return nil, errLongExtra
	}

	var extra []byte

	if h.extraFieldEscapeTag != 0 {
		var buf [4]byte
		eb := writeBuf(buf[:])
		eb.uint16(h.extraFieldEscapeTag)
		eb.uint16(h.extraLengthExcess)
		extra = append(h.Extra, buf[:]...)
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
	b.uint16(uint16(len(extra)) + h.extraLengthExcess)

	if _, err := buffer.Write(buf[:]); err != nil {
		return nil, err
	}

	if _, err := io.WriteString(buffer, h.Name); err != nil {
		return nil, err
	}

	if _, err := buffer.Write(extra); err != nil {
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

func fileModeToUnixMode(mode fs.FileMode) uint32 {
	var m uint32

	// nolint
	switch mode & fs.ModeType {
	default:
		m = IFREG
	case fs.ModeDir:
		m = IFDIR
	case fs.ModeSymlink:
		m = IFLNK
	case fs.ModeNamedPipe:
		m = IFIFO
	case fs.ModeSocket:
		m = IFSOCK
	case fs.ModeDevice:
		m = IFBLK
	case fs.ModeDevice | fs.ModeCharDevice:
		m = IFCHR
	}

	if mode&fs.ModeSetuid != 0 {
		m |= ISUID
	}

	if mode&fs.ModeSetgid != 0 {
		m |= ISGID
	}

	if mode&fs.ModeSticky != 0 {
		m |= ISVTX
	}

	return m | uint32(mode&0777)
}

func min64(x, y uint64) uint64 {
	if x < y {
		return x
	}

	return y
}
