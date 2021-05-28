package gm

import (
	"errors"
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

func (a *Adder) Dispose() {
	a.r.Dispose()
	a.r = nil
}

func (a *Adder) Push(v Mark) {
	a.r.Push(v)
}

func (a *Adder) GetValue() int64 {
	return a.r.GetValue().x
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
	return NewAdder("")
}

func NewAdderPersecond(name string) (*Window, error) {
	if len(name) == 0 {
		return nil, errors.New("adder persecond without a name")
	}
	adder, err := NewAdder(name + "_adder")
	if err != nil {
		return nil, err
	}

	w, err1 := NewWindow(name, "per_second", DisplayOnAll, defaultDumpInterval,
		adder.r.GetWindowSampler(), SeriesInSecond, adder.r.op, adder.r.seriesDivOp)
	if err1 != nil {
		srv.remove(adder.vb.id)
		return nil, err1
	}
	f := func(v Value) string {
		return strconv.Itoa(int(v.x))
	}
	w.SetDescriber(f, func(v Value, idx int) string {
		return f(v)
	})
	w.SetReceiver(adder)
	w.VarBase().AddChild(adder.vb.id)

	return w, nil
}
func NewAdder(name string) (*Adder, error) {
	adder := &Adder{}
	adder.r = NewReducer(
		func(dst, src Value) Value {
			return dst.Add(&src)
		},
		func(dst, src Value) Value {
			return dst.Sub(&src)
		},
		func(left Value, right int) Value {
			var v Value
			if right != 0 {
				v.x = left.x / int64(right)
			}
			return v
		})

	if len(name) > 0 {
		var err error
		adder.vb, err = Expose("", name, DisplayOnAll, adder)
		if err != nil {
			return nil, err
		}
		adder.r.OnExpose()
	}
	return adder, nil
}
