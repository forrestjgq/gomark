package gm

import (
	"io"

	"github.com/forrestjgq/gomark/gmi"
)

type Counter struct {
	vb      *VarBase
	latency *IntRecorder
	count   *PassiveStatus
}

func (c *Counter) Mark(n int32) {
	if c.vb != nil && c.vb.Valid() {
		s := makeStub(c.vb.ID(), Mark(n))
		PushStub(s)
	}
}

func (c *Counter) Cancel() {
	if c.vb != nil && c.vb.Valid() {
		RemoveVariable(c.vb.ID())
	}
}

func (c *Counter) VarBase() *VarBase {
	return c.vb
}

func (c *Counter) Push(v Mark) {
	c.latency.Push(v)
}

func (c *Counter) OnExpose(vb *VarBase) error {
	c.vb = vb
	if err := Expose(vb.name, "count", DisplayOnAll, c.count); err != nil {
		return err
	}
	return nil
}

func (c *Counter) Dispose() []Identity {
	count := c.count
	c.count = nil
	return []Identity{count.VarBase().ID()}
}

func (c *Counter) Describe(w io.StringWriter, quote bool) {
	c.count.Describe(w, quote)
}

func (c *Counter) DescribeSeries(w io.StringWriter, opt *SeriesOption) error {
	return c.count.DescribeSeries(w, opt)
}

func NewCounter(name string) gmi.Marker {
	c := &Counter{
		latency: NewIntRecorder(),
	}
	op, invOp := c.latency.Operators()

	c.count = NewPassiveStatus(func() Value {
		return c.latency.GetValue() // should use value.y
	}, op, invOp, statOperatorInt)
	c.count.SetDescriber(YValueSerializer, func(v Value, idx int) string {
		return YValueSerializer(v)
	})

	err := Expose("", name, DisplayOnNothing, c)
	if err != nil {
		return nil
	}
	return c
}
