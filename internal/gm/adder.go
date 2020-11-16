package gm

import (
	"io"
	"strconv"
)

type Adder struct {
	vb *VarBase
	r  *Reducer
}

func (a *Adder) VarBase() *VarBase {
	return a.vb
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

func NewAdderNoExpose() (*Adder, error) {
	return NewAdder("", "", DisplayOnNothing)
}
func NewAdderWithName(name string) (*Adder, error) {
	return NewAdder("", name, DisplayOnAll)
}

// NewAdder create an adder
func NewAdder(prefix, name string, filter DisplayFilter) (*Adder, error) {
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

	if len(name) > 0 {
		var err error
		adder.vb, err = Expose("", name, filter, adder)
		if err != nil {
			return nil, err
		}
		adder.r.OnExpose()
	}
	return adder, nil
}
