// Package gmi defines Marker to represent each variable by interface.
package gmi

// Marker is an interface to provide variable marking.
type Marker interface {
	// Mark a number, the number definition is bound to marker itself.
	Mark(n int32)
	// Stop this marking.
	Cancel()
}
type Route string

const (
	RouteVars  Route = "vars"
	RouteDebug Route = "debug"
	RouteJs    Route = "js"
)

type Request struct {
	Router  Route
	Headers map[string]string
	Params  map[string]string
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

type Response struct {
	Status  int
	Headers map[string]string
	Body    []byte
}
