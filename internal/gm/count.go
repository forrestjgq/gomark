package gm

func NewCounterNoExpose() (*PassiveStatus, error) {
	return NewCounter("", "", DisplayOnNothing)
}
func NewCounterWithName(name string) (*PassiveStatus, error) {
	return NewCounter("", name, DisplayOnAll)
}
func NewCounter(prefix, name string, filter DisplayFilter) (*PassiveStatus, error) {
	latency, _ := NewIntRecorderNoExpose()
	op, invOp := latency.Operators()

	count, err := NewPassiveStatus(prefix, name, filter, func() Value {
		return latency.GetValue() // should use value.y
	}, op, invOp, statOperatorInt)
	if err != nil {
		return nil, err
	}
	count.setReceiver(latency)
	count.SetDescriber(YValueSerializer, func(v Value, idx int) string {
		return YValueSerializer(v)
	})

	return count, nil
}
