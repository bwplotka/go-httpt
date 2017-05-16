# go-httpt

[![Build Status](https://travis-ci.org/Bplotka/go-httpt.svg?branch=master)](https://travis-ci.org/Bplotka/go-httpt) [![Go Report Card](https://goreportcard.com/badge/github.com/Bplotka/go-httpt)](https://goreportcard.com/report/github.com/Bplotka/go-httpt)

Standard httptest package is great, but there is lot of boilerplate needed to mock your single HTTP response.

*httpt* is a small library that provides quick request-response mock for your Golang tests!

## Usage

*httpt* provides `Server` struct. Using that you can specify any round trip scenario (on what request what round trip you want).

For example inside unit-test:
```go
package somepackage

import (
    "net/http"
    "testing"
    
    "github.com/Bplotka/go-httpt"
    "github.com/Bplotka/go-httpt/rt"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)


func TestSomething(t *testing.T) {
    request, err := http.NewRequest(...)
    // handle err...
            
    s := httpt.NewServer(t)
    s.On(httpt.GET, "/test/path").Push(rt.StringResponseFunc(http.StatusBadRequest, "really_bad_request"))
    s.On(httpt.POST, httpt.AnyPath).Push(rt.JSONResponseFunc(http.StatusOK, []byte(`{"error": "really_bad_request"}`)))

    testClient := s.HTTPClient()
    // Pass testClient to your components for mocked HTTP calls...
}
```

Having the scenario we can pass mocked HTTP client anywhere. It is common for complex libraries to not use interface but 
rather enabling custom clients from context.Context e.g standard oauth2 package. In this case, we can just pass our httpt.Server's client:

```go
ctx = context.WithValue(ctx, oauth2.HTTPClient, s.HTTPClient())

// Pass ctx into oauth component...
```

Using that pattern is really convenient. Imagine now using such libraries (that take custom clients and make lots of HTTP requests internally) within
HTTP Handler itself. To properly test your Server with HTTP handlers in your unit test you can use this approach:

```go
package somepackage

import (
    "context"
    "net/http"
    "net/http/httptest"
    "testing"
    
    "github.com/Bplotka/go-httpt"
    "github.com/Bplotka/go-httpt/rt"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
    "golang.org/x/oauth2"
)


func TestYourServer(t *testing.T) {
    request, err := http.NewRequest(...)
    // handle err...
    
    s := httpt.NewServer(t)
    s.On(httpt.GET, "/test/path").Push(rt.StringResponseFunc(http.StatusBadRequest, "really_bad_request"))
    s.On(httpt.POST, httpt.AnyPath).Push(rt.JSONResponseFunc(http.StatusOK, []byte(`{"error": "really_bad_request"}`)))

    rec := httptest.NewRecorder()
    yourServer.ServeHTTP(
        rec,
        // Pass test HTTP client.
        request.WithContext(
            context.WithValue(context.TODO(), oauth2.HTTPClient, s.HTTPClient()),
        ),
    )
    // ...
}
```