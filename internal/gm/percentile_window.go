package gm

import (
	"io"
)

type PercentileWinSampler interface {
	SetWindow(window int)
	SamplesInWindow(window int) []*PercentileSamples
	ValueInWindow(window int) PercentileSampleInRange
}

type PercentileWindow struct {
	vb           *VarBase
	op           PercentileOperator
	seriesDivOp  PercentileOperatorInt
	sampler      PercentileWinSampler
	window       int
	series       *PercentileSeries
	frequency    SeriesFrequency
	removeSample disposer
}

func (w *PercentileWindow) VarBase() *VarBase {
	return w.vb
}
func (w *PercentileWindow) Push(v Mark) {
	panic("implement me")
}
func (w *PercentileWindow) Dispose() []Identity {
	if w.series != nil && w.removeSample != nil {
		w.removeSample()
	}
	w.series = nil
	w.sampler = nil
	w.op = nil
	w.seriesDivOp = nil
	w.window = 0
	return nil
}

func (w *PercentileWindow) OnExpose(vb *VarBase) {
	w.vb = vb
	// todo
	if w.series == nil && flagSaveSeries {
		w.series = NewPercentileSeries(w.op, w.seriesDivOp)
		w.removeSample = AddSampler(w)
	}
	panic("should not be called")
}

func (w *PercentileWindow) takeSample() {
	if w.series != nil {
		if w.frequency == SeriesInSecond {
			w.series.Append(w.ValueOf(1))
		} else {
			w.series.Append(w.Value())
		}
	}
	panic("should not be called")
}

func (w *PercentileWindow) Describe(sw io.StringWriter, _ bool) {
	panic("should not be called")
}

func (w *PercentileWindow) DescribeSeries(sw io.StringWriter, opt *SeriesOption) error {
	panic("should not be called")
	return nil
}
func (w *PercentileWindow) GetSpanOf(window int) PercentileSampleInRange {
	return w.sampler.ValueInWindow(window)
}
func (w *PercentileWindow) GetSpan() PercentileSampleInRange {
	return w.sampler.ValueInWindow(w.window)
}
func (w *PercentileWindow) ValueOf(window int) *PercentileSamples {
	v := w.GetSpanOf(window).value
	if v != nil {
		return v
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
		window = defaultDumpInterval
	}
	sampler.SetWindow(window)
	return &PercentileWindow{
		op:          op,
		seriesDivOp: seriesDivOp,
		window:      window,
		frequency:   freq,
		sampler:     sampler,
	}
}
