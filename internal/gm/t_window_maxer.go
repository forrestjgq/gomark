package gm

import (
	"io"
	"strconv"
)

type WindowMaxer struct {
	vb               *VarBase
	maxLatency       *Maxer
	maxLatencyWindow *Window
}

func (wm *WindowMaxer) Mark(n int32) {
	mark := makeStub(wm.VarBase().ID(), Mark(n))
	PushStub(mark)
}

func (wm *WindowMaxer) Cancel() {
	RemoveVariable(wm.VarBase().ID())
}

func (wm *WindowMaxer) VarBase() *VarBase {
	return wm.vb
}

func (wm *WindowMaxer) Push(v Mark) {
	wm.maxLatency.Push(v)
}

func (wm *WindowMaxer) OnExpose(vb *VarBase) error {
	wm.vb = vb
	name := vb.name
	if err := Expose(name, "max", DisplayOnAll, wm.maxLatencyWindow); err != nil {
		return err
	}
	return nil
}

func (wm *WindowMaxer) Dispose() []Identity {
	id := wm.maxLatencyWindow.VarBase().ID()
	wm.maxLatencyWindow = nil
	return []Identity{id}
}

func (wm *WindowMaxer) Describe(w io.StringWriter, quote bool) {
	panic("implement me")
}

func (wm *WindowMaxer) DescribeSeries(w io.StringWriter, opt *SeriesOption) error {
	panic("implement me")
}

func NewWindowMaxerIn(name string, window int) *WindowMaxer {
	wm := &WindowMaxer{}
	f := func(v Value) string {
		//glog.Info(">> value: ", v)
		return strconv.Itoa(int(v.x))
	}
	wm.maxLatency = NewMaxer()
	maxOp, _ := wm.maxLatency.r.Operators()
	wm.maxLatencyWindow = NewWindow(window,
		wm.maxLatency.r.GetWindowSampler(),
		SeriesInSecond,
		maxOp,
		func(left Value, right int) Value {
			return left
		})
	wm.maxLatencyWindow.SetDescriber(f, func(v Value, idx int) string {
		return f(v)
	})

	// this is a variable that does not display
	err := Expose("", name, DisplayOnNothing, wm)
	if err != nil {
		return nil
	}
	return wm
}
func NewWindowMaxer(name string) *WindowMaxer {
	return NewWindowMaxerIn(name, defaultDumpInterval)
}
