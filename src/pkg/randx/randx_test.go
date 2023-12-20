package randx

import (
	"io"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRand(t *testing.T) {
	r := New(rand.NewSource(56746)) // fixed seed
	assert.Equal(t, 2171523247376636545, r.Int())
	assert.Equal(t, 2550976, r.Intn(3933409))
	assert.Equal(t, int32(785575095), r.Int31())
	assert.Equal(t, int32(2006), r.Int31n(3534))
	assert.Equal(t, int64(5483311004276762746), r.Int63())
	assert.Equal(t, int64(50783875), r.Int63n(239034904))
	assert.Equal(t, float32(0.070779815), r.Float32())
	assert.Equal(t, 0.7544574084626943, r.Float64())
	assert.Equal(t, "kevmfbrhpb", r.String(10, []rune("abcdefghijklmnopqrstuvwxyz")))
}

func TestSecureStringRand(t *testing.T) {
	rng := NewSecureStringRand(&ringBufferReader{
		b: []byte("abcdefghijklmnopqrstuvwxyz0123456789"),
	})

	s := rng.String(10, []rune("abcdefghijklmnopqrstuvwxyz"))
	assert.Equal(t, "bcdefghijk", s)
}

type ringBufferReader struct {
	b   []byte
	pos int
}

func (r *ringBufferReader) Read(p []byte) (n int, err error) {
	for i := range p {
		if r.pos > len(r.b) {
			r.pos = 0
		}

		p[i] = r.b[r.pos]
		r.pos++
	}
	return len(p), nil
}

var (
	_ io.Reader = &ringBufferReader{}
)
