package gm

import (
	"io"
)

type CDF struct {
	w *PercentileWindow
}

func (C CDF) Name() string {
	panic("implement me")
}

func (C CDF) Identity() Identity {
	panic("implement me")
}

func (C CDF) Push(v Mark) {
	panic("implement me")
}

func (C CDF) OnExpose() {
	panic("implement me")
}

func (C CDF) OnSample() {
	panic("implement me")
}

func (C CDF) Describe(w io.Writer, quote bool) {
	panic("implement me")
}

func (C CDF) DescribeSeries(w io.Writer, opt *SeriesOption) error {
	panic("implement me")
}

func newCDF(w *PercentileWindow) *CDF {
	return &CDF{w: w}
}
