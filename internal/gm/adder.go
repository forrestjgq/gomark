package gm

import (
	"io"

	"github.com/forrestjgq/gomark/gmi"
)

type Adder struct {
	vb *VarBase
	r  *Reducer
}

func (a *Adder) VarBase() *VarBase {
	return a.vb
}

func (a *Adder) Expose(prefix, name string, displayFilter DisplayFilter) error {
	var err error
	a.vb, err = AddVariable("", name, DisplayOnAll, a)
	return err
}

func (a *Adder) Dispose() {
	panic("implement me")
}

func (a *Adder) Push(v Mark) {
	a.r.Push(v)
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
	if a.vb != nil && a.vb.Valid() {
		s := makeStub(cmdMark, a.vb.ID(), Mark(n))
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
			return Value{
				x: left.x / int64(right),
				y: 0,
			}
		})

	adder := &Adder{
		r: r,
	}

	defer Lock().Unlock()
	err := adder.Expose("", name, DisplayOnAll)
	if err != nil {
		return nil, err
	}
	return adder, nil
}
