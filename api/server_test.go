package api_test

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Ullaakut/digosaur/api"
	"github.com/Ullaakut/digosaur/pkg/influx"
	"github.com/hamba/cmd/v2/observe"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

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

func setupTestServer(store *storeMock, t *testing.T) string {
	t.Helper()

	obsvr := observe.NewFake()

	srv := api.New(store, obsvr)

	httpSrv := httptest.NewServer(srv)
	t.Cleanup(func() { httpSrv.Close() })

	return httpSrv.URL
}

type storeMock struct {
	mock.Mock
}

func (m *storeMock) Add(_ context.Context, pt influx.Point) error {
	args := m.Called(pt)
	return args.Error(0)
}
