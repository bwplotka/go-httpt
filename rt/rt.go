package rt

import (
	"bytes"
	"io/ioutil"
	"net/http"
)

func StringResponse(code int, msg string) func(*http.Request) (*http.Response, error) {
	return func(_ *http.Request) (*http.Response, error) {
		body := ioutil.NopCloser(bytes.NewBufferString(msg))
		return &http.Response{
			StatusCode: code,
			Body:       body,
		}, nil
	}
}

func JSONResponse(code int, json []byte) func(*http.Request) (*http.Response, error) {
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
