package gm

import (
	"errors"
	"io"
	"strconv"
)

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
		ir.sampler = nil
	}
}

func (ir *IntRecorder) Describe(w io.StringWriter, _ bool) {
	v := ir.IntAverage()
	if v != 0 {
		_, _ = w.WriteString(strconv.Itoa(int(v)))
	} else {
		_, _ = w.WriteString(strconv.FormatFloat(ir.FloatAverage(), 'f', 3, 64))
	}
}

func (ir *IntRecorder) DescribeSeries(_ io.StringWriter, _ *SeriesOption) error {
	return errors.New("describe series not supported")
}

func (ir *IntRecorder) Operators() (op Operator, invOp Operator) {
	op, invOp = ir.op, ir.invOp
	return
}

func (ir *IntRecorder) Push(v Mark) {
	last := ir.value
	ir.value.x += int64(v)
	ir.value.y += 1
	//glog.Infof("IntRecord value: %v", ir.value)
	if (v > 0 && last.x > 0 && ir.value.x < 0) || last.y < 0 {
		ir.value = Value{}
	} else if v < 0 && last.x < 0 && ir.value.x > 0 {
		ir.value = Value{}
	}
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
	return NewIntRecorder(name, "recoder", DisplayOnAll)
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
