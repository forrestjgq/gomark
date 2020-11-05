package gm

import (
	"io"
)

type IntRecorder struct {
	VarBase
	op, invOp Operator
	value     Value // x: sum, y: num
	sampler   *ReducerSampler
}

func (ir *IntRecorder) Operators() (op Operator, invOp Operator) {
	op, invOp = ir.op, ir.invOp
	return
}
func (ir *IntRecorder) Name() string {
	return ir.name
}

func (ir *IntRecorder) Identity() Identity {
	return ir.id
}

func (ir *IntRecorder) Push(v Mark) {
	ir.value.x += int64(v)
	ir.value.y += 1
}

func (ir *IntRecorder) Reset() Value {
	v := ir.value
	ir.value.Reset()
	return v
}

func (ir *IntRecorder) GetValue() Value {
	return ir.value
}

func (ir *IntRecorder) OnExpose() {
	panic("implement me")
}

func (ir *IntRecorder) OnSample() {
	if ir.sampler != nil {
		ir.sampler.takeSample()
	}
}

func (ir *IntRecorder) Describe(w io.Writer, quote bool) {
	panic("implement me")
}

func (ir *IntRecorder) DescribeSeries(w io.Writer, opt *SeriesOption) error {
	panic("implement me")
}

func (ir *IntRecorder) sum() int64 {
	return ir.value.x
}
func (ir *IntRecorder) num() int64 {
	return ir.value.y
}
func (ir *IntRecorder) IntAverage() int64 {
	num := ir.num()
	if num == 0 {
		return 0
	}

	return ir.sum() / num
}
func (ir *IntRecorder) FloatAverage() float64 {
	num := ir.num()
	if num == 0 {
		return 0
	}

	return float64(ir.sum()) / float64(ir.num())
}

func (ir *IntRecorder) GetWindowSampler() winSampler {
	if ir.sampler == nil {
		ir.sampler = NewReducerSampler(ir)
	}
	return ir.sampler
}
func NewIntRecorder() Variable {
	ir := &IntRecorder{
		VarBase: VarBase{},
		op: func(left, right Value) Value {
			return left.Add(&right)
		},
		invOp: func(left, right Value) Value {
			return left.Sub(&right)
		},
		value:   Value{},
		sampler: nil,
	}
	ir.id = AddVariable(ir)
	return ir
}
