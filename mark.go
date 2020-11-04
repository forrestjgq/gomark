package gomark

import (
	"github.com/forrestjgq/gomark/gmi"
	"github.com/forrestjgq/gomark/internal/gm"
)

func NewAdder(name string) gmi.Marker {
	return gm.NewAdder(name)
}
