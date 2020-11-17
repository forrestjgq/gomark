package gm

import (
	"errors"
	"io"
)

type PassiveCallback func() Value

type PassiveStatus struct {
	vb          *VarBase
	receiver    Pushable
	op, invOp   Operator
	seriesDivOp OperatorInt
	log         bool

	callback      PassiveCallback
	sampler       *ReducerSampler
	series        *IntSeries
	names         []string
	serializer    ValueSerializer
	converter     ValueConverter
	seriesDispose disposer
}

func (p *PassiveStatus) Dispose() {
	if p.sampler != nil {
		p.sampler.dispose()
		p.sampler = nil
	}
	if p.seriesDispose != nil {
		p.seriesDispose()
		p.seriesDispose = nil
	}
	p.series = nil
}

func (p *PassiveStatus) VarBase() *VarBase {
	return p.vb
}

func (p *PassiveStatus) Operators() (op Operator, invOp Operator) {
	op, invOp = p.op, p.invOp
	return
}

func (p *PassiveStatus) Push(v Mark) {
	if p.receiver != nil {
		p.receiver.Push(v)
	} else {
		panic("PassiveStatus should never be pushing with a mark")
	}
}

func (p *PassiveStatus) takeSample() {
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
	if !opt.TestOnly {
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
func (p *PassiveStatus) setReceiver(reciever Pushable) {
	p.receiver = reciever
}
func (p *PassiveStatus) setLog(log bool) {
	p.log = log
	if p.series != nil {
		p.series.log = log
	}
}
func NewPassiveStatusNoExpose(callback PassiveCallback, op, invOp Operator) (*PassiveStatus, error) {
	return NewPassiveStatus("", "", DisplayOnNothing, callback, op, invOp, nil)
}
func NewPassiveStatusWithName(name string, callback PassiveCallback, op, invOp Operator, divOp OperatorInt) (*PassiveStatus, error) {
	return NewPassiveStatus("", name, DisplayOnAll, callback, op, invOp, divOp)
}
func NewPassiveStatus(prefix, name string, filter DisplayFilter,
	callback PassiveCallback, op, invOp Operator, divOp OperatorInt) (*PassiveStatus, error) {
	p := &PassiveStatus{
		op:          op,
		invOp:       invOp,
		seriesDivOp: divOp,
		callback:    callback,
	}

	if len(name) > 0 {
		var err error
		if p.vb, err = Expose(prefix, name, filter, p); err != nil {
			return nil, err
		}

		if p.series == nil && flagSaveSeries {
			p.series = NewIntSeries(p.op, p.seriesDivOp)
			p.series.log = p.log
			p.seriesDispose = AddSampler(p)
		}
	}

	p.setLog(false)
	return p, nil
}
