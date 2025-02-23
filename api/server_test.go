package api_test

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gamefabric/go-template/api"
	"github.com/hamba/cmd/v2/observe"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestServer_HandleHello(t *testing.T) {
	srvUrl := setupTestServer(t)

	resp := requireDoRequest(t, http.MethodGet, srvUrl+"/", nil)
	t.Cleanup(func() { _ = resp.Body.Close() })

	require.Equal(t, http.StatusOK, resp.StatusCode)

	got, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	assert.Equal(t, `{"message":"Hello, World!"}`, string(got))
}

func requireDoRequest(t *testing.T, method, url string, body []byte) *http.Response {
	t.Helper()

	var r io.Reader
	if body != nil {
		r = bytes.NewReader(body)
	}
	req, err := http.NewRequestWithContext(context.Background(), method, url, r)
	require.NoError(t, err)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	return resp
}

func setupTestServer(t *testing.T) string {
	t.Helper()

	obsvr := observe.NewFake()

	srv := api.New(obsvr)

	httpSrv := httptest.NewServer(srv)
	t.Cleanup(func() { httpSrv.Close() })

	return httpSrv.URL
}
