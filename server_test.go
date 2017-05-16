package httpt

import (
	"bytes"
	"errors"
	"io/ioutil"
	"net/http"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func mockedRT(_ *http.Request) (*http.Response, error) {
	return &http.Response{
		StatusCode: http.StatusOK,
		Body:       ioutil.NopCloser(bytes.NewBufferString("hello")),
	}, nil
}

func TestServer_NoDefaultRT_Err(t *testing.T) {
	r, err := http.NewRequest(string(GET), "/test/path", nil)
	require.NoError(t, err)

	s := NewRawServer()
	_, err = s.HTTPClient().Do(r)
	require.Error(t, err, "Server without Default RoundTripper should fail")
}

func TestServer_DefaultRT_OK(t *testing.T) {
	r, err := http.NewRequest(string(GET), "/test/path", nil)
	require.NoError(t, err)

	s := NewRawServer()
	s.DefaultRoundTrip = mockedRT
	resp, err := s.HTTPClient().Do(r)
	require.NoError(t, err)

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	body, err := ioutil.ReadAll(resp.Body)
	assert.NoError(t, err)

	assert.Equal(t, "hello", string(body))

	assert.Equal(t, 0, s.Len())
}

func TestServer_ConnErr_Err(t *testing.T) {
	r, err := http.NewRequest(string(GET), "/test/path", nil)
	require.NoError(t, err)

	s := NewRawServer()
	expectedErr := errors.New("Context Deadline exceeded")
	s.DefaultRoundTrip = FailureFunc(expectedErr)
	_, err = s.HTTPClient().Do(r)
	require.Error(t, err)

	urlErr, ok := err.(*url.Error)
	require.True(t, ok)
	assert.Equal(t, expectedErr, urlErr.Err)

	assert.Equal(t, 0, s.Len())
}

func TestServer_OnAnyMethodAnyPathRT_OK(t *testing.T) {
	r, err := http.NewRequest(string(GET), "/test/path", nil)
	require.NoError(t, err)

	s := NewRawServer()
	s.Push(mockedRT)
	resp, err := s.HTTPClient().Do(r)
	require.NoError(t, err)

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	body, err := ioutil.ReadAll(resp.Body)
	assert.NoError(t, err)

	assert.Equal(t, "hello", string(body))

	assert.Equal(t, 0, s.Len())
}

func TestServer_OnAnyMethodProperPathRT_OK(t *testing.T) {
	r, err := http.NewRequest(string(GET), "/test/path", nil)
	require.NoError(t, err)

	s := NewRawServer()
	s.On(ANY, "/test/path").Push(mockedRT)
	resp, err := s.HTTPClient().Do(r)
	require.NoError(t, err)

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	body, err := ioutil.ReadAll(resp.Body)
	assert.NoError(t, err)

	assert.Equal(t, "hello", string(body))

	assert.Equal(t, 0, s.Len())
}

func TestServer_OnProperMethodAnyPathRT_OK(t *testing.T) {
	r, err := http.NewRequest(string(GET), "/test/path", nil)
	require.NoError(t, err)

	s := NewRawServer()
	s.On(GET, AnyPath).Push(mockedRT)
	resp, err := s.HTTPClient().Do(r)
	require.NoError(t, err)

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	body, err := ioutil.ReadAll(resp.Body)
	assert.NoError(t, err)

	assert.Equal(t, "hello", string(body))

	assert.Equal(t, 0, s.Len())
}

func TestServer_OnProperMethodProperPathRT_OK(t *testing.T) {
	r, err := http.NewRequest(string(GET), "/test/path", nil)
	require.NoError(t, err)

	s := NewRawServer()
	s.On(GET, "/test/path").Push(mockedRT)
	resp, err := s.HTTPClient().Do(r)
	require.NoError(t, err)

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	body, err := ioutil.ReadAll(resp.Body)
	assert.NoError(t, err)

	assert.Equal(t, "hello", string(body))

	assert.Equal(t, 0, s.Len())
}

func TestServer_OnWrongMethodProperPathRT_OK(t *testing.T) {
	r, err := http.NewRequest(string(GET), "/test/path", nil)
	require.NoError(t, err)

	s := NewRawServer()
	s.On(POST, "/test/path").Push(mockedRT)
	_, err = s.HTTPClient().Do(r)
	require.Error(t, err)

	assert.Equal(t, 1, s.Len())
}

func TestServer_OnWrongMethodWrongPathRT_OK(t *testing.T) {
	r, err := http.NewRequest(string(GET), "/test/path", nil)
	require.NoError(t, err)

	s := NewRawServer()
	s.On(POST, "/test/path1").Push(mockedRT)
	_, err = s.HTTPClient().Do(r)
	require.Error(t, err)

	assert.Equal(t, 1, s.Len())
}

func TestServer_OnProperMethodWrongPathRT_OK(t *testing.T) {
	r, err := http.NewRequest(string(GET), "/test/path", nil)
	require.NoError(t, err)

	s := NewRawServer()
	s.On(GET, "/test/path1").Push(mockedRT)
	_, err = s.HTTPClient().Do(r)
	require.Error(t, err)

	assert.Equal(t, 1, s.Len())
}

func TestServer_RightOrder(t *testing.T) {
	r, err := http.NewRequest(string(GET), "/test/path", nil)
	require.NoError(t, err)

	s := NewRawServer()
	s.On(GET, "/test/path").Push(mockedRT)
	s.On(GET, "/test/path").Push(mockedRT)

	resp, err := s.HTTPClient().Do(r)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	body, err := ioutil.ReadAll(resp.Body)
	assert.NoError(t, err)
	assert.Equal(t, "hello", string(body))

	resp, err = s.HTTPClient().Do(r)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	body, err = ioutil.ReadAll(resp.Body)
	assert.NoError(t, err)
	assert.Equal(t, "hello", string(body))

	// Last one is err, since the queue is exhausted.
	_, err = s.HTTPClient().Do(r)
	require.Error(t, err)
}

func TestServer_RightOrderWithReset(t *testing.T) {
	r, err := http.NewRequest(string(GET), "/test/path", nil)
	require.NoError(t, err)

	s := NewRawServer()
	s.On(GET, "/test/path").Push(mockedRT)
	s.On(GET, "/test/path").Push(mockedRT)

	assert.Equal(t, 2, s.Len())
	s.Reset()
	s.On(GET, "/test/path").Push(mockedRT)
	s.On(GET, "/test/path").Push(mockedRT)

	assert.Equal(t, 2, s.Len())
	resp, err := s.HTTPClient().Do(r)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	body, err := ioutil.ReadAll(resp.Body)
	assert.NoError(t, err)
	assert.Equal(t, "hello", string(body))

	assert.Equal(t, 1, s.Len())
	resp, err = s.HTTPClient().Do(r)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	body, err = ioutil.ReadAll(resp.Body)
	assert.NoError(t, err)
	assert.Equal(t, "hello", string(body))

	assert.Equal(t, 0, s.Len())
	// Last one is err, since the queue is exhausted.
	_, err = s.HTTPClient().Do(r)
	require.Error(t, err)
}
