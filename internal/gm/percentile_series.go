package gm

type PercentileTrend struct {
	Label string                 `json:"label"`
	Data  [][]*PercentileSamples `json:"data"`
}
type PercentileSeries struct {
	op                        PercentileOperator
	divOp                     PercentileOperatorInt
	second, minute, hour, day int8
	data                      [SeriesSize]*PercentileSamples
}

func (s *PercentileSeries) Append(v *PercentileSamples) {
	s.appendSecond(v.Dup())
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
/*
func (s *PercentileSeries) GetTrend() *PercentileTrend {
	t := &PercentileTrend{
		Label: "trend",
		Data:  make([][]PercentileSamples, 60+60+30+24),
	}

	secondBegin := int(s.second)
	minuteBegin := int(s.minute)
	hourBegin := int(s.hour)
	dayBegin := int(s.day)

	c := int64(0)
	for i := 0; i < 30; i++ {
		t.Data[c] = []PercentileSamples{c, s.getDay(int8((i + dayBegin) % 30))}
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

*/
func (s *PercentileSeries) getSecond(idx int8) *PercentileSamples {
	return s.data[idx]
}
func (s *PercentileSeries) setSecond(idx int8, v *PercentileSamples) {
	s.data[idx] = v
}

func (s *PercentileSeries) getMinute(idx int8) *PercentileSamples {
	return s.data[60+idx]
}
func (s *PercentileSeries) setMinute(idx int8, v *PercentileSamples) {
	s.data[60+idx] = v
}

func (s *PercentileSeries) getHour(idx int8) *PercentileSamples {
	return s.data[120+int(idx)]
}
func (s *PercentileSeries) setHour(idx int8, v *PercentileSamples) {
	s.data[120+int(idx)] = v
}
func (s *PercentileSeries) getDay(idx int8) *PercentileSamples {
	return s.data[144+int(idx)]
}
func (s *PercentileSeries) setDay(idx int8, v *PercentileSamples) {
	s.data[144+int(idx)] = v
}

func (s *PercentileSeries) appendSecond(v *PercentileSamples) {
	s.setSecond(s.second, v)
	s.second++
	if s.second >= 60 {
		s.second = 0

		acc := s.getSecond(0).Dup()
		for i := int8(1); i < 60; i++ {
			s.op(acc, s.getSecond(i))
		}

		if s.divOp != nil {
			s.divOp(acc, 60)
		}
		s.appendMinute(acc)
	}
}
func (s *PercentileSeries) appendMinute(v *PercentileSamples) {
	s.setMinute(s.minute, v)
	s.minute++
	if s.minute >= 60 {
		s.minute = 0

		acc := s.getMinute(0).Dup()
		for i := int8(1); i < 60; i++ {
			s.op(acc, s.getMinute(i))
		}

		if s.divOp != nil {
			s.divOp(acc, 60)
		}
		s.appendHour(acc)
	}
}
func (s *PercentileSeries) appendHour(v *PercentileSamples) {
	s.setHour(s.hour, v)
	s.hour++
	if s.hour >= 24 {
		s.hour = 0

		acc := s.getHour(0).Dup()
		for i := int8(1); i < 24; i++ {
			s.op(acc, s.getHour(i))
		}

		if s.divOp != nil {
			s.divOp(acc, 24)
		}
		s.appendDay(acc)
	}
}
func (s *PercentileSeries) appendDay(v *PercentileSamples) {
	s.setDay(s.day, v)
	s.day++
	if s.day >= 30 {
		s.day = 0
	}
}

func NewPercentileSeries(op PercentileOperator, divOp PercentileOperatorInt) *PercentileSeries {
	if op == nil || divOp == nil {
		return nil
	}

	return &PercentileSeries{
		op:    op,
		divOp: divOp,
		data:  [SeriesSize]*PercentileSamples{},
	}
}
