package httpt

import (
	"fmt"
	"net/http"
	"strings"
	"testing"
)

type Method string

const (
	ANY     = Method("")
	CONNECT = Method("CONNECT")
	DELETE  = Method("DELETE")
	GET     = Method("GET")
	HEAD    = Method("HEAD")
	OPTIONS = Method("OPTIONS")
	PATCH   = Method("PATCH")
	POST    = Method("POST")
	PUT     = Method("PUT")
	TRACE   = Method("TRACE")

	AnyPath = "!AnyPath!"
)

type RoundTrip func(*http.Request) (*http.Response, error)

type Server struct {
	*tripBuilder

	DefaultRoundTrip RoundTrip
}

func New() *Server {
	return &Server{
		tripBuilder: newTripBuilder(),
	}
}

func NotMockedRT(t *testing.T) func(*http.Request) (*http.Response, error) {
	return func(r *http.Request) (*http.Response, error) {
		t.Errorf("httpt.Server: RoundTrip not mocked for this request %s:%s",
			r.Method, getPathOnly(r))
		return &http.Response{
			StatusCode: http.StatusInternalServerError,
		}, nil
	}
}

func NewWithT(t *testing.T) *Server {
	return &Server{
		tripBuilder:      newTripBuilder(),
		DefaultRoundTrip: NotMockedRT(t),
	}
}

func (s *Server) HTTPClient() *http.Client {
	return &http.Client{
		Transport: s,
	}
}

type tripEntry struct {
	method Method
	path   string
	trip   RoundTrip
}

type tripQueue struct {
	queue []tripEntry
}

func (q *tripQueue) push(method Method, path string, r RoundTrip) {
	q.queue = append(q.queue, tripEntry{
		method: method,
		path:   path,
		trip:   r,
	})
}

func (q *tripQueue) reset() {
	q.queue = []tripEntry(nil)
}

func (q *tripQueue) pop(method Method, path string) (RoundTrip, bool) {
	for i, e := range q.queue {
		if e.method != method && e.method != ANY {
			continue
		}

		if e.path != path && e.path != AnyPath {
			continue
		}

		q.queue = append(q.queue[:i], q.queue[i+1:]...)
		return e.trip, true
	}

	return nil, false
}

func getPathOnly(req *http.Request) string {
	path := req.URL.String()
	if strings.Contains(path, "?") {
		return strings.Split(path, "?")[0]
	}
	return path
}

func (s *Server) RoundTrip(req *http.Request) (*http.Response, error) {
	method := Method(req.Method)
	path := getPathOnly(req)

	if r, ok := s.engine.pop(method, path); ok {
		return r(req)
	}

	if s.DefaultRoundTrip == nil {
		return nil, fmt.Errorf(
			"httpt.Server request not mocked for this request %s:%s", method, path)
	}
	return s.DefaultRoundTrip(req)
}

func (s *Server) Reset() {
	s.engine.reset()
}

func (s *Server) Len() int {
	return len(s.engine.queue)
}

type tripPusher struct {
	engine *tripQueue
	method Method
	path   string
}

func newTripPusher(engine *tripQueue, method Method, path string) *tripPusher {
	return &tripPusher{
		engine: engine,
		method: method,
		path:   path,
	}
}

func (t *tripPusher) Push(f RoundTrip) {
	t.engine.push(t.method, t.path, f)
}

type tripBuilder struct {
	*tripPusher
}

func newTripBuilder() *tripBuilder {
	return &tripBuilder{
		tripPusher: newTripPusher(&tripQueue{}, ANY, AnyPath),
	}
}

func (t *tripBuilder) On(method Method, path string) *tripPusher {
	return newTripPusher(t.engine, method, path)
}

func ConnectionFailure(err error) func(*http.Request) (*http.Response, error) {
	return func(_ *http.Request) (*http.Response, error) {
		return nil, err
	}
}
