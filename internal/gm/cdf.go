package gm

import (
	"errors"
	"fmt"
	"io"
)

type CDF struct {
	vb *VarBase
	w  *PercentileWindow
}

func (c *CDF) Dispose() []Identity {
	c.w = nil
	return nil
}

func (c *CDF) VarBase() *VarBase {
	return c.vb
}

func (c *CDF) Push(_ Mark) {
	panic("CDF push should not be called")
}

func (c *CDF) OnSample() {
}

func (c *CDF) Describe(w io.StringWriter, _ bool) {
	_, _ = w.WriteString("\"click to view\"")
}

func (c *CDF) DescribeSeries(w io.StringWriter, opt *SeriesOption) error {
	if c.w == nil {
		return errors.New("CDF does not take a window")
	}
	if opt.testOnly {
		return nil
	}

	samples := NewPercentileSamples(1022)
	buckets := c.w.GetSamples()
	samples.CombineOf(buckets)

	type pair struct {
		first  int
		second uint32
	}
	values := make([]pair, 20)
	n := 0
	for i := 1; i < 10; i++ {
		values[n].first = i * 10
		values[n+1].second = samples.GetNumber(float64(i) * 0.1)
		n += 2
	}
	for i := 91; i < 100; i++ {
		values[n].first = i * 10
		values[n+1].second = samples.GetNumber(float64(i) * 0.01)
		n += 2
	}
	values[n].first = 100
	values[n+1].second = samples.GetNumber(0.999)
	n += 2
	values[n].first = 101
	values[n+1].second = samples.GetNumber(0.9999)
	n += 2

	_, _ = w.WriteString("{\"label\":\"cdf\",\"data\":[")
	for i := 0; i < n; i++ {
		if i > 0 {
			_, _ = w.WriteString(",")
		}
		_, _ = w.WriteString(fmt.Sprintf("[%d,%d]", values[i].first, values[i].second))
	}
	_, _ = w.WriteString("]}")
	return nil
}

func (c *CDF) OnExpose(vb *VarBase) error {
	c.vb = vb
	return nil
}

func newCDF(w *PercentileWindow) *CDF {
	return &CDF{w: w}
}
