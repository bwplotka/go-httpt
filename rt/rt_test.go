package rt

import (
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/Bplotka/go-httpt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStringResponse(t *testing.T) {
	r, err := http.NewRequest(string(httpt.GET), "/test/path", nil)
	require.NoError(t, err)

	s := httpt.NewRawServer()
	s.Push(StringResponseFunc(http.StatusBadRequest, "test_msg"))
	resp, err := s.HTTPClient().Do(r)
	require.NoError(t, err)

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	body, err := ioutil.ReadAll(resp.Body)
	assert.NoError(t, err)

	assert.Equal(t, "test_msg", string(body))
}

func TestJSONResponse(t *testing.T) {
	r, err := http.NewRequest(string(httpt.GET), "/test/path", nil)
	require.NoError(t, err)

	s := httpt.NewRawServer()
	s.Push(JSONResponseFunc(http.StatusBadRequest, []byte(`{"error": "test_err"}`)))
	resp, err := s.HTTPClient().Do(r)
	require.NoError(t, err)

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	body, err := ioutil.ReadAll(resp.Body)
	assert.NoError(t, err)

	assert.Equal(t, `{"error": "test_err"}`, string(body))
}
