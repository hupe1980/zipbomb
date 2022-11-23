package filename

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDefaultGenerator(t *testing.T) {
	t.Run("DefaultAlphabet", func(t *testing.T) {
		gen := NewDefaultGenerator(DefaultAlphabet, "")
		name := gen.Generate(0)
		assert.Equal(t, "0", name)
		name = gen.Generate(42)
		assert.Equal(t, "06", name)
		name = gen.Generate(100)
		assert.Equal(t, "1S", name)
	})

	t.Run("Custom alphabet", func(t *testing.T) {
		gen := NewDefaultGenerator([]byte("XYZ"), "")
		name := gen.Generate(0)
		assert.Equal(t, "X", name)
	})

	t.Run("Extension", func(t *testing.T) {
		gen := NewDefaultGenerator(DefaultAlphabet, "pdf")
		name := gen.Generate(0)
		assert.Equal(t, "0.pdf", name)
	})
}
