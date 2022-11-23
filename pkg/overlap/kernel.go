package overlap

import (
	"bytes"
	"compress/flate"
	"hash/crc32"
)

type kernel struct {
	rawBytes        []byte
	compressedBytes []byte
	crc32           uint32
	lfh             *fileHeader
	name            string
}

func newKernel(name string, data []byte) (*kernel, error) {
	compressedBytes, err := CompressKernel(data)
	if err != nil {
		return nil, err
	}

	crc32 := crc32.ChecksumIEEE(data)

	k := &kernel{
		rawBytes:        data,
		compressedBytes: compressedBytes,
		crc32:           crc32,
		name:            name,
	}

	k.lfh = newFileHeader(k.CompressedSize(), k.UncompressedSize(), k.CRC32(), name)

	return k, nil
}

func (k *kernel) LocalFileHeader() *fileHeader {
	return k.lfh
}

func (k *kernel) LocalFileHeaderBytes() ([]byte, error) {
	return k.lfh.MarshalBinary()
}

func (k *kernel) CRC32() uint32 {
	return k.crc32
}

func (k *kernel) UncompressedSize() uint64 {
	return uint64(len(k.rawBytes))
}

func (k *kernel) CompressedSize() uint64 {
	return uint64(len(k.compressedBytes))
}

func (k *kernel) Bytes() []byte {
	return k.rawBytes
}

func (k *kernel) Name() string {
	return k.name
}

func CompressKernel(data []byte) ([]byte, error) {
	buffer := new(bytes.Buffer)

	fw, err := flate.NewWriter(buffer, 5)
	if err != nil {
		return nil, err
	}

	if _, err := fw.Write(data); err != nil {
		return nil, err
	}

	if err := fw.Close(); err != nil {
		return nil, err
	}

	return buffer.Bytes(), nil
}
