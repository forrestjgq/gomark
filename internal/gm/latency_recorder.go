package gm

import (
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/forrestjgq/gomark/internal/util"
)

type LatencyRecorder struct {
	vb                      *VarBase
	latency                 *IntRecorder
	maxLatency              *Maxer
	latencyPercentile       *Percentile
	latencyWindow           *Window
	maxLatencyWindow        *Window
	count                   *PassiveStatus // x: not used, y: count
	qps                     *PassiveStatus // x: qps, y: num
	latencyPercentileWindow *PercentileWindow
	latencyP1               *PassiveStatus
	latencyP2               *PassiveStatus
	latencyP3               *PassiveStatus
	latencyP999             *PassiveStatus
	latencyP9999            *PassiveStatus
	latencyPercentiles      *PassiveStatus
	latencyCdf              *CDF
}

func (lr *LatencyRecorder) Dispose() {
	lr.latencyPercentile.Dispose()
	lr.latency.Dispose()
	lr.maxLatency.Dispose()
	lr.latencyPercentileWindow.Dispose()

	lr.latency = nil
	lr.maxLatency = nil
	lr.latencyPercentile = nil
	lr.latencyWindow = nil
	lr.count = nil
	lr.qps = nil
	lr.latencyPercentileWindow = nil
	lr.latencyP1 = nil
	lr.latencyP2 = nil
	lr.latencyP3 = nil
	lr.latencyP999 = nil
	lr.latencyP9999 = nil
	lr.latencyPercentiles = nil
	lr.latencyCdf = nil
}

func (lr *LatencyRecorder) VarBase() *VarBase {
	return lr.vb
}

func (lr *LatencyRecorder) Describe(_ io.StringWriter, _ bool) {
	panic("LatencyRecorder should not be described")
}

func (lr *LatencyRecorder) DescribeSeries(_ io.StringWriter, _ *SeriesOption) error {
	panic("LatencyRecorder should not be described")
}

func (lr *LatencyRecorder) WindowSize() int {
	return lr.latencyWindow.WindowSize()
}
func (lr *LatencyRecorder) LatencyIn(window int) int64 {
	v := lr.latencyWindow.ValueOf(window)
	return v.AverageInt()
}
func (lr *LatencyRecorder) Latency() int64 {
	v := lr.latencyWindow.Value()
	return v.AverageInt()
}

func (lr *LatencyRecorder) LatencyPercentiles() []uint32 {
	cb := NewPercentileSamples(1022)
	buckets := lr.latencyPercentileWindow.GetSamples()
	cb.CombineOf(buckets)
	result := []uint32{
		cb.GetNumber(varLatencyP1 / 100.0),
		cb.GetNumber(varLatencyP2 / 100.0),
		cb.GetNumber(varLatencyP3 / 100.0),
		cb.GetNumber(0.999),
	}
	return result
}
func (lr *LatencyRecorder) MaxLatency() int64 {
	return lr.maxLatencyWindow.Value().x
}
func (lr *LatencyRecorder) Count() int64 {
	return lr.latency.GetValue().y
}
func (lr *LatencyRecorder) QpsIn(window int) int64 {
	s := lr.latencyWindow.GetSpanOf(window)
	if s.du <= 0 {
		return 0
	}
	return int64(float64(s.value.y) / s.du.Seconds())
}
func (lr *LatencyRecorder) Qps() int64 {
	return lr.qps.GetValue().x
}
func (lr *LatencyRecorder) LatencyPercentile(ratio float64) int64 {
	cb := NewPercentileSamples(1022)
	buckets := lr.latencyPercentileWindow.GetSamples()
	cb.CombineOf(buckets)
	return int64(cb.GetNumber(ratio))
}
func (lr *LatencyRecorder) LatencyName() string {
	return lr.latencyWindow.VarBase().Name()
}

func (lr *LatencyRecorder) LatencyPercentilesName() string {
	return lr.latencyPercentiles.VarBase().name
}
func (lr *LatencyRecorder) LatencyCDFName() string {
	return lr.latencyCdf.VarBase().name
}
func (lr *LatencyRecorder) MaxLatencyName() string {
	return lr.maxLatencyWindow.VarBase().Name()
}
func (lr *LatencyRecorder) CountName() string {
	return lr.count.VarBase().name
}
func (lr *LatencyRecorder) QpsName() string {
	return lr.qps.VarBase().name
}
func (lr *LatencyRecorder) Push(v Mark) {
	lr.latency.Push(v)
	lr.maxLatency.Push(v)
	lr.latencyPercentile.Push(v)
}

func NewLatencyRecorder(name string) (*LatencyRecorder, error) {
	return NewLatencyRecorderInWindow(name, defaultDumpInterval)
}

