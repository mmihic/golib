// Package randx contains randomization functions.
package randx

import (
	"math/rand"
	"time"
)

var (
	alphanum = []rune{
		'a', 'b', 'c', 'd', 'e', 'f', 'g',
		'h', 'i', 'j', 'k', 'l', 'm', 'n',
		'o', 'p', 'q', 'r', 's', 't', 'u',
		'v', 'x', 'y', 'z', '0', '1', '2',
		'3', '4', '5', '6', '7', '8', '9',
		'A', 'B', 'C', 'D', 'E', 'F', 'G',
		'H', 'I', 'J', 'K', 'L', 'M', 'N',
		'O', 'P', 'Q', 'R', 'S', 'T', 'U',
		'V', 'X', 'Y', 'Z'}
)

// Rand can generate random values of various types.
type Rand interface {
	Int() int
	Intn(n int) int
	Int31() int32
	Int31n(n int32) int32
	Int63() int64
	Int63n(n int64) int64
	Float32() float32
	Float64() float64
	Read(p []byte) (n int, err error)
	String(n int, alphabet []rune) string
}

// New creates a new RNG around a given source.
func New(src rand.Source) Rand {
	return &rng{
		Rand: rand.New(src),
	}
}

type rng struct {
	*rand.Rand // delegates most functions to the underlying rng
}

func (r *rng) String(n int, alphabet []rune) string {
	if alphabet == nil {
		alphabet = alphanum
	}

	out := make([]rune, n)
	for i := range out {
		out[i] = alphabet[r.Intn(len(alphabet))]
	}

	return string(out)
}

var _ Rand = &rng{}

// String returns a string of length n, composed of random
// characters from the provided alphabet. If the input
// slice is nil, returns a random mixed case alphanumeric
// string.
func String(n int, alphabet []rune) string {
	return defaultRNG.String(n, alphabet)
}

var (
	defaultRNG = New(rand.NewSource(time.Now().UnixMilli()))
)
