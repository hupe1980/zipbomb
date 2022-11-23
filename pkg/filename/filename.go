package filename

import "fmt"

var DefaultAlphabet = []byte("0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ")

type Generator interface {
	Generate(i int) string
}

type DefaultGenerator struct {
	alphabet  []byte
	extension string
}

func NewDefaultGenerator(alphabet []byte, extension string) Generator {
	g := &DefaultGenerator{
		alphabet:  alphabet,
		extension: extension,
	}

	if len(g.alphabet) == 0 {
		g.alphabet = DefaultAlphabet
	}

	return g
}

func (g *DefaultGenerator) Generate(i int) string {
	letters := []byte{}

	for {
		letters = append([]byte{g.alphabet[i%len(g.alphabet)]}, letters...)

		i = i/len(g.alphabet) - 1
		if i < 0 {
			break
		}
	}

	if g.extension != "" {
		return fmt.Sprintf("%s.%s", letters, g.extension)
	}

	return string(letters)
}
