package oauth2x

import (
	"bytes"
	"io"
	"os"

	"github.com/natefinch/atomic"
)

// NewFileTokenCache returns a TokenCache that writes to a file.
func NewFileTokenCache(filename string) TokenCache {
	return &fileTokenCache{
		filename: filename,
	}
}

type fileTokenCache struct {
	filename string
}

func (fc *fileTokenCache) Load() ([]byte, error) {
	f, err := os.Open(fc.filename)
	if err != nil {
		return nil, err
	}

	return io.ReadAll(f)
}

func (fc *fileTokenCache) Store(token []byte) error {
	return atomic.WriteFile(fc.filename, bytes.NewReader(token))
}

func (fc *fileTokenCache) Name() string {
	return fc.filename
}
