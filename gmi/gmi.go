// Package gmi defines Marker to represent each variable by interface.
package gmi

import "net/textproto"

// Marker is an interface to provide variable marking.
type Marker interface {
	// Mark a number, the number definition is bound to marker itself.
	Mark(n int32)
	// Stop this marking.
	Cancel()
}
type Route string

const (
	RouteVars    Route = "vars"
	RouteDebug   Route = "debug"
	RouteJs      Route = "js"
	RouteMetrics Route = "metrics"
)

type Request struct {
	Router  Route
	Params  map[string]string
	headers map[string]string
}

func (r *Request) GetParam(key string) string {
	if r.Params == nil {
		return ""
	}
	if v, ok := r.Params[key]; ok {
		return v
	}
	return ""
}
func (r *Request) HasParam(key string) bool {
	if r.Params == nil {
		return false
	}
	_, ok := r.Params[key]
	return ok
}
func (r *Request) GetHeader(key string) string {
	if r.headers == nil {
		return ""
	}
	if v, ok := r.headers[textproto.CanonicalMIMEHeaderKey(key)]; ok {
		return v
	}
	return ""
}
func (r *Request) SetHeader(key, value string) {
	if r.headers == nil {
		r.headers = make(map[string]string)
	}
	r.headers[textproto.CanonicalMIMEHeaderKey(key)] = value
}

type Response struct {
	Status  int
	headers map[string]string
	Body    []byte
}

func (r *Response) SetHeader(key, value string) {
	if r.headers == nil {
		r.headers = make(map[string]string)
	}
	r.headers[textproto.CanonicalMIMEHeaderKey(key)] = value
}
func (r *Response) GetHeaders() map[string]string {
	if r.headers == nil {
		r.headers = make(map[string]string)
	}
	return r.headers
}
func (r *Response) GetHeader(key string) string {
	if r.headers == nil {
		return ""
	}
	if v, ok := r.headers[textproto.CanonicalMIMEHeaderKey(key)]; ok {
		return v
	}
	return ""
}
