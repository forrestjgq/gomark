package gm

import (
	"errors"
	"io"
)

type PassiveCallback func() Value

type PassiveStatus struct {
	VarBase
	op, invOp   Operator
	seriesDivOp OperatorInt
	callback    PassiveCallback
	sampler     *ReducerSampler
	series      *IntSeries
}

func (p *PassiveStatus) Operators() (op Operator, invOp Operator) {
	op, invOp = p.op, p.invOp
	return
}

func (p *PassiveStatus) Name() string {
	return p.name
}

func (p *PassiveStatus) Identity() Identity {
	return p.id
}

func (p *PassiveStatus) Push(v Mark) {
	panic("PassiveStatus should never be pushing with a mark")
}

func (p *PassiveStatus) OnExpose() {
	if p.series == nil && flagSaveSeries {
		p.series = NewIntSeries(p.op, p.seriesDivOp)
	}
}

func (p *PassiveStatus) OnSample() {
	if p.sampler != nil {
		p.sampler.takeSample()
	}
	if p.series != nil {
		p.series.Append(p.GetValue())
	}
}

func (p *PassiveStatus) Describe(w io.Writer, quote bool) {
	panic("implement me")
}

func (p *PassiveStatus) DescribeSeries(w io.Writer, opt *SeriesOption) error {
	if p.series == nil {
		return errors.New("no series defined")
	}
	panic("implement me")
}

func (p *PassiveStatus) GetValue() Value {
	return p.callback()
}

func (p *PassiveStatus) Reset() Value {
	// invOp is not nil, so reset should not be called
	panic("PassiveStatus should not be reset")
}
func (p *PassiveStatus) GetWindowSampler() winSampler {
	if p.sampler == nil {
		p.sampler = NewReducerSampler(p)
	}
	return p.sampler
}
func NewPassiveStatus(callback PassiveCallback) Variable {
	return &PassiveStatus{}
}
