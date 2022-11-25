package bzip2

import (
	"io"

	"github.com/dsnet/compress/bzip2"
)

func NewWriter(w io.Writer, level int) (*bzip2.Writer, error) {
	return bzip2.NewWriter(w, &bzip2.WriterConfig{
		Level: level,
	})
}
