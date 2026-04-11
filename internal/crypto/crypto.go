package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/ecdh"
	"crypto/rand"
	"crypto/sha256"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"os"

	"golang.org/x/crypto/hkdf"
)

// hkdfInfo is the fixed info string used for HKDF key derivation.
// both Encrypt and Decrypt must use the same value.
var hkdfInfo = []byte("go-metrics x25519 v1")

// deriveKey derives a 32-byte AES key from the X25519 shared secret using HKDF-SHA256.
func deriveKey(sharedSecret []byte) ([]byte, error) {
	r := hkdf.New(sha256.New, sharedSecret, nil, hkdfInfo)
	key := make([]byte, 32)
	if _, err := io.ReadFull(r, key); err != nil {
		return nil, fmt.Errorf("hkdf read: %w", err)
	}
	return key, nil
}

// Encrypt encrypts plaintext using the server's X25519 public key.
// returns: [32B ephemeral pubkey] + [12B nonce] + [ciphertext + 16B GCM tag]
func Encrypt(pubKey *ecdh.PublicKey, plaintext []byte) ([]byte, error) {
	ephemeral, err := ecdh.X25519().GenerateKey(nil)
	if err != nil {
		return nil, fmt.Errorf("generate ephemeral key: %w", err)
	}

	sharedSecret, err := ephemeral.ECDH(pubKey)
	if err != nil {
		return nil, fmt.Errorf("ecdh: %w", err)
	}

	aesKey, err := deriveKey(sharedSecret)
	if err != nil {
		return nil, err
	}

	block, err := aes.NewCipher(aesKey)
	if err != nil {
		return nil, fmt.Errorf("aes cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("gcm: %w", err)
	}

	nonce := make([]byte, gcm.NonceSize()) // 12 bytes
	if _, err = rand.Read(nonce); err != nil {
		return nil, fmt.Errorf("nonce: %w", err)
	}

	ciphertext := gcm.Seal(nil, nonce, plaintext, nil)

	result := make([]byte, 0, 32+len(nonce)+len(ciphertext))
	result = append(result, ephemeral.PublicKey().Bytes()...)
	result = append(result, nonce...)
	result = append(result, ciphertext...)
	return result, nil
}

// Decrypt decrypts data encrypted by Encrypt using the server's X25519 private key.
// data layout: [0:32] ephemeral pubkey | [32:44] nonce | [44:] ciphertext+tag
func Decrypt(privKey *ecdh.PrivateKey, data []byte) ([]byte, error) {
	const minLen = 32 + 12 + 16 // ephemeral key + nonce + GCM tag
	if len(data) < minLen {
		return nil, errors.New("data too short")
	}

	ephemeralPub, err := ecdh.X25519().NewPublicKey(data[:32])
	if err != nil {
		return nil, fmt.Errorf("parse ephemeral pubkey: %w", err)
	}

	sharedSecret, err := privKey.ECDH(ephemeralPub)
	if err != nil {
		return nil, fmt.Errorf("ecdh: %w", err)
	}

	aesKey, err := deriveKey(sharedSecret)
	if err != nil {
		return nil, err
	}

	block, err := aes.NewCipher(aesKey)
	if err != nil {
		return nil, fmt.Errorf("aes cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("gcm: %w", err)
	}

	nonce := data[32:44]
	ciphertext := data[44:]

	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("decrypt: %w", err)
	}

	return plaintext, nil
}

// LoadPublicKey loads an X25519 public key from a PEM file (PKIX format).
func LoadPublicKey(path string) (*ecdh.PublicKey, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read public key: %w", err)
	}

	block, _ := pem.Decode(data)
	if block == nil {
		return nil, errors.New("failed to decode PEM block")
	}

	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("parse public key: %w", err)
	}

	ecdhPub, ok := pub.(*ecdh.PublicKey)
	if !ok {
		return nil, fmt.Errorf("expected *ecdh.PublicKey, got %T", pub)
	}

	return ecdhPub, nil
}

// LoadPrivateKey loads an X25519 private key from a PEM file (PKCS#8 format).
func LoadPrivateKey(path string) (*ecdh.PrivateKey, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read private key: %w", err)
	}

	block, _ := pem.Decode(data)
	if block == nil {
		return nil, errors.New("failed to decode PEM block")
	}

	priv, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("parse private key: %w", err)
	}

	ecdhPriv, ok := priv.(*ecdh.PrivateKey)
	if !ok {
		return nil, fmt.Errorf("expected *ecdh.PrivateKey, got %T", priv)
	}

	return ecdhPriv, nil
}
