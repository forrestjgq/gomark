package gm

import (
	"errors"
	"io"

	"github.com/golang/glog"
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
	vb           *VarBase
	op           Operator
	log          bool
	seriesDivOp  OperatorInt
	sampler      winSampler
	window       int
	series       *IntSeries
	frequency    SeriesFrequency
	serializer   ValueSerializer
	converter    ValueConverter
	removeSample disposer
	receiver     Pushable
}

func (w *Window) SetReceiver(receiver Pushable) {
	w.receiver = receiver
}
func (w *Window) Dispose() {
	if w.series != nil && w.removeSample != nil {
		w.removeSample()
	}
	w.sampler = nil
	w.series = nil
}

func (w *Window) VarBase() *VarBase {
	return w.vb
}

func (w *Window) Push(v Mark) {
	if w.receiver != nil {
		w.receiver.Push(v)
	}
}

func (w *Window) takeSample() {
	if w.series != nil {
		var v Value
		if w.frequency == SeriesInSecond {
			v = w.ValueOf(1)
		} else {
			v = w.Value()
		}
		if w.log {
			glog.Infof("window take sample value: %v", v)
		}
		w.series.Append(v)
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
		glog.Info("No series")
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

func NewWindowNoExpose(window int, sampler winSampler, op Operator) (*Window, error) {
	return NewWindow("", "", DisplayOnNothing, window, sampler, SeriesInSecond, op, nil)
}
func NewWindowWithName(name string, window int, sampler winSampler, op Operator, seriesDivOp OperatorInt) (*Window, error) {
	return NewWindow(name, "window", DisplayOnAll, window, sampler, SeriesInSecond, op, seriesDivOp)
}
func NewWindow(prefix, name string, filter DisplayFilter, window int,
	sampler winSampler, freq SeriesFrequency, op Operator, seriesDivOp OperatorInt) (*Window, error) {
	if window <= 0 {
		window = defaultDumpInterval
	}
	sampler.SetWindow(window)
	w := &Window{
		op:          op,
		seriesDivOp: seriesDivOp,
		window:      window,
		frequency:   freq,
		sampler:     sampler,
	}
	if len(name) > 0 {
		var err error
		if w.vb, err = Expose(prefix, name, filter, w); err != nil {
			return nil, err
		}

		if w.series == nil && flagSaveSeries {
			w.series = NewIntSeries(w.op, w.seriesDivOp)
			w.removeSample = AddSampler(w)
		}
	}

	return w, nil
}