func NewLatencyRecorderInWindow(name string, window int) (*LatencyRecorder, error) {
	lr := &LatencyRecorder{}

	name = strings.TrimSuffix(name, "latency")
	name = strings.TrimSuffix(name, "Latency")
	if len(name) == 0 {
		return nil, fmt.Errorf("invalid name %s", name)
	}

	var err error
	lr.vb, err = Expose("", name, DisplayOnNothing, lr)
	if err != nil {
		return nil, err
	}

	em := util.NewErrorMerge()
	defer func() {
		// clean up
		if em.Failed() {
			srv.remove(lr.vb.id)
		}
	}()

	name = lr.vb.name

	////////////////////////////////////////////////////////////////////////////////
	// Average Latency
	lr.latency, err = NewIntRecorderNoExpose()
	if em.Merge(err).Failed() {
		return nil, err
	}
	op, invOp := lr.latency.Operators()

	////////////////////////////////////////////////////////////////////////////////
	// Latency Window
	// Window<IntRecorder> has no effect on series divide
	// detail::DivideOnAddition<::bvar::Stat, Op>::inplace_divide(tmp, op, 60);
	lr.latencyWindow, err = NewWindow(name, "latency", DisplayOnAll,
		window, lr.latency.GetWindowSampler(), SeriesInSecond, op, nil)
	if em.Merge(err).Failed() {
		return nil, err
	}
	lr.vb.AddChild(lr.latencyWindow.VarBase().ID())
	f := func(v Value) string {
		avg := v.AverageInt()
		if avg != 0 {
			return strconv.Itoa(int(avg))
		}
		return strconv.FormatFloat(v.AverageFloat(), 'f', 3, 64)
	}
	lr.latencyWindow.SetDescriber(f, func(v Value, idx int) string {
		return f(v)
	})

	////////////////////////////////////////////////////////////////////////////////
	// Max latency, Not exposed
	lr.maxLatency, err = NewMaxerNoExpose()
	if em.Merge(err).Failed() {
		return nil, err
	}
	maxOp, _ := lr.maxLatency.r.Operators()

	////////////////////////////////////////////////////////////////////////////////
	// Max latency Window
	lr.maxLatencyWindow, err = NewWindow(name, "max_latency", DisplayOnAll,
		window,
		lr.maxLatency.r.GetWindowSampler(),
		SeriesInSecond,
		maxOp,
		func(left Value, right int) Value {
			return left
		})
	if em.Merge(err).Failed() {
		return nil, err
	}
	lr.vb.AddChild(lr.maxLatencyWindow.VarBase().ID())

	maxf := func(v Value) string {
		//glog.Info(">> value: ", v)
		return strconv.Itoa(int(v.x))
	}
	lr.maxLatencyWindow.SetDescriber(maxf, func(v Value, idx int) string {
		return maxf(v)
	})
	//lr.maxLatencyWindow.log = true

	////////////////////////////////////////////////////////////////////////////////
	// Counter
	lr.count, err = NewPassiveStatus(name, "count", DisplayOnAll, func() Value {
		return lr.latency.GetValue() // should use value.y
	}, op, invOp, statOperatorInt)
	if em.Merge(err).Failed() {
		return nil, err
	}
	lr.vb.AddChild(lr.count.VarBase().ID())
	lr.count.SetDescriber(YValueSerializer, func(v Value, idx int) string {
		return YValueSerializer(v)
	})

	////////////////////////////////////////////////////////////////////////////////
	// QPS
	lr.qps, err = NewPassiveStatus(name, "qps", DisplayOnAll, func() Value {
		var v Value
		s := lr.latencyWindow.GetSpanOf(1)
		if s.du <= 0 {
			return v
		}

		// x: qps, y: total count
		v.x = int64(float64(s.value.y) / s.du.Seconds())
		v.y = s.value.y
		return v
	}, op, invOp, statOperatorInt)
	if em.Merge(err).Failed() {
		return nil, err
	}
	lr.vb.AddChild(lr.qps.VarBase().ID())
	lr.qps.SetDescriber(XValueSerializer, func(v Value, idx int) string {
		return XValueSerializer(v)
	})

	////////////////////////////////////////////////////////////////////////////////
	// Latency Percentile, not variable
	lr.latencyPercentile = NewPercentile()
	pOp, _ := lr.latencyPercentile.Operators()

	////////////////////////////////////////////////////////////////////////////////
	// Latency Percentile, not variable
	lr.latencyPercentileWindow, err = NewPercentileWindowNoExpose(window, lr.latencyPercentile.GetWindowSampler(), pOp)
	if em.Merge(err).Failed() {
		return nil, err
	}

	// all latency passive status returns value with same x and y

	////////////////////////////////////////////////////////////////////////////////
	// Latency P1
	lr.latencyP1, err = NewPassiveStatus(name, "latency_"+strconv.Itoa(int(varLatencyP1)), DisplayOnPlainText,
		func() Value {
			var v Value
			v.x = lr.LatencyPercentile(varLatencyP1 / 100.0)
			v.y = v.x
			return v
		}, op, invOp, statOperatorInt)
	if em.Merge(err).Failed() {
		return nil, err
	}
	lr.vb.AddChild(lr.latencyP1.VarBase().ID())
	lr.latencyP1.SetDescriber(XValueSerializer, func(v Value, idx int) string {
		return XValueSerializer(v)
	})

	////////////////////////////////////////////////////////////////////////////////
	// Latency P2
	lr.latencyP2, err = NewPassiveStatus(name, "latency_"+strconv.Itoa(int(varLatencyP2)), DisplayOnPlainText,
		func() Value {
			var v Value
			v.x = lr.LatencyPercentile(varLatencyP2 / 100.0)
			v.y = v.x
			return v
		}, op, invOp, statOperatorInt)
	if em.Merge(err).Failed() {
		return nil, err
	}
	lr.vb.AddChild(lr.latencyP2.VarBase().ID())
	lr.latencyP2.SetDescriber(XValueSerializer, func(v Value, idx int) string {
		return XValueSerializer(v)
	})

	////////////////////////////////////////////////////////////////////////////////
	// Latency P3
	lr.latencyP3, err = NewPassiveStatus(name, "latency_"+strconv.Itoa(int(varLatencyP3)), DisplayOnPlainText,
		func() Value {
			var v Value
			v.x = lr.LatencyPercentile(varLatencyP3 / 100.0)
			v.y = v.x
			return v
		}, op, invOp, statOperatorInt)
	if em.Merge(err).Failed() {
		return nil, err
	}
	lr.vb.AddChild(lr.latencyP3.VarBase().ID())
	lr.latencyP3.SetDescriber(XValueSerializer, func(v Value, idx int) string {
		return XValueSerializer(v)
	})

	////////////////////////////////////////////////////////////////////////////////
	// Latency P999
	lr.latencyP999, err = NewPassiveStatus(name, "latency_999", DisplayOnPlainText,
		func() Value {
			var v Value
			v.x = lr.LatencyPercentile(999.0 / 1000.0)
			v.y = v.x
			return v
		}, op, invOp, statOperatorInt)
	if em.Merge(err).Failed() {
		return nil, err
	}
	lr.vb.AddChild(lr.latencyP999.VarBase().ID())
	lr.latencyP999.SetDescriber(XValueSerializer, func(v Value, idx int) string {
		return XValueSerializer(v)
	})

	////////////////////////////////////////////////////////////////////////////////
	// Latency P9999
	lr.latencyP9999, err = NewPassiveStatus(name, "latency_9999", DisplayOnAll,
		func() Value {
			var v Value
			v.x = lr.LatencyPercentile(9999.0 / 10000.0)
			v.y = v.x
			return v
		}, op, invOp, statOperatorInt)
	if em.Merge(err).Failed() {
		return nil, err
	}
	lr.vb.AddChild(lr.latencyP9999.VarBase().ID())
	lr.latencyP9999.SetDescriber(XValueSerializer, func(v Value, idx int) string {
		return XValueSerializer(v)
	})

	////////////////////////////////////////////////////////////////////////////////
	// Latency CDF
	lr.latencyCdf, err = NewCDF(name, "latency_cdf", DisplayOnHTML, lr.latencyPercentileWindow)
	if em.Merge(err).Failed() {
		return nil, err
	}
	lr.vb.AddChild(lr.latencyCdf.VarBase().ID())

	////////////////////////////////////////////////////////////////////////////////
	// Latency Percentiles
	lr.latencyPercentiles, err = NewPassiveStatus(name, "latency_percentiles", DisplayOnHTML,
		func() Value {
			return CombineToValueU32(lr.LatencyPercentiles())
		}, func(left, right Value) Value {
			var v Value
			for i := 0; i < 4; i++ {
				v.SetU32(i, left.GetU32(i)+right.GetU32(i))
			}
			return v
		}, func(left, right Value) Value {
			var v Value
			for i := 0; i < 4; i++ {
				v.SetU32(i, left.GetU32(i)-right.GetU32(i))
			}
			return v

		}, func(left Value, right int) Value {
			var v Value
			if right > 0 {
				for i := 0; i < 4; i++ {
					v.SetU32(i, left.GetU32(i)/uint32(right))
				}
			}
			return v
		})
	if em.Merge(err).Failed() {
		return nil, err
	}
	lr.vb.AddChild(lr.latencyPercentiles.VarBase().ID())
	lr.latencyPercentiles.SetDescriber(VectorValueSerializer, func(v Value, idx int) string {
		if idx >= 4 {
			panic("invalid idx " + strconv.Itoa(idx))
		}

		return strconv.Itoa(int(v.GetU32(idx)))
	})
	//lr.latencyPercentiles.setLog(true)

	names := []string{
		strconv.Itoa(int(varLatencyP1)) + "%",
		strconv.Itoa(int(varLatencyP2)) + "%",
		strconv.Itoa(int(varLatencyP3)) + "%",
		"99.9%",
	}
	lr.latencyPercentiles.SetVectorNames(names)
	return lr, nil
}
