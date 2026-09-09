package core

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"testing"

	sdkerrors "github.com/LerianStudio/lerian-sdk-golang/pkg/errors"
	"github.com/LerianStudio/lerian-sdk-golang/pkg/retry"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// errClosingBody is a response body that reads cleanly and then fails to close,
// the shape a broken chunked encoding or an aborted stream produces: the bytes
// that arrived look like a complete payload and only Close reports otherwise.
type errClosingBody struct {
	reader   io.Reader
	closeErr error
}

func (b *errClosingBody) Read(p []byte) (int, error) { return b.reader.Read(p) }
func (b *errClosingBody) Close() error               { return b.closeErr }

// bodyRoundTripper answers every request with a 200 carrying the given body.
type bodyRoundTripper struct {
	body io.ReadCloser
}

func (rt *bodyRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	return &http.Response{
		StatusCode: http.StatusOK,
		Header:     http.Header{"Content-Type": []string{"application/json"}},
		Body:       rt.body,
		Request:    req,
	}, nil
}

func newClosePathBackend(body io.ReadCloser) *BackendImpl {
	return NewBackendImpl(BackendConfig{
		BaseURL:     "http://example.invalid",
		RetryConfig: retry.Config{},
		HTTPClient:  &http.Client{Transport: &bodyRoundTripper{body: body}},
	})
}

// TestDoRejectsBodyThatFailsToClose pins the failure the dropped Close error
// used to hide. The payload below is TRUNCATED: it is the first half of a
// larger document, so it decodes into a value that looks complete. Only the
// close failure says the response never finished arriving. Before the fix the
// request succeeded and the caller acted on the partial value.
func TestDoRejectsBodyThatFailsToClose(t *testing.T) {
	t.Parallel()

	truncated := `{"id":"acct-1","name":"Acme`

	backend := newClosePathBackend(&errClosingBody{
		reader:   bytes.NewBufferString(truncated),
		closeErr: errors.New("unexpected EOF reading trailer"),
	})

	resp, err := backend.Do(context.Background(), Request{Method: http.MethodGet, Path: "/accounts/acct-1"})

	require.Error(t, err, "a body that fails to close did not arrive intact and must not read as success")
	assert.Nil(t, resp)
	assert.True(t, errors.Is(err, sdkerrors.ErrInternal),
		"an incomplete response is a read failure, the same as any other")
	assert.Contains(t, err.Error(), "failed to read response body")
}

// TestDoAcceptsBodyThatClosesCleanly is the other half: the guard must not turn
// every response into an error.
func TestDoAcceptsBodyThatClosesCleanly(t *testing.T) {
	t.Parallel()

	backend := newClosePathBackend(io.NopCloser(bytes.NewBufferString(`{"id":"acct-1"}`)))

	resp, err := backend.Do(context.Background(), Request{Method: http.MethodGet, Path: "/accounts/acct-1"})

	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.JSONEq(t, `{"id":"acct-1"}`, string(resp.Body))
}
