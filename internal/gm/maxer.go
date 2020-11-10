package gm

import (
	"io"
	"strconv"
)

type Maxer struct {
	r *Reducer
}

func (m *Maxer) Push(v Mark) {
	m.r.Push(v)
}

func (m *Maxer) OnExpose() {
	m.r.OnExpose()
}

func (m *Maxer) OnSample() {
	m.r.OnSample()
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

// NewMaxer create an maxer
func NewMaxer() *Maxer {
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
	return maxer
}
