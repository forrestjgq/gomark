package gm

import (
	"io"

	"github.com/forrestjgq/gomark/gmi"
)

type Adder struct {
	VarBase
	r *Reducer
}

func (a *Adder) Name() string {
	return a.name
}

func (a *Adder) Identity() Identity {
	return a.id
}

func (a *Adder) Push(v Mark) {
	panic("implement me")
}

func (a *Adder) OnExpose() {
	panic("implement me")
}

func (a *Adder) OnSample() {
	panic("implement me")
}

func (a *Adder) Describe(w io.Writer, quote bool) {
	panic("implement me")
}

func (a *Adder) DescribeSeries(w io.Writer, opt *SeriesOption) error {
	panic("implement me")
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
	r := NewReducer(
		func(dst, src Value) Value {
			return dst.Add(&src)
		},
		func(dst, src Value) Value {
			return dst.Sub(&src)
		},
		func(left Value, right int) Value {
			return Value{
				x: left.x / int64(right),
				y: 0,
			}
		})

	adder := &Adder{
		VarBase: VarBase{
			name: name,
			id:   0,
		},
		r: r,
	}
	adder.id = AddVariable(adder)
	return adder
}
