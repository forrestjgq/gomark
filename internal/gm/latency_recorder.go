package gm

import (
	"fmt"
	"io"
	"strconv"
	"strings"
	"unsafe"
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

func (lr *LatencyRecorder) VarBase() *VarBase {
	return lr.vb
}

func (lr *LatencyRecorder) OnExpose(vb *VarBase) error {
	lr.vb = vb

	var err error
	name := vb.name
	if err = Expose(name, "latency", DisplayOnAll, lr.latencyWindow); err != nil {
		return err
	}
	if err = Expose(name, "max_latency", DisplayOnAll, lr.maxLatencyWindow); err != nil {
		return err
	}
	if err = Expose(name, "count", DisplayOnAll, lr.count); err != nil {
		return err
	}
	if err = Expose(name, "qps", DisplayOnAll, lr.qps); err != nil {
		return err
	}
	if err = Expose(name, "latency_"+strconv.Itoa(int(varLatencyP1)), DisplayOnPlainText, lr.latencyP1); err != nil {
		return err
	}
	if err = Expose(name, "latency_"+strconv.Itoa(int(varLatencyP2)), DisplayOnPlainText, lr.latencyP2); err != nil {
		return err
	}
	if err = Expose(name, "latency_"+strconv.Itoa(int(varLatencyP3)), DisplayOnPlainText, lr.latencyP3); err != nil {
		return err
	}
	if err = Expose(name, "latency_999", DisplayOnPlainText, lr.latencyP999); err != nil {
		return err
	}
	if err = Expose(name, "latency_9999", DisplayOnAll, lr.latencyP9999); err != nil {
		return err
	}
	if err = Expose(name, "latency_cdf", DisplayOnHTML, lr.latencyCdf); err != nil {
		return err
	}
	if err = Expose(name, "latency_percentiles", DisplayOnHTML, lr.latencyPercentiles); err != nil {
		return err
	}
	names := []string{
		strconv.Itoa(int(varLatencyP1)) + "%",
		strconv.Itoa(int(varLatencyP2)) + "%",
		strconv.Itoa(int(varLatencyP3)) + "%",
		"99.9%",
	}
	lr.latencyPercentiles.SetVectorNames(names)
	return nil
}

func (lr *LatencyRecorder) OnSample() {
}

func (lr *LatencyRecorder) Describe(w io.StringWriter, quote bool) {
	panic("LatencyRecorder should not be described")
}

func (lr *LatencyRecorder) DescribeSeries(w io.StringWriter, opt *SeriesOption) error {
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
	return lr.latencyWindow.Name()
}

func (lr *LatencyRecorder) LatencyPercentilesName() string {
	return lr.latencyPercentiles.VarBase().name
}
func (lr *LatencyRecorder) LatencyCDFName() string {
	return lr.latencyCdf.VarBase().name
}
func (lr *LatencyRecorder) MaxLatencyName() string {
	return lr.maxLatencyWindow.Name()
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

var statOperatorInt OperatorInt = func(left Value, right int) Value {
	if right == 0 {
		return left
	}
	left.x /= int64(right)
	left.y /= int64(right)
	return left
}

func NewLatencyRecorderInWindow(name string, window int) (*LatencyRecorder, error) {
	lr := &LatencyRecorder{}

	name = strings.TrimSuffix(name, "latency")
	name = strings.TrimSuffix(name, "Latency")
	if len(name) == 0 {
		return nil, fmt.Errorf("invalid name %s", name)
	}

	// if len(prefix) > 0 {
	// 	name = prefix + "_" + name
	// }

	lr.latency = NewIntRecorder()
	op, invOp := lr.latency.Operators()
	// Window<IntRecorder> has no effect on series divide
	// detail::DivideOnAddition<::bvar::Stat, Op>::inplace_divide(tmp, op, 60);
	lr.latencyWindow = NewWindow(window, lr.latency.GetWindowSampler(), SeriesInSecond, op, nil)
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

	lr.maxLatency = NewMaxer()
	maxOp, _ := lr.maxLatency.r.Operators()
	lr.maxLatencyWindow = NewWindow(window, lr.maxLatency.r.GetWindowSampler(), SeriesInSecond, maxOp, statOperatorInt)

	lr.count = NewPassiveStatus(func() Value {
		return lr.latency.GetValue() // should use value.y
	}, op, invOp, statOperatorInt)
	lr.count.SetDescriber(YValueSerializer, func(v Value, idx int) string {
		return YValueSerializer(v)
	})

	lr.qps = NewPassiveStatus(func() Value {
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
	lr.count.SetDescriber(XValueSerializer, func(v Value, idx int) string {
		return XValueSerializer(v)
	})

	lr.latencyPercentile = NewPercentile()
	pOp, _ := lr.latencyPercentile.Operators()
	lr.latencyPercentileWindow = NewPercentileWindow(window,
		lr.latencyPercentile.GetWindowSampler(),
		SeriesInSecond,
		pOp, nil)

	// all latency passive status returns value with same x and y
	lr.latencyP1 = NewPassiveStatus(func() Value {
		var v Value
		v.x = lr.LatencyPercentile(varLatencyP1 / 100.0)
		v.y = v.x
		return v
	}, op, invOp, statOperatorInt)
	lr.latencyP1.SetDescriber(XValueSerializer, func(v Value, idx int) string {
		return XValueSerializer(v)
	})

	lr.latencyP2 = NewPassiveStatus(func() Value {
		var v Value
		v.x = lr.LatencyPercentile(varLatencyP2 / 100.0)
		v.y = v.x
		return v
	}, op, invOp, statOperatorInt)
	lr.latencyP2.SetDescriber(XValueSerializer, func(v Value, idx int) string {
		return XValueSerializer(v)
	})

	lr.latencyP3 = NewPassiveStatus(func() Value {
		var v Value
		v.x = lr.LatencyPercentile(varLatencyP3 / 100.0)
		v.y = v.x
		return v
	}, op, invOp, statOperatorInt)
	lr.latencyP3.SetDescriber(XValueSerializer, func(v Value, idx int) string {
		return XValueSerializer(v)
	})

	lr.latencyP999 = NewPassiveStatus(func() Value {
		var v Value
		v.x = lr.LatencyPercentile(999.0 / 1000.0)
		v.y = v.x
		return v
	}, op, invOp, statOperatorInt)
	lr.latencyP999.SetDescriber(XValueSerializer, func(v Value, idx int) string {
		return XValueSerializer(v)
	})

	lr.latencyP9999 = NewPassiveStatus(func() Value {
		var v Value
		v.x = lr.LatencyPercentile(9999.0 / 10000.0)
		v.y = v.x
		return v
	}, op, invOp, statOperatorInt)
	lr.latencyP9999.SetDescriber(XValueSerializer, func(v Value, idx int) string {
		return XValueSerializer(v)
	})

	lr.latencyCdf = newCDF(lr.latencyPercentileWindow)
	lr.latencyPercentiles = NewPassiveStatus(func() Value {
		return CombineToValueU32(lr.LatencyPercentiles())
	}, op, invOp, statOperatorInt)
	lr.latencyPercentiles.SetDescriber(VectorValueSerializer, func(v Value, idx int) string {
		if idx >= 4 {
			panic("invalid idx " + strconv.Itoa(idx))
		}
		p1 := (*uint32)(unsafe.Pointer(uintptr(unsafe.Pointer(&v.x)) + unsafe.Sizeof(uint32(0))*uintptr(idx)))
		val := int(*p1)
		return strconv.Itoa(val)
	})

	// this is a variable that does not display
	err := AddVariable("", name, DisplayOnNothing, lr)
	if err != nil {
		return nil, err
	}
	return lr, nil
}
