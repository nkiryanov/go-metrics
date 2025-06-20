package httpreporter

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
)

// Encode postCxt data to json and compress with gzip
func jsonGzipMiddleware(postCtx *postContext) error {
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
