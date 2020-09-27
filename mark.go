package gomark


type Marker interface {
	Mark(n int64)
	Cancel()
}

type options struct {
	name string
	adder bool
	latency bool
	max bool
	min bool
	average bool
}

type Config func(*options)

func Adder() Config {
	return func(o *options) {
		o.adder = true
	}
}
func Latency() Config {
	return func(o *options) {
		o.latency = true
	}
}
func Max() Config {
	return func(o *options) {
		o.max = true
	}
}
func Min() Config {
	return func(o *options) {
		o.min = true
	}
}
func Average() Config {
	return func(o *options) {
		o.average = true
	}
}
func NewMarker(name string, configures ...Config) Marker {
	opt := &options{name: name}

	for _, cfg := range configures {
		if cfg != nil {
			cfg(opt)
		}
	}

	return nil
}