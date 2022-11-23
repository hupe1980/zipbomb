package overlap

import (
	"archive/zip"
	"bytes"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBomb(t *testing.T) {
	buffer := new(bytes.Buffer)

	zbomb := New(buffer)

	err := zbomb.Generate([]byte{'A'}, 3)
	assert.NoError(t, err)

	r, err := zip.NewReader(bytes.NewReader(buffer.Bytes()), int64(buffer.Len()))
	assert.NoError(t, err)

	for _, file := range r.File {
		fr, err := file.Open()
		assert.NoError(t, err)

		// nolint gosec testcase
		_, err = io.Copy(io.Discard, fr)
		assert.NoError(t, err)

		fr.Close()
	}
}
