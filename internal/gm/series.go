package gm

import (
	"fmt"
	"io"
	"strconv"
)

const (
	SeriesSize = 60 + 60 + 24 + 30
)

type Trend struct {
	Label string  `json:"label"`
	Data  []Value `json:"data"`
}
type IntSeries struct {
	op                        Operator
	divOp                     OperatorInt
	second, minute, hour, day int8
	data                      [SeriesSize]Value
}
type ValueConverter func(v Value, idx int) string

func (s *IntSeries) Append(v Value) {
	s.appendSecond(v)
}

func (s *IntSeries) GetTrend() *Trend {
	t := &Trend{
		Label: "trend",
		Data:  make([]Value, SeriesSize),
	}

	secondBegin := int(s.second)
	minuteBegin := int(s.minute)
	hourBegin := int(s.hour)
	dayBegin := int(s.day)

	c := int64(0)
	for i := 0; i < 30; i++ {
		t.Data[c] = s.getDay(int8((i + dayBegin) % 30))
		c++
	}
	for i := 0; i < 24; i++ {
		t.Data[c] = s.getHour(int8((i + hourBegin) % 24))
		c++
	}
	for i := 0; i < 60; i++ {
		t.Data[c] = s.getMinute(int8((i + minuteBegin) % 60))
		c++
	}
	for i := 0; i < 60; i++ {
		t.Data[c] = s.getSecond(int8((i + secondBegin) % 60))
		c++
	}

	return t
}
func (s *IntSeries) Describe(w io.StringWriter, splitName []string, cvt ValueConverter) {
	t := s.GetTrend()
	if splitName == nil {
		_, _ = w.WriteString("{\"label\":\"trend\",\"data\":[")
		for i, v := range t.Data {
			if i > 0 {
				_, _ = w.WriteString(",")
			}
			_, _ = w.WriteString(fmt.Sprintf("[%d,%s]", i, cvt(v, 0)))
		}
		_, _ = w.WriteString("]}")
	} else {
		_, _ = w.WriteString("[")
		for j, s := range splitName {
			if j > 0 {
				_, _ = w.WriteString(",")
			}
			name := s
			if len(name) == 0 {
				name = "Vector[" + strconv.Itoa(j) + "]"
			}
			_, _ = w.WriteString(fmt.Sprintf("{\"label\":\"%s\",\"data\":[", name))
			for i, v := range t.Data {
				if i > 0 {
					_, _ = w.WriteString(",")
				}
				_, _ = w.WriteString(fmt.Sprintf("[%d,%s]", i, cvt(v, 0)))
			}
			_, _ = w.WriteString("]}")
		}
		_, _ = w.WriteString("]")
	}
}

func (s *IntSeries) getSecond(idx int8) Value {
	return s.data[idx]
}
func (s *IntSeries) setSecond(idx int8, v Value) {
	s.data[idx] = v
}

func (s *IntSeries) getMinute(idx int8) Value {
	return s.data[60+idx]
}
func (s *IntSeries) setMinute(idx int8, v Value) {
	s.data[60+idx] = v
}

func (s *IntSeries) getHour(idx int8) Value {
	return s.data[120+int(idx)]
}
func (s *IntSeries) setHour(idx int8, v Value) {
	s.data[120+int(idx)] = v
}
func (s *IntSeries) getDay(idx int8) Value {
	return s.data[144+int(idx)]
}
func (s *IntSeries) setDay(idx int8, v Value) {
	s.data[144+int(idx)] = v
}

func (s *IntSeries) appendSecond(v Value) {
	s.setSecond(s.second, v)
	s.second++
	if s.second >= 60 {
		s.second = 0

		acc := s.getSecond(0)
		for i := int8(1); i < 60; i++ {
			acc = s.op(acc, s.getSecond(i))
		}

		if s.divOp != nil {
			acc = s.divOp(acc, 60)
		}
		s.appendMinute(acc)
	}
}
func (s *IntSeries) appendMinute(v Value) {
	s.setMinute(s.minute, v)
	s.minute++
	if s.minute >= 60 {
		s.minute = 0

		acc := s.getMinute(0)
		for i := int8(1); i < 60; i++ {
			acc = s.op(acc, s.getMinute(i))
		}

		if s.divOp != nil {
			acc = s.divOp(acc, 60)
		}
		s.appendHour(acc)
	}
}
func (s *IntSeries) appendHour(v Value) {
	s.setHour(s.hour, v)
	s.hour++
	if s.hour >= 24 {
		s.hour = 0

		acc := s.getHour(0)
		for i := int8(1); i < 24; i++ {
			acc = s.op(acc, s.getHour(i))
		}

		if s.divOp != nil {
			acc = s.divOp(acc, 24)
		}
		s.appendDay(acc)
	}
}
func (s *IntSeries) appendDay(v Value) {
	s.setDay(s.day, v)
	s.day++
	if s.day >= 30 {
		s.day = 0
	}
}

func NewIntSeries(op Operator, divOp OperatorInt) *IntSeries {
	if op == nil || divOp == nil {
		return nil
	}

	return &IntSeries{
		op:    op,
		divOp: divOp,
		data:  [SeriesSize]Value{},
	}
}
