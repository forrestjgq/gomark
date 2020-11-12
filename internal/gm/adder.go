package gm

import (
	"io"
	"strconv"

	"github.com/forrestjgq/gomark/gmi"
)

type Adder struct {
	vb *VarBase
	r  *Reducer
}

func (a *Adder) VarBase() *VarBase {
	return a.vb
}

func (a *Adder) OnExpose(vb *VarBase) error {
	a.vb = vb
	a.r.OnExpose()
	return nil
}

func (a *Adder) Dispose() []Identity {
	a.r.Dispose()
	a.r = nil
	return nil
}

func (a *Adder) Push(v Mark) {
	a.r.Push(v)
}

func (a *Adder) Describe(w io.StringWriter, _ bool) {
	a.r.Describe(w, func(v Value) string {
		return strconv.Itoa(int(v.x))
	})
}

func (a *Adder) DescribeSeries(w io.StringWriter, opt *SeriesOption) error {
	return a.r.DescribeSeries(w, opt, nil, func(v Value, idx int) string {
		return strconv.Itoa(int(v.x))
	})
}

/********************************************************************************
                       Implementation of gmi.Marker
********************************************************************************/

// Mark a value
func (a *Adder) Mark(n int32) {
	if a.vb != nil && a.vb.Valid() {
		s := makeStub(a.vb.ID(), Mark(n))
		PushStub(s)
	}
}
func (a *Adder) Cancel() {
	if a.vb != nil && a.vb.Valid() {
		RemoveVariable(a.vb.ID())
	}
}

// NewAdder create an adder
func NewAdder(name string) (gmi.Marker, error) {
	r := NewReducer(
		func(dst, src Value) Value {
			return dst.Add(&src)
		},
		func(dst, src Value) Value {
			return dst.Sub(&src)
		},
		func(left Value, right int) Value {
			var v Value
			if right == 0 {
				return v
			}
			v.x = left.x / int64(right)
			return v
		})

	adder := &Adder{
		r: r,
	}

	err := Expose("", name, DisplayOnAll, adder)
	if err != nil {
		return nil, err
	}
	return adder, nil
}
