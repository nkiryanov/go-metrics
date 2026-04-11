package httpreporter

import (
	"bytes"
	"compress/gzip"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"

	"github.com/nkiryanov/go-metrics/internal/crypto"
)

// Encode postCxt data to json and compress with gzip
func (r *HTTPReporter) jsonGzipMiddleware(postCtx *postContext) error {
	// Do nothing if data empty
	if postCtx.data == nil {
		return nil
	}

	var body bytes.Buffer
	var err error

	gzipWriter := gzip.NewWriter(&body)
	encoder := json.NewEncoder(gzipWriter)

	err = encoder.Encode(postCtx.data)
	if err != nil {
		return err
	}
	err = gzipWriter.Close()
	if err != nil {
		return err
	}

	postCtx.headers["Content-Encoding"] = "gzip"
	postCtx.headers["Content-Type"] = "application/json"
	postCtx.body = &body
	postCtx.data = nil

	return nil
}

// encrypt body if public key is set
func (r *HTTPReporter) encryptMiddleware(postCtx *postContext) error {
	if r.publicKey == nil || postCtx.body == nil {
		return nil
	}

	encrypted, err := crypto.Encrypt(r.publicKey, postCtx.body.Bytes())
	if err != nil {
		return err
	}

	postCtx.body = bytes.NewBuffer(encrypted)
	return nil
}

// if secret key is set, calculate hash and set HashSHA256 header
func (r *HTTPReporter) hmacSha256Middleware(postCtx *postContext) error {
	// Do nothing if secret key not set or body empty
	if r.secretKey == "" || postCtx.body == nil {
		return nil
	}

	h := hmac.New(sha256.New, []byte(r.secretKey))
	h.Write(postCtx.body.Bytes())

	postCtx.headers["HashSHA256"] = hex.EncodeToString(h.Sum(nil))

	return nil
}
