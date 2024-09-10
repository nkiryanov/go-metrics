package reporter

import (
	"bytes"
	"compress/gzip"
	"io"
	"net/http"
	"testing"

	"github.com/nkiryanov/go-metrics/internal/models"

	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func decompress(buf *bytes.Buffer) string {
	decoder, _ := gzip.NewReader(buf)
	defer decoder.Close()

	data, _ := io.ReadAll(decoder)
	return string(data)
}

func TestHTTPReporter_ReportOnce(t *testing.T) {
	rept := NewHTTPReporter("http://reports.server", &http.Client{})

	httpmock.ActivateNonDefault(rept.client)
	defer httpmock.DeactivateAndReset()

	tests := []struct {
		name         string
		metric       *models.Metric
		expectedBody string
	}{
		{
			"report counter, ok",
			&models.Metric{ID: "poll-count", MType: "counter", Delta: 213},
			`{"id": "poll-count", "type": "counter", "delta": 213}`,
		},
		{
			"report gauge, ok",
			&models.Metric{ID: "mem-usage", MType: "gauge", Value: 239239.3983},
			`{"id": "mem-usage", "type": "gauge", "value": 239239.3983}`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var capturedBody = &bytes.Buffer{}
			var capturedEncoding string
			httpmock.RegisterResponder("POST", "http://reports.server/update",
				func(req *http.Request) (*http.Response, error) {
					// body, _ := io.ReadAll(req.Body)
					_, _ = capturedBody.ReadFrom(req.Body)
					capturedEncoding = req.Header.Get("Content-Encoding")
					return httpmock.NewStringResponse(200, "got it!"), nil
				})

			err := rept.ReportOnce(tc.metric)

			require.NoError(t, err)
			require.Equal(t, 1, httpmock.GetCallCountInfo()["POST http://reports.server/update"])
			require.Contains(t, "gzip", capturedEncoding)
			assert.JSONEq(t, tc.expectedBody, decompress(capturedBody))
		})
	}

}

func TestHTTPReporter_ReportBatch(t *testing.T) {
	rept := NewHTTPReporter("http://pornhub.com", &http.Client{})

	httpmock.ActivateNonDefault(rept.client)
	defer httpmock.DeactivateAndReset()

	t.Run("do batch reports", func(t *testing.T) {
		httpmock.RegisterResponder("POST", `http://pornhub.com/update`, httpmock.NewStringResponder(200, "got it!"))

		err := rept.ReportBatch(
			[]models.Metric{
				{ID: "smth", MType: "counter", Delta: 2},     // Valid metric
				{ID: "ya-smth", MType: "counter", Delta: 22}, // Yet another valid metric
			},
		)

		require.NoError(t, err)
		info := httpmock.GetCallCountInfo()
		assert.Equal(t, 2, info["POST http://pornhub.com/update"])
	})

	t.Run("return any happened error", func(t *testing.T) {
		httpmock.RegisterResponder("POST", `http://pornhub.com/update`, httpmock.NewStringResponder(500, "go fuck yourself!"))

		err := rept.ReportBatch(
			[]models.Metric{
				{ID: "smth", MType: "counter", Delta: 2},     // Valid metric
				{ID: "ya-smth", MType: "counter", Delta: 22}, // Yet another valid metric
				{ID: "smth-invalid", MType: "not-valid"},     // Invalid
				{ID: "fuck", MType: "not-valid"},             // Ya invalid
			},
		)

		require.Error(t, err)
		info := httpmock.GetCallCountInfo()
		assert.Equal(t, 4, info["POST http://pornhub.com/update"])
	})
}
