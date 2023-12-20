// Package randx contains randomization functions.
package randx

import (
	cryptrand "crypto/rand"
	"io"
	"math/big"
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
	StringRand

	Int() int
	Intn(n int) int
	Int31() int32
	Int31n(n int32) int32
	Int63() int64
	Int63n(n int64) int64
	Float32() float32
	Float64() float64
	Read(p []byte) (n int, err error)
}

// StringRand can generate random strings for an alphabet.
// Supports both math/rand and crypto/rand.
type StringRand interface {
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

// NewSecureStringRand returns a StringRand that uses
// a secure RNG.
func NewSecureStringRand(r io.Reader) StringRand {
	return secureStringRand{
		r: r,
	}
}

// SecureStringRand is a StringRand that uses a secure RNG.
var (
	SecureStringRand StringRand = NewSecureStringRand(cryptrand.Reader)
)

type secureStringRand struct {
	r io.Reader
}

func (rnd secureStringRand) String(n int, alphabet []rune) string {
	output := make([]rune, n)
	for i := 0; i < n; i++ {
		num, err := cryptrand.Int(rnd.r, big.NewInt(int64(len(alphabet))))
		if err != nil {
			panic(err)
		}
		output[i] = alphabet[num.Int64()]
	}

	return string(output)
}

// SecureString returns a string of length n, composed
// of random characters from the provided alphabet, generated
// from a secure RNG. If the input slice is nil, returns a random
// mixed case alphanumeric string.
func SecureString(n int, alphabet []rune) string {
	return SecureStringRand.String(n, alphabet)
}

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
