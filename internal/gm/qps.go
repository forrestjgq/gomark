package gm

import (
	"strconv"
)

func NewQPSNoExpose() (*PassiveStatus, error) {
	return NewQPS("", "", DisplayOnNothing, SeriesInSecond)
}
func NewQPSWithName(name string) (*PassiveStatus, error) {
	return NewQPS("", name, DisplayOnAll, SeriesInSecond)
}
func NewQPS(prefix, name string, filter DisplayFilter, freq SeriesFrequency) (*PassiveStatus, error) {
	latency, _ := NewIntRecorderNoExpose()
	op, invOp := latency.Operators()

	window := defaultDumpInterval

	latencyWindow, err := NewWindow(name, "latency", DisplayOnAll, window, latency.GetWindowSampler(), SeriesInSecond, op, nil)
	if err != nil {
		return nil, err
	}
	f := func(v Value) string {
		avg := v.AverageInt()
		if avg != 0 {
			return strconv.Itoa(int(avg))
		}
		return strconv.FormatFloat(v.AverageFloat(), 'f', 3, 64)
	}
	latencyWindow.SetDescriber(f, func(v Value, idx int) string {
		return f(v)
	})

	qps, err1 := NewPassiveStatus(name, "qps", DisplayOnAll, func() Value {
		var v Value
		s := latencyWindow.GetSpanOf(1)
		if s.du <= 0 {
			return v
		}

		// x: qps, y: total count
		v.x = int64(float64(s.value.y) / s.du.Seconds())
		v.y = s.value.y
		return v
	}, op, invOp, statOperatorInt)
	if err1 != nil {
		srv.remove(latencyWindow.vb.ID())
		return nil, err1
	}
	qps.setReceiver(latency)
	qps.vb.AddChild(latencyWindow.vb.ID())
	qps.SetDescriber(XValueSerializer, func(v Value, idx int) string {
		return XValueSerializer(v)
	})

	return qps, nil
}
