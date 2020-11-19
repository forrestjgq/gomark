package gm

import (
	"io"
	"strconv"
)

type Maxer struct {
	vb *VarBase
	r  *Reducer
}

func (m *Maxer) VarBase() *VarBase {
	return m.vb
}

func (m *Maxer) Dispose() {
	m.r.Dispose()
	m.r = nil
}

func (m *Maxer) Push(v Mark) {
	m.r.Push(v)
}

func (m *Maxer) Describe(w io.StringWriter, quote bool) {
	m.r.Describe(w, func(v Value) string {
		return strconv.Itoa(int(v.x))
	})
}

func (m *Maxer) DescribeSeries(w io.StringWriter, opt *SeriesOption) error {
	return m.r.DescribeSeries(w, opt, nil, func(v Value, idx int) string {
		return strconv.Itoa(int(v.x))
	})
}

func NewMaxerNoExpose() (*Maxer, error) {
	return NewMaxer("")
}

// NewMaxer create an maxer
func NewMaxer(name string) (*Maxer, error) {
	r := NewReducer(
		func(dst, src Value) Value {
			if dst.x >= src.x {
				return dst
			}
			return src
		},
		nil,
		nil) // reducer do not create series sampler if invOp is nil

	maxer := &Maxer{
		r: r,
	}

	if len(name) > 0 {
		var err error
		maxer.vb, err = Expose(name, "max", DisplayOnAll, maxer)
		if err != nil {
			return nil, err
		}
		maxer.r.OnExpose()
	}
	return maxer, nil
}
