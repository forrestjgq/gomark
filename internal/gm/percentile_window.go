package gm

import (
	"io"
)

type PercentileWinSampler interface {
	SetWindow(window int)
	SamplesInWindow(window int) []*PercentileSamples
	ValueInWindow(window int) *PercentileSampleInRange
}

type PercentileWindow struct {
	op          PercentileOperator
	seriesDivOp PercentileOperatorInt
	sampler     PercentileWinSampler
	window      int
	series      *PercentileSeries
	frequency   SeriesFrequency
}

func (w *PercentileWindow) Name() string {
	panic("implement me")
}

func (w *PercentileWindow) Identity() Identity {
	panic("implement me")
}

func (w *PercentileWindow) Push(v Mark) {
	panic("implement me")
}

func (w *PercentileWindow) OnExpose() {
	// todo
	if w.series == nil && flagSaveSeries {
		w.series = NewPercentileSeries(w.op, w.seriesDivOp)
	}
}

func (w *PercentileWindow) OnSample() {
	if w.series != nil {
		if w.frequency == SeriesInSecond {
			w.series.Append(w.ValueOf(1))
		} else {
			w.series.Append(w.Value())
		}
	}
}

func (w *PercentileWindow) Describe(wr io.Writer, quote bool) {
	panic("implement me")
}

func (w *PercentileWindow) DescribeSeries(wr io.Writer, opt *SeriesOption) error {
	panic("implement me")
}

func (w *PercentileWindow) GetSpanOf(window int) *PercentileSampleInRange {
	return w.sampler.ValueInWindow(window)
}
func (w *PercentileWindow) GetSpan() *PercentileSampleInRange {
	return w.sampler.ValueInWindow(w.window)
}
func (w *PercentileWindow) ValueOf(window int) *PercentileSamples {
	v := w.GetSpanOf(window)
	if v != nil {
		return v.value
	}
	return &PercentileSamples{}
}
func (w *PercentileWindow) Value() *PercentileSamples {
	return w.ValueOf(w.window)
}
func (w *PercentileWindow) WindowSize() int {
	return w.window
}
func (w *PercentileWindow) GetSamples() []*PercentileSamples {
	return w.sampler.SamplesInWindow(w.window)
}

func NewPercentileWindow(window int, sampler PercentileWinSampler, freq SeriesFrequency, op PercentileOperator, seriesDivOp PercentileOperatorInt) *PercentileWindow {
	if window <= 0 {
		panic("window size must > 0")
	}
	return &PercentileWindow{
		op:          op,
		seriesDivOp: seriesDivOp,
		window:      window,
		frequency:   freq,
		sampler:     sampler,
	}
}
