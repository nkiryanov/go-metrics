package reporter

import (
	"bytes"
	"compress/gzip"
	"errors"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/nkiryanov/go-metrics/internal/models"

	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func decompress(buf *bytes.Buffer) string {
	decoder, _ := gzip.NewReader(buf)
	defer decoder.Close() // nolint:errcheck

	data, _ := io.ReadAll(decoder)
	return string(data)
}

func TestHTTPReporter_post(t *testing.T) {
	reporter := NewHTTPReporter("http://test.server", &http.Client{}, nil)
	metric := models.Metric{Name: "test", Type: "counter", Delta: 1} // Any valid metric should ok

	httpmock.ActivateNonDefault(reporter.client)
	t.Cleanup(httpmock.DeactivateAndReset)

	t.Run("successful request", func(t *testing.T) {
		httpmock.RegisterResponder(
			"POST",
			"http://test.server/some-shit",
			httpmock.NewStringResponder(200, "OK"),
		)

		err := reporter.post("/some-shit", metric)

		assert.NoError(t, err)
	})

	t.Run("connection error", func(t *testing.T) {
		httpmock.RegisterResponder(
			"POST",
			"http://test.server/some-shit",
			httpmock.NewErrorResponder(errors.New("connection refused")),
		)
		var errReport *reportError

		err := reporter.post("/some-shit", metric)

		require.Error(t, err)
		require.True(t, errors.As(err, &errReport), "post has to return reportErr error")
		assert.True(t, errReport.connErr, "connErr must be true on connection errors")
	})

	t.Run("server error 500", func(t *testing.T) {
		httpmock.RegisterResponder(
			"POST",
			"http://test.server/some-shit",
			httpmock.NewStringResponder(500, "Internal Server Error"),
		)
		var errReport *reportError

		err := reporter.post("/some-shit", metric)

		require.Error(t, err)
		require.True(t, errors.As(err, &errReport), "post has to return reportErr error")
		assert.False(t, errReport.connErr, "connError must false if server response with valid response")
	})

	t.Run("proper gzip encoding and headers", func(t *testing.T) {
		var capturedBody = &bytes.Buffer{}
		var capturedHeaders http.Header

		httpmock.RegisterResponder("POST", "http://test.server/update",
			func(req *http.Request) (*http.Response, error) {
				_, err := capturedBody.ReadFrom(req.Body)
				require.NoError(t, err)
				capturedHeaders = req.Header
				return httpmock.NewStringResponse(200, "OK"), nil
			})

		err := reporter.post("/update", metric)

		require.NoError(t, err)
		assert.Equal(t, "gzip", capturedHeaders.Get("Content-Encoding"))
		assert.Equal(t, "application/json", capturedHeaders.Get("Content-Type"))

		expectedJSON := `{"id":"test","type":"counter","delta":1}`
		assert.JSONEq(t, expectedJSON, decompress(capturedBody))
	})
}

func TestHTTPReporter_postWithRetry(t *testing.T) {
	reporter := NewHTTPReporter(
		"http://test.server",
		&http.Client{},
		[]time.Duration{ // Two retries max
			100 * time.Millisecond,
			200 * time.Millisecond,
		},
	)
	httpmock.ActivateNonDefault(reporter.client)
	t.Cleanup(httpmock.DeactivateAndReset)
	metric := models.Metric{Name: "test", Type: "counter", Delta: 1} // Any valid metric should ok

	t.Run("return ok if first response connection error", func(t *testing.T) {
		httpmock.RegisterResponder("POST", "http://test.server/update",
			// On first request return connection error, but later calls return ok
			func() httpmock.Responder {
				attempt := 0
				return func(req *http.Request) (*http.Response, error) {
					attempt += 1
					if attempt == 1 {
						return httpmock.ConnectionFailure(req)
					}
					return httpmock.NewStringResponse(200, "ok"), nil
				}
			}(),
		)

		err := reporter.postWithRetry("/update", metric)

		assert.NoError(t, err)
		assert.Equal(t, 2, httpmock.GetCallCountInfo()["POST http://test.server/update"])
	})

	t.Run("do not retry if not connection err", func(t *testing.T) {
		callCount := 0
		httpmock.RegisterResponder("POST", "http://test.server/update",
			// On first request return 400 fail, on later call return ok
			func(res *http.Request) (*http.Response, error) {
				callCount += 1
				if callCount == 1 {
					return httpmock.NewStringResponse(400, "fail"), nil
				}
				return httpmock.NewStringResponse(200, "ok"), nil
			},
		)

		err := reporter.postWithRetry("/update", metric)

		require.Error(t, err, "Should return err case not first response not connection error")
		require.Equal(t, 1, callCount, "Should not retry cause not connection error occurred")
	})

	t.Run("stop trying if maxRetries happened", func(t *testing.T) {
		callCount := 0
		httpmock.RegisterResponder("POST", "http://test.server/update",
			// Response connection error three times than response ok
			func(res *http.Request) (*http.Response, error) {
				callCount += 1
				if callCount <= 3 {
					return httpmock.ConnectionFailure(res)
				}
				return httpmock.NewStringResponse(200, "ok"), nil
			},
		)

		err := reporter.postWithRetry("/update", metric)

		require.Error(t, err, "Should return err cause maxRetries exceeded")
		require.Equal(t, 3, callCount, "Should called 3 times (first attempt and 2 retries)")
	})
}

func TestHTTPReporter_Smoke(t *testing.T) {
	reporter := NewHTTPReporter("http://pornhub.com", &http.Client{}, nil)
	httpmock.ActivateNonDefault(reporter.client)
	t.Cleanup(httpmock.DeactivateAndReset)

	t.Run("report once ok", func(t *testing.T) {
		httpmock.RegisterResponder("POST", `http://pornhub.com/update`, httpmock.NewStringResponder(200, "got it!"))

		err := reporter.ReportOnce(models.Metric{Name: "smth", Type: "counter", Delta: 2})

		require.NoError(t, err)
		info := httpmock.GetCallCountInfo()
		assert.Equal(t, 1, info["POST http://pornhub.com/update"])
	})

	t.Run("batch report ok", func(t *testing.T) {
		httpmock.RegisterResponder("POST", `http://pornhub.com/updates`, httpmock.NewStringResponder(200, "got it!"))

		err := reporter.ReportBatch(
			[]models.Metric{
				{Name: "smth", Type: "counter", Delta: 2},     // Valid metric
				{Name: "ya-smth", Type: "counter", Delta: 22}, // Yet another valid metric
			},
		)

		require.NoError(t, err)
		info := httpmock.GetCallCountInfo()
		assert.Equal(t, 1, info["POST http://pornhub.com/updates"])
	})
}
