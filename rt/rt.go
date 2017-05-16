package rt

import (
	"bytes"
	"io/ioutil"
	"net/http"
)

// StringResponseFunc is a round trip function that for request returns code and string.
func StringResponseFunc(code int, msg string) func(*http.Request) (*http.Response, error) {
	return func(_ *http.Request) (*http.Response, error) {
		body := ioutil.NopCloser(bytes.NewBufferString(msg))
		return &http.Response{
			StatusCode: code,
			Body:       body,
		}, nil
	}
}

// JSONResponseFunc is a round trip function that for request returns code and particular bytes with the JSON content header.
func JSONResponseFunc(code int, json []byte) func(*http.Request) (*http.Response, error) {
	return func(_ *http.Request) (*http.Response, error) {
		body := ioutil.NopCloser(bytes.NewBuffer(json))
		r := &http.Response{
			StatusCode: code,
			Body:       body,
			Header:     make(http.Header),
		}

		r.Header.Set("Content-Type", "application/json")
		return r, nil
	}
}
