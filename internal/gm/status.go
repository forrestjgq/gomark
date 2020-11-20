package gm

import (
	"errors"
	"io"
	"strconv"
)

type Status struct {
	vb *VarBase
	v  int64
}

func (a *Status) VarBase() *VarBase {
	return a.vb
}

func (a *Status) Dispose() {
}

func (a *Status) Push(v Mark) {
	a.v = int64(v)
}

func (a *Status) GetValue() int64 {
	return a.v
}

func (a *Status) Describe(w io.StringWriter, _ bool) {
	_, _ = w.WriteString(strconv.Itoa(int(a.v)))
}

func (a *Status) DescribeSeries(_ io.StringWriter, _ *SeriesOption) error {
	return errors.New("no supported")
}

func NewStatus(name string) (*Status, error) {
	st := &Status{}
	if len(name) > 0 {
		var err error
		st.vb, err = Expose(name, "status", DisplayOnAll, st)
		if err != nil {
			return nil, err
		}
	}

	return st, nil
}
