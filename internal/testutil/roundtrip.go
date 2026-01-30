package testutil

import (
	"net/http"
	"net/http/httptest"
)

type HandlerRoundTripper struct {
	Handler http.Handler
}

func (rt HandlerRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	recorder := httptest.NewRecorder()
	rt.Handler.ServeHTTP(recorder, req)
	return recorder.Result(), nil
}

func NewTestClient(handler http.Handler) (*http.Client, string) {
	return &http.Client{Transport: HandlerRoundTripper{Handler: handler}}, "http://lark.test"
}
