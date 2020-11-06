package gm

import (
	"io"
)

type SeriesFrequency int

const (
	SeriesInWindow SeriesFrequency = iota
	SeriesInSecond
)

type winSampler interface {
	SetWindow(window int)
	SamplesInWindow(window int) []Value
	ValueInWindow(window int) *sampleInRange
}

type Window struct {
	op          Operator
	seriesDivOp OperatorInt
	sampler     winSampler
	window      int
	series      *IntSeries
	frequency   SeriesFrequency
}

func (w *Window) Name() string {
	panic("implement me")
}

func (w *Window) Identity() Identity {
	panic("implement me")
}

func (w *Window) Push(v Mark) {
	panic("implement me")
}

func (w *Window) OnExpose() {
	if w.series == nil && flagSaveSeries {
		w.series = NewIntSeries(w.op, w.seriesDivOp)
	}
}

func (w *Window) OnSample() {
	if w.series != nil {
		if w.frequency == SeriesInSecond {
			w.series.Append(w.ValueOf(1))
		} else {
			w.series.Append(w.Value())
		}
	}
}

func (w *Window) Describe(wr io.Writer, quote bool) {
	panic("implement me")
}

func (w *Window) DescribeSeries(wr io.Writer, opt *SeriesOption) error {
	panic("implement me")
}

func (w *Window) GetSpanOf(window int) *sampleInRange {
	return w.sampler.ValueInWindow(window)
}
func (w *Window) GetSpan() *sampleInRange {
	return w.sampler.ValueInWindow(w.window)
}
func (w *Window) ValueOf(window int) Value {
	v := w.GetSpanOf(window)
	if v != nil {
		return v.value
	}
	return Value{}
}
func (w *Window) Value() Value {
	return w.ValueOf(w.window)
}
func (w *Window) WindowSize() int {
	return w.window
}
func (w *Window) GetSamples() []Value {
	return w.sampler.SamplesInWindow(w.window)
}

func NewWindow(window int, sampler winSampler, freq SeriesFrequency, op Operator, seriesDivOp OperatorInt) *Window {
	if window <= 0 {
		window = defaultDumpInterval
	}
	return &Window{
		op:          op,
		seriesDivOp: seriesDivOp,
		window:      window,
		frequency:   freq,
		sampler:     sampler,
	}
}
