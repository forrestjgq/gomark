package gm

import (
	"strconv"
)

func NewWindowMaxerIn(name string, window int) (*Window, error) {
	f := func(v Value) string {
		//glog.Info(">> value: ", v)
		return strconv.Itoa(int(v.x))
	}
	maxLatency, _ := NewMaxerNoExpose()
	maxOp, _ := maxLatency.r.Operators()
	maxLatencyWindow, err := NewWindow("", name, DisplayOnAll, window,
		maxLatency.r.GetWindowSampler(),
		SeriesInSecond,
		maxOp,
		func(left Value, right int) Value {
			return left
		})
	if err != nil {
		return nil, err
	}
	maxLatencyWindow.SetDescriber(f, func(v Value, idx int) string {
		return f(v)
	})
	maxLatencyWindow.SetReceiver(maxLatency)
	maxLatencyWindow.VarBase().AddDisposer(func() {
		maxLatency.Dispose()
	})

	return maxLatencyWindow, nil
}
func NewWindowMaxer(name string) (*Window, error) {
	return NewWindowMaxerIn(name, defaultDumpInterval)
}
