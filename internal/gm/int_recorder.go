package gm

import "github.com/golang/glog"

type IntRecorder struct {
	op, invOp Operator
	value     Value // x: sum, y: num
	sampler   *ReducerSampler
}

func (ir *IntRecorder) Operators() (op Operator, invOp Operator) {
	op, invOp = ir.op, ir.invOp
	return
}

func (ir *IntRecorder) Push(v Mark) {
	ir.value.x += int64(v)
	ir.value.y += 1
	glog.Infof("IntRecord value: %v", ir.value)
}

func (ir *IntRecorder) Reset() Value {
	v := ir.value
	ir.value.Reset()
	return v
}

func (ir *IntRecorder) GetValue() Value {
	return ir.value
}

func (ir *IntRecorder) Dispose() {
	if ir.sampler != nil {
		ir.sampler.dispose()
	}
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
func NewIntRecorder() *IntRecorder {
	ir := &IntRecorder{
		op: func(left, right Value) Value {
			return left.Add(&right)
		},
		invOp: func(left, right Value) Value {
			return left.Sub(&right)
		},
		value:   Value{},
		sampler: nil,
	}
	return ir
}
