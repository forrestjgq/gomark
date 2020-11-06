package gm

import (
	"io"
)

type Maxer struct {
	VarBase
	r *Reducer
}

func (m *Maxer) Name() string {
	return m.name
}

func (m *Maxer) Identity() Identity {
	return m.id
}

func (m *Maxer) Push(v Mark) {
	panic("implement me")
}

func (m *Maxer) OnExpose() {
	m.r.OnExpose()
}

func (m *Maxer) OnSample() {
	m.r.OnSample()
}

func (m *Maxer) Describe(w io.Writer, quote bool) {
	panic("implement me")
}

func (m *Maxer) DescribeSeries(w io.Writer, opt *SeriesOption) error {
	panic("implement me")
}

// NewMaxer create an maxer
func NewMaxer(name string) *Maxer {
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
		VarBase: VarBase{
			name: name,
			id:   0,
		},
		r: r,
	}
	if len(name) != 0 {
		maxer.id = AddVariable(maxer)
	}
	return maxer
}
