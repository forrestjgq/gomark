package gm

import (
	"errors"
	"io"
)

type PassiveCallback func() Value

type PassiveStatus struct {
	vb          *VarBase
	op, invOp   Operator
	seriesDivOp OperatorInt

	callback   PassiveCallback
	sampler    *ReducerSampler
	series     *IntSeries
	names      []string
	serializer ValueSerializer
	converter  ValueConverter
}

func (p *PassiveStatus) VarBase() *VarBase {
	return p.vb
}

func (p *PassiveStatus) Operators() (op Operator, invOp Operator) {
	op, invOp = p.op, p.invOp
	return
}

func (p *PassiveStatus) Push(_ Mark) {
	panic("PassiveStatus should never be pushing with a mark")
}

func (p *PassiveStatus) OnExpose(vb *VarBase) error {
	p.vb = vb
	if p.series == nil && flagSaveSeries {
		p.series = NewIntSeries(p.op, p.seriesDivOp)
	}
	return nil
}
func (p *PassiveStatus) OnSample() {
	if p.sampler != nil {
		p.sampler.takeSample()
	}
	if p.series != nil {
		p.series.Append(p.GetValue())
	}
}

func (p *PassiveStatus) SetDescriber(serial ValueSerializer, cvt ValueConverter) {
	p.serializer = serial
	p.converter = cvt
}
func (p *PassiveStatus) Describe(w io.StringWriter, _ bool) {
	_, _ = w.WriteString(p.serializer(p.GetValue()))
}

func (p *PassiveStatus) DescribeSeries(w io.StringWriter, opt *SeriesOption) error {
	if p.series == nil {
		return errors.New("no series defined")
	}
	if !opt.testOnly {
		p.series.Describe(w, p.names, p.converter)
	}
	return nil
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

func (p *PassiveStatus) SetVectorNames(names []string) {
	p.names = names
}
func NewPassiveStatus(callback PassiveCallback, op, invOp Operator, divOp OperatorInt) *PassiveStatus {
	return &PassiveStatus{
		op:          op,
		invOp:       invOp,
		seriesDivOp: divOp,
		callback:    callback,
	}
}
