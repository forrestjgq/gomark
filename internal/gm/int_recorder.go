package gm

import "io"

type IntRecorder struct {
	vb        *VarBase
	op, invOp Operator
	value     Value // x: sum, y: num
	sampler   *ReducerSampler
}

func (ir *IntRecorder) VarBase() *VarBase {
	return ir.vb
}

func (ir *IntRecorder) Dispose() {
	if ir.sampler != nil {
		ir.sampler.dispose()
	}
}

func (ir *IntRecorder) Describe(w io.StringWriter, quote bool) {
	panic("implement me")
}

func (ir *IntRecorder) DescribeSeries(w io.StringWriter, opt *SeriesOption) error {
	panic("implement me")
}

func (ir *IntRecorder) Operators() (op Operator, invOp Operator) {
	op, invOp = ir.op, ir.invOp
	return
}

func (ir *IntRecorder) Push(v Mark) {
	ir.value.x += int64(v)
	ir.value.y += 1
	//glog.Infof("IntRecord value: %v", ir.value)
}

func (ir *IntRecorder) Reset() Value {
	v := ir.value
	ir.value.Reset()
	return v
}

func (ir *IntRecorder) GetValue() Value {
	return ir.value
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
func NewIntRecorderNoExpose() (*IntRecorder, error) {
	return NewIntRecorder("", "", DisplayOnNothing)
}
func NewIntRecorderWithName(name string) (*IntRecorder, error) {
	return NewIntRecorder("", name, DisplayOnAll)
}
func NewIntRecorder(prefix, name string, filter DisplayFilter) (*IntRecorder, error) {
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
	if len(name) > 0 {
		var err error
		ir.vb, err = Expose(prefix, name, filter, ir)
		if err != nil {
			return nil, err
		}
	}
	return ir, nil
}
