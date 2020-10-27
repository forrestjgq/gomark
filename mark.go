package gomark

import "github.com/forrestjgq/gomark/internal/gm"

type Marker interface {
	Mark(n int32)
	Cancel()
}

func NewAdder(name string) Marker {
	return gm.NewAdder(name)
}
