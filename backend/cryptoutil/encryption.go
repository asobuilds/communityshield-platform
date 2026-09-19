package cryptoutil

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"io"
	"os"
	"strings"
)

// EncryptedPrefix marks a value as AES-256-GCM ciphertext.
// Values without this prefix are treated as legacy plaintext.
const EncryptedPrefix = "enc:v1:"

func loadKey() ([]byte, error) {
	raw := os.Getenv("ENCRYPTION_KEY")
	if raw == "" {
		return nil, errors.New("ENCRYPTION_KEY is not set")
	}
	k, err := hex.DecodeString(raw)
	if err != nil {
		return nil, errors.New("ENCRYPTION_KEY must be hex-encoded")
	}
	if len(k) != 32 {
		return nil, errors.New("ENCRYPTION_KEY must decode to exactly 32 bytes (64 hex chars)")
	}
	return k, nil
}

// Encrypt returns EncryptedPrefix + base64(nonce || ciphertext).
// Empty input returns empty output (nothing to encrypt).
// Returns an error if ENCRYPTION_KEY is missing or invalid — callers must
// treat that as a fail-closed condition and refuse to persist plaintext.
func Encrypt(plaintext string) (string, error) {
	if plaintext == "" {
		return "", nil
	}
	k, err := loadKey()
	if err != nil {
		return "", err
	}
	block, err := aes.NewCipher(k)
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}
	ct := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return EncryptedPrefix + base64.StdEncoding.EncodeToString(ct), nil
}

// Decrypt reverses Encrypt. Values without the prefix are returned as-is
// (legacy plaintext rows). On any error, the raw value is returned unchanged —
// decryption failures must not brick reads.
func Decrypt(value string) string {
	if value == "" {
		return ""
	}
	if !strings.HasPrefix(value, EncryptedPrefix) {
		return value
	}
	k, err := loadKey()
	if err != nil {
		return value
	}
	raw := strings.TrimPrefix(value, EncryptedPrefix)
	data, err := base64.StdEncoding.DecodeString(raw)
	if err != nil {
		return value
	}
	block, err := aes.NewCipher(k)
	if err != nil {
		return value
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return value
	}
	if len(data) < gcm.NonceSize() {
		return value
	}
	nonce, ct := data[:gcm.NonceSize()], data[gcm.NonceSize():]
	pt, err := gcm.Open(nil, nonce, ct, nil)
	if err != nil {
		return value
	}
	return string(pt)
}
