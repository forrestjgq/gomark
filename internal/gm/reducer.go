package gm

import (
	"errors"
	"io"
)

type Reducer struct {
	op, invOp Operator
	value     int64
	name      string
	id        Identity
	sampler   *ReducerSampler
	series    *IntSeries
}

func (r *Reducer) Operators() (op Operator, invOp Operator) {
	return op, invOp
}

func (r *Reducer) Name() string {
	return r.name
}

func (r *Reducer) Identity() Identity {
	return r.id
}

func (r *Reducer) Push(v Mark) {
	r.value = r.op(r.value, int64(v))
}

func (r *Reducer) Value() int64 {
	return r.value
}

func (r *Reducer) Reset() int64 {
	v := r.value
	r.value = 0
	return v
}
func (r *Reducer) GetWindowSampler() winSampler {
	if r.sampler == nil {
		r.sampler = NewReducerSampler(r)
	}
	return r.sampler
}
func (r *Reducer) Describe(w io.Writer, quote bool) {
	// see reducer.h, Reducer::describe
	panic("not implemented")
}
func (r *Reducer) DescribeSeries(w io.Writer, opt *SeriesOption) error {
	// see reducer.h, Reducer::describe_series
	if r.sampler == nil {
		return errors.New("sampler is not created")
	}

	panic("not implemented")
	return nil
}
func (r *Reducer) OnExpose() {
	if r.series == nil && r.invOp != nil && flagSaveSeries {
		r.series = NewIntSeries(r.op)
	}
}
func (r *Reducer) OnSample() {
	if r.sampler != nil {
		r.sampler.takeSample()
	}
	if r.series != nil {
		r.series.Append(r.Value())
	}
}

func NewReducer(name string, op, invOp Operator) *Reducer {
	r := &Reducer{
		op:      op,
		invOp:   invOp,
		value:   0,
		name:    name,
		id:      0,
		sampler: nil,
	}
	return r
}
