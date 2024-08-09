package reporter

import (
	"testing"

	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/go-resty/resty/v2"
)

type mValue string

func (m mValue) String() string { return string(m) }

func TestHTTPReporter_ReportOnce(t *testing.T) {
	rept := NewHTTPReporter("http://reports.server", resty.New())

	httpmock.ActivateNonDefault(rept.client.GetClient())
	defer httpmock.DeactivateAndReset()

	tests := []struct {
		name        string
		metric      *Metric
		expectedReq string
		responder   httpmock.Responder
		isError     bool
	}{
		{
			"report counter, ok",
			&Metric{Type: "counter", Name: "poll-count", Value: mValue("213")},
			"POST http://reports.server/update/counter/poll-count/213",
			httpmock.NewStringResponder(200, "got it!"),
			false,
		},
		{
			"report gauge, ok",
			&Metric{Type: "gauge", Name: "mem-usage", Value: mValue("239239.3983")},
			"POST http://reports.server/update/gauge/mem-usage/239239.3983",
			httpmock.NewStringResponder(200, "got it!"),
			false,
		},
		{
			"report counter, bad",
			&Metric{Type: "counter", Name: "poll-count", Value: mValue("777")},
			"POST http://reports.server/update/counter/poll-count/777",
			httpmock.NewStringResponder(500, "go fuck yourself!"),
			true,
		},
		{
			"report not valid url, bad",
			&Metric{Type: "not-valid", Name: "not-valid", Value: mValue("smth")},
			"requested invalid server",
			nil,
			true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			httpmock.RegisterResponder("POST", `=~^http://reports\.server/[counter|gauge].*`,
				tc.responder,
			)

			err := rept.ReportOnce(tc.metric)

			if tc.isError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
			if tc.expectedReq != "requested invalid server" {
				info := httpmock.GetCallCountInfo()
				assert.Equal(t, info[tc.expectedReq], 1)
			}
		})
	}

}

func TestHTTPReporter_ReportBatch(t *testing.T) {
	rept := NewHTTPReporter("http://pornhub.com", resty.New())

	httpmock.ActivateNonDefault(rept.client.GetClient())
	defer httpmock.DeactivateAndReset()

	// Server return ok on counter update
	httpmock.RegisterResponder("POST", `=~^http://pornhub\.com/update/counter`,
		httpmock.NewStringResponder(200, "got it!"),
	)

	// Server return 500 server error on not-valid metric
	httpmock.RegisterResponder("POST", `=~^http://pornhub\.com/update/not-valid`,
		httpmock.NewStringResponder(500, "go fuck yourself!"),
	)

	t.Run("do batch reports", func(t *testing.T) {
		httpmock.ZeroCallCounters()

		err := rept.ReportBatch(
			[]*Metric{
				{Type: "counter", Name: "smth", Value: mValue("2")},     // Valid metric
				{Type: "counter", Name: "ya-smth", Value: mValue("22")}, // Yet another valid metric
			},
		)

		require.NoError(t, err)
		info := httpmock.GetCallCountInfo()
		assert.Equal(t, 1, info["POST http://pornhub.com/update/counter/smth/2"])
		assert.Equal(t, 1, info["POST http://pornhub.com/update/counter/ya-smth/22"])
	})

	t.Run("return any happened error", func(t *testing.T) {
		httpmock.ZeroCallCounters()

		err := rept.ReportBatch(
			[]*Metric{
				{Type: "counter", Name: "smth", Value: mValue("2")},            // Valid metric
				{Type: "counter", Name: "ya-smth", Value: mValue("22")},        // Yet another valid metric
				{Type: "not-valid", Name: "smth-invalid", Value: mValue("hi")}, // Invalid
				{Type: "not-valid", Name: "fuck", Value: mValue("me")},         // Ya invalid
			},
		)

		require.Error(t, err)
		info := httpmock.GetCallCountInfo()
		assert.Equal(t, 1, info["POST http://pornhub.com/update/counter/smth/2"])
		assert.Equal(t, 1, info["POST http://pornhub.com/update/counter/ya-smth/22"])
		assert.Equal(t, 1, info["POST http://pornhub.com/update/not-valid/smth-invalid/hi"])
		assert.Equal(t, 1, info["POST http://pornhub.com/update/not-valid/fuck/me"])
	})
}
