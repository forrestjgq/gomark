package gm

import (
	"errors"
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
	ValueInWindow(window int) sampleInRange
}

type Window struct {
	vb          *VarBase
	op          Operator
	seriesDivOp OperatorInt
	sampler     winSampler
	window      int
	series      *IntSeries
	frequency   SeriesFrequency
	serializer  ValueSerializer
	converter   ValueConverter
}

func (w *Window) Dispose() []Identity {
	w.sampler = nil
	w.series = nil
	return nil
}

func (w *Window) VarBase() *VarBase {
	return w.vb
}

func (w *Window) OnExpose(vb *VarBase) error {
	w.vb = vb
	if w.series == nil && flagSaveSeries {
		w.series = NewIntSeries(w.op, w.seriesDivOp)
	}
	return nil
}

func (w *Window) Push(v Mark) {
	panic("implement me")
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

func (w *Window) SetDescriber(serial ValueSerializer, cvt ValueConverter) {
	w.serializer = serial
	w.converter = cvt
}
func (w *Window) Describe(sw io.StringWriter, _ bool) {
	_, _ = sw.WriteString(w.serializer(w.Value()))
}

func (w *Window) DescribeSeries(sw io.StringWriter, opt *SeriesOption) error {
	if w.series == nil {
		return errors.New("no series defined")
	}
	if !opt.TestOnly {
		w.series.Describe(sw, nil, w.converter)
	}
	return nil
}
func (w *Window) GetSpanOf(window int) sampleInRange {
	return w.sampler.ValueInWindow(window)
}
func (w *Window) GetSpan() sampleInRange {
	return w.sampler.ValueInWindow(w.window)
}

func (w *Window) ValueOf(window int) Value {
	return w.GetSpanOf(window).value
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
