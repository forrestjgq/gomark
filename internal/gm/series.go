package gm

type Trend struct {
	Label string    `json:"label"`
	Data  [][]int64 `json:"data"`
}
type IntSeries struct {
	op                        Operator
	second, minute, hour, day int8
	data                      [60 + 60 + 24 + 30]int64
}

func (s *IntSeries) Append(v int64) {
	s.appendSecond(v)
}

/**
template <typename T, typename Op>
void Series<T, Op>::describe(std::ostream& os,
                             const std::string* vector_names) const {
    CHECK(vector_names == NULL);
    pthread_mutex_lock(&this->_mutex);
    const int second_begin = this->_nsecond;
    const int minute_begin = this->_nminute;
    const int hour_begin = this->_nhour;
    const int day_begin = this->_nday;
    // NOTE: we don't save _data which may be inconsistent sometimes, but
    // this output is generally for "peeking the trend" and does not need
    // to exactly accurate.
    pthread_mutex_unlock(&this->_mutex);
    int c = 0;
    os << "{\"label\":\"trend\",\"data\":[";
    for (int i = 0; i < 30; ++i, ++c) {
        if (c) {
            os << ',';
        }
        os << '[' << c << ',' << this->_data.day((i + day_begin) % 30) << ']';
    }
    for (int i = 0; i < 24; ++i, ++c) {
        if (c) {
            os << ',';
        }
        os << '[' << c << ',' << this->_data.hour((i + hour_begin) % 24) << ']';
    }
    for (int i = 0; i < 60; ++i, ++c) {
        if (c) {
            os << ',';
        }
        os << '[' << c << ',' << this->_data.minute((i + minute_begin) % 60) << ']';
    }
    for (int i = 0; i < 60; ++i, ++c) {
        if (c) {
            os << ',';
        }
        os << '[' << c << ',' << this->_data.second((i + second_begin) % 60) << ']';
    }
    os << "]}";
}

*/
func (s *IntSeries) GetTrend() *Trend {
	t := &Trend{
		Label: "trend",
		Data:  make([][]int64, 60+60+30+24),
	}

	secondBegin := int(s.second)
	minuteBegin := int(s.minute)
	hourBegin := int(s.hour)
	dayBegin := int(s.day)

	c := int64(0)
	for i := 0; i < 30; i++ {
		t.Data[c] = []int64{c, s.getDay(int8((i + dayBegin) % 30))}
		c++
	}
	for i := 0; i < 24; i++ {
		t.Data[c] = []int64{c, s.getHour(int8((i + hourBegin) % 24))}
		c++
	}
	for i := 0; i < 60; i++ {
		t.Data[c] = []int64{c, s.getMinute(int8((i + minuteBegin) % 60))}
		c++
	}
	for i := 0; i < 60; i++ {
		t.Data[c] = []int64{c, s.getSecond(int8((i + secondBegin) % 60))}
		c++
	}

	return t
}
func (s *IntSeries) getSecond(idx int8) int64 {
	return s.data[idx]
}
func (s *IntSeries) setSecond(idx int8, v int64) {
	s.data[idx] = v
}

func (s *IntSeries) getMinute(idx int8) int64 {
	return s.data[60+idx]
}
func (s *IntSeries) setMinute(idx int8, v int64) {
	s.data[60+idx] = v
}

func (s *IntSeries) getHour(idx int8) int64 {
	return s.data[120+int(idx)]
}
func (s *IntSeries) setHour(idx int8, v int64) {
	s.data[120+int(idx)] = v
}
func (s *IntSeries) getDay(idx int8) int64 {
	return s.data[144+int(idx)]
}
func (s *IntSeries) setDay(idx int8, v int64) {
	s.data[144+int(idx)] = v
}

func (s *IntSeries) appendSecond(v int64) {
	s.setSecond(s.second, v)
	s.second++
	if s.second >= 60 {
		s.second = 0

		acc := s.getSecond(0)
		for i := int8(1); i < 60; i++ {
			acc = s.op(acc, s.getSecond(i))
		}

		m := acc / 60
		s.appendMinute(m)
	}
}
func (s *IntSeries) appendMinute(v int64) {
	s.setMinute(s.minute, v)
	s.minute++
	if s.minute >= 60 {
		s.minute = 0

		acc := s.getMinute(0)
		for i := int8(1); i < 60; i++ {
			acc = s.op(acc, s.getMinute(i))
		}

		m := acc / 60
		s.appendHour(m)
	}
}
func (s *IntSeries) appendHour(v int64) {
	s.setHour(s.hour, v)
	s.hour++
	if s.hour >= 24 {
		s.hour = 0

		acc := s.getHour(0)
		for i := int8(1); i < 24; i++ {
			acc = s.op(acc, s.getHour(i))
		}

		m := acc / 24
		s.appendDay(m)
	}
}
func (s *IntSeries) appendDay(v int64) {
	s.setDay(s.day, v)
	s.day++
	if s.day >= 30 {
		s.day = 0
	}
}

func NewIntSeries(op Operator) *IntSeries {
	if op == nil {
		return nil
	}

	return &IntSeries{
		op:   op,
		data: [174]int64{},
	}
}
