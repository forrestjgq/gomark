package gm

import (
	"io"
	"strconv"

	"github.com/forrestjgq/gomark/gmi"
)

type QPS struct {
	vb            *VarBase
	latency       *IntRecorder
	latencyWindow *Window
	qps           *PassiveStatus
}

func (q *QPS) Mark(n int32) {
	if q.vb != nil && q.vb.Valid() {
		s := makeStub(q.vb.ID(), Mark(n))
		PushStub(s)
	}
}

func (q *QPS) Cancel() {
	if q.vb != nil && q.vb.Valid() {
		RemoveVariable(q.vb.ID())
	}
}

func (q *QPS) VarBase() *VarBase {
	return q.vb
}

func (q *QPS) Push(v Mark) {
	q.latency.Push(v)
}

func (q *QPS) OnExpose(vb *VarBase) error {
	q.vb = vb
	if err := Expose(vb.name, "latency", DisplayOnAll, q.latencyWindow); err != nil {
		return err
	}
	if err := Expose(vb.name, "qps", DisplayOnAll, q.qps); err != nil {
		return err
	}
	return nil
}

func (q *QPS) Dispose() []Identity {
	ret := []Identity{q.qps.vb.id, q.latencyWindow.vb.id}
	q.qps = nil
	q.latencyWindow = nil
	return ret
}

func (q *QPS) Describe(w io.StringWriter, quote bool) {
	panic("should not be called")
}

func (q *QPS) DescribeSeries(w io.StringWriter, opt *SeriesOption) error {
	panic("should not be called")
}

func NewQPS(name string) gmi.Marker {
	q := &QPS{
		latency: NewIntRecorder(),
	}
	op, invOp := q.latency.Operators()

	window := defaultDumpInterval

	q.latencyWindow = NewWindow(window, q.latency.GetWindowSampler(), SeriesInSecond, op, nil)
	f := func(v Value) string {
		avg := v.AverageInt()
		if avg != 0 {
			return strconv.Itoa(int(avg))
		}
		return strconv.FormatFloat(v.AverageFloat(), 'f', 3, 64)
	}
	q.latencyWindow.SetDescriber(f, func(v Value, idx int) string {
		return f(v)
	})

	q.qps = NewPassiveStatus(func() Value {
		var v Value
		s := q.latencyWindow.GetSpanOf(1)
		if s.du <= 0 {
			return v
		}

		// x: qps, y: total count
		v.x = int64(float64(s.value.y) / s.du.Seconds())
		v.y = s.value.y
		return v
	}, op, invOp, statOperatorInt)
	q.qps.SetDescriber(XValueSerializer, func(v Value, idx int) string {
		return XValueSerializer(v)
	})

	err := Expose("", name, DisplayOnNothing, q)
	if err != nil {
		return nil
	}
	return q
}
