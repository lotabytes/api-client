package provider

import (
	"net/http"
)

type HttpRequester interface {
	Do(req *http.Request) (*http.Response, error)
}

type HttpGetterFunc func(req *http.Request) (*http.Response, error)

func (f HttpGetterFunc) Do(req *http.Request) (*http.Response, error) {
	return f(req)
}
