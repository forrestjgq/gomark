package gm

import "github.com/forrestjgq/gomark/gmi"

type Adder struct {
	id Identity
}

/********************************************************************************
                       Implementation of gmi.Marker
********************************************************************************/

// Mark a value
func (a *Adder) Mark(n int32) {
	s := makeStub(cmdMark, a.id, Mark(n))
	PushStub(s)
}
func (a *Adder) Cancel() {
	RemoveVariable(a.id)
}

// NewAdder create an adder
func NewAdder(name string) gmi.Marker {
	r := NewReducer(name,
		func(dst, src int64) int64 {
			return dst + src
		},
		func(dst, src int64) int64 {
			return dst - src
		})
	r.id = AddVariable(r)

	adder := &Adder{id: r.id}
	return adder
}
