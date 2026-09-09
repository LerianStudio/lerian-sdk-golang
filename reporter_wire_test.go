package lerian

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/LerianStudio/lerian-sdk-golang/reporter"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// wireRecorder captures the method and path of every request that reaches the
// stub server, so a test can assert what the SDK actually put on the wire.
type wireRecorder struct {
	mu    sync.Mutex
	calls []string
}

func (w *wireRecorder) record(method, path string) {
	w.mu.Lock()
	defer w.mu.Unlock()

	w.calls = append(w.calls, method+" "+path)
}

func (w *wireRecorder) snapshot() []string {
	w.mu.Lock()
	defer w.mu.Unlock()

	out := make([]string, len(w.calls))
	copy(out, w.calls)

	return out
}

// newReporterWireServer starts a stub Reporter server that records requests and
// answers every one with an empty JSON list envelope, which parses as both a
// single resource and a page.
func newReporterWireServer(t *testing.T) (*httptest.Server, *wireRecorder) {
	t.Helper()

	rec := &wireRecorder{}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rec.record(r.Method, r.URL.Path)
		w.Header().Set("Content-Type", "application/json")

		_, err := w.Write([]byte(`{"items":[],"pagination":{"total":0,"limit":10}}`))
		require.NoError(t, err)
	}))
	t.Cleanup(server.Close)

	return server, rec
}

func newReporterWireClient(t *testing.T, baseURL string) *Client {
	t.Helper()

	client, err := New(Config{
		Reporter: &reporter.Config{
			BaseURL: baseURL + "/v1",
		},
	})
	require.NoError(t, err)

	return client
}

// TestReporterDataSourcePathsMatchService pins the data-source paths to the
// only ones Reporter serves: /v1/data-sources and /v1/data-sources/{id}. The
// SDK shipped the unhyphenated spelling, so every data-source call 404'd.
func TestReporterDataSourcePathsMatchService(t *testing.T) {
	t.Parallel()

	server, rec := newReporterWireServer(t)
	client := newReporterWireClient(t, server.URL)

	ctx := context.Background()

	_, err := client.Reporter.DataSources.Get(ctx, "ds-1")
	require.NoError(t, err)

	_, err = client.Reporter.DataSources.List(ctx, nil).Collect(ctx)
	require.NoError(t, err)

	assert.Equal(t, []string{
		"GET /v1/data-sources/ds-1",
		"GET /v1/data-sources",
	}, rec.snapshot())
}
