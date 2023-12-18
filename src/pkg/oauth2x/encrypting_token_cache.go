package oauth2x

import (
	"crypto/rand"
	"fmt"

	"golang.org/x/crypto/nacl/secretbox"
)

const (
	// KeySize is the size of the encryption key.
	KeySize   = 32
	nonceSize = 24
)

// EncryptTokenCache wraps a token cache with an encryption layer.
func EncryptTokenCache(cache TokenCache, key [KeySize]byte) TokenCache {
	return &encryptingTokenCache{
		cache: cache,
		key:   key,
	}
}

type encryptingTokenCache struct {
	cache TokenCache
	key   [KeySize]byte
}

func (ec *encryptingTokenCache) Load() ([]byte, error) {
	encrypted, err := ec.cache.Load()
	if err != nil {
		return nil, err
	}

	// Nonce is prepended to the message
	var nonce [nonceSize]byte
	copy(nonce[:], encrypted[:nonceSize])
	box := encrypted[nonceSize:]
	plainText, ok := secretbox.Open(nil, box, &nonce, &ec.key)
	if !ok {
		return nil, fmt.Errorf("unable to unbox message")
	}

	return plainText, nil
}

func (ec *encryptingTokenCache) Store(contents []byte) error {
	// Generate a random nonce
	var nonce [nonceSize]byte
	if _, err := rand.Read(nonce[:]); err != nil {
		return fmt.Errorf("unable to generate nonce: %w", err)
	}

	// Prepend the nonce
	stored := make([]byte, len(contents)+secretbox.Overhead+len(nonce))
	copy(stored, nonce[:])

	// Seal into the remainder of the bytes
	secretbox.Seal(stored[:nonceSize], contents, &nonce, &ec.key)
	return ec.cache.Store(stored)
}

func (ec *encryptingTokenCache) Name() string {
	return ec.cache.Name()
}
