package api_test

import (
	"net/http"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestServer_HandleApple(t *testing.T) {
	store := &storeMock{}
	store.On("Add", mock.Anything).Return(nil)

	srvUrl := setupTestServer(store, t)

	data, err := os.ReadFile("testdata/sample.json")
	require.NoError(t, err)

	resp := requireDoRequest(t, http.MethodPost, srvUrl+"/apple", data)
	t.Cleanup(func() { _ = resp.Body.Close() })

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	store.AssertExpectations(t)
}
