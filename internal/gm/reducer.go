package gm

import (
	"errors"
	"io"
)

type Reducer struct {
	op, invOp   Operator
	seriesDivOp OperatorInt
	value       Value
	sampler     *ReducerSampler
	series      *IntSeries
}

func (r *Reducer) Operators() (op Operator, invOp Operator) {
	op, invOp = r.op, r.invOp
	return
}

func (r *Reducer) Push(v Mark) {
	r.value = r.op(r.value, OneValue(int64(v)))
}

func (r *Reducer) GetValue() Value {
	return r.value
}

func (r *Reducer) Reset() Value {
	v := r.value
	r.value.Reset()
	return v
}
func (r *Reducer) GetWindowSampler() winSampler {
	if r.sampler == nil {
		r.sampler = NewReducerSampler(r)
	}
	return r.sampler
}
func (r *Reducer) Describe(w io.StringWriter, serial func(v Value) string) {
	_, _ = w.WriteString(serial(r.GetValue()))
}
func (r *Reducer) DescribeSeries(w io.StringWriter, opt *SeriesOption, splitName []string, cvt ValueConverter) error {
	// see reducer.h, Reducer::describe_series
	if r.sampler == nil {
		return errors.New("sampler is not created")
	}

	if !opt.testOnly {
		r.series.Describe(w, splitName, cvt)
	}
	return nil
}
func (r *Reducer) OnExpose() {
	if r.series == nil && r.invOp != nil && r.seriesDivOp != nil && flagSaveSeries {
		r.series = NewIntSeries(r.op, r.seriesDivOp)
	}
}
func (r *Reducer) OnSample() {
	if r.sampler != nil {
		r.sampler.takeSample()
	}
	if r.series != nil {
		r.series.Append(r.GetValue())
	}
}

func NewReducer(op, invOp Operator, seriesDivOp OperatorInt) *Reducer {
	r := &Reducer{
		op:          op,
		invOp:       invOp,
		seriesDivOp: seriesDivOp,
		sampler:     nil,
	}
	return r
}
