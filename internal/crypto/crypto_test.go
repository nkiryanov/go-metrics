package crypto

import (
	"crypto/ecdh"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func generateX25519Pair(t *testing.T) (*ecdh.PrivateKey, *ecdh.PublicKey) {
	t.Helper()
	priv, err := ecdh.X25519().GenerateKey(rand.Reader)
	require.NoError(t, err)
	return priv, priv.PublicKey()
}

func writePrivKeyPEM(t *testing.T, priv *ecdh.PrivateKey) string {
	t.Helper()
	der, err := x509.MarshalPKCS8PrivateKey(priv)
	require.NoError(t, err)
	block := &pem.Block{Type: "PRIVATE KEY", Bytes: der}
	path := filepath.Join(t.TempDir(), "priv.pem")
	err = os.WriteFile(path, pem.EncodeToMemory(block), 0o600)
	require.NoError(t, err)
	return path
}

func writePubKeyPEM(t *testing.T, pub *ecdh.PublicKey) string {
	t.Helper()
	der, err := x509.MarshalPKIXPublicKey(pub)
	require.NoError(t, err)
	block := &pem.Block{Type: "PUBLIC KEY", Bytes: der}
	path := filepath.Join(t.TempDir(), "pub.pem")
	err = os.WriteFile(path, pem.EncodeToMemory(block), 0o644)
	require.NoError(t, err)
	return path
}

func TestEncryptDecrypt_Roundtrip(t *testing.T) {
	priv, pub := generateX25519Pair(t)
	plaintext := []byte("hello, metrics!")

	ciphertext, err := Encrypt(pub, plaintext)
	require.NoError(t, err)

	got, err := Decrypt(priv, ciphertext)
	require.NoError(t, err)
	assert.Equal(t, plaintext, got)
}

func TestDecrypt_WrongKey(t *testing.T) {
	_, pub := generateX25519Pair(t)
	wrongPriv, _ := generateX25519Pair(t)

	ciphertext, err := Encrypt(pub, []byte("secret"))
	require.NoError(t, err)

	_, err = Decrypt(wrongPriv, ciphertext)
	assert.Error(t, err)
}

func TestDecrypt_TruncatedData(t *testing.T) {
	priv, pub := generateX25519Pair(t)

	ciphertext, err := Encrypt(pub, []byte("secret"))
	require.NoError(t, err)

	_, err = Decrypt(priv, ciphertext[:20])
	assert.Error(t, err)
}

func TestDecrypt_EmptyData(t *testing.T) {
	priv, _ := generateX25519Pair(t)

	_, err := Decrypt(priv, []byte{})
	assert.Error(t, err)
}

func TestLoadPublicKey(t *testing.T) {
	t.Run("valid key", func(t *testing.T) {
		_, pub := generateX25519Pair(t)
		path := writePubKeyPEM(t, pub)

		got, err := LoadPublicKey(path)
		require.NoError(t, err)
		assert.Equal(t, pub.Bytes(), got.Bytes())
	})

	t.Run("file not found", func(t *testing.T) {
		_, err := LoadPublicKey("/nonexistent/path/pub.pem")
		assert.Error(t, err)
	})

	t.Run("wrong key type (RSA)", func(t *testing.T) {
		rsaKey, err := rsa.GenerateKey(rand.Reader, 2048)
		require.NoError(t, err)
		der, err := x509.MarshalPKIXPublicKey(&rsaKey.PublicKey)
		require.NoError(t, err)
		block := &pem.Block{Type: "PUBLIC KEY", Bytes: der}
		path := filepath.Join(t.TempDir(), "rsa_pub.pem")
		err = os.WriteFile(path, pem.EncodeToMemory(block), 0o644)
		require.NoError(t, err)

		_, err = LoadPublicKey(path)
		assert.Error(t, err)
	})
}

func TestLoadPrivateKey(t *testing.T) {
	t.Run("valid key", func(t *testing.T) {
		priv, _ := generateX25519Pair(t)
		path := writePrivKeyPEM(t, priv)

		got, err := LoadPrivateKey(path)
		require.NoError(t, err)
		assert.Equal(t, priv.Bytes(), got.Bytes())
	})

	t.Run("file not found", func(t *testing.T) {
		_, err := LoadPrivateKey("/nonexistent/path/priv.pem")
		assert.Error(t, err)
	})

	t.Run("wrong key type (RSA)", func(t *testing.T) {
		rsaKey, err := rsa.GenerateKey(rand.Reader, 2048)
		require.NoError(t, err)
		der, err := x509.MarshalPKCS8PrivateKey(rsaKey)
		require.NoError(t, err)
		block := &pem.Block{Type: "PRIVATE KEY", Bytes: der}
		path := filepath.Join(t.TempDir(), "rsa_priv.pem")
		err = os.WriteFile(path, pem.EncodeToMemory(block), 0o600)
		require.NoError(t, err)

		_, err = LoadPrivateKey(path)
		assert.Error(t, err)
	})
}
