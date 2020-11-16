package gomark

import (
	"github.com/forrestjgq/gomark/gmi"
	"github.com/forrestjgq/gomark/internal/gm"
	"github.com/forrestjgq/gomark/internal/httpsrv"
)

// StartHTTPServer will create an http server for gomark.
func StartHTTPServer(port int) {
	httpsrv.Start(port)
}
func NewLatencyRecorder(name string) gmi.Marker {
	var ret gmi.Marker
	gm.RemoteCall(func() {
		lr, err := gm.NewLatencyRecorder(name)
		if err == nil {
			ret = lr
		}
	})
	return ret
}
func NewAdder(name string) gmi.Marker {
	var ret gmi.Marker
	gm.RemoteCall(func() {
		add, err := gm.NewAdderWithName(name)
		if err == nil {
			ret = add.VarBase()
		}
	})
	return ret
}
func NewCounter(name string) gmi.Marker {
	var ret gmi.Marker
	gm.RemoteCall(func() {
		c, err := gm.NewCounterWithName(name)
		if err != nil {
			ret = nil
		} else {
			ret = c.VarBase()
		}
	})
	return ret
}
func NewQPS(name string) gmi.Marker {
	var ret gmi.Marker
	gm.RemoteCall(func() {
		q, err := gm.NewQPSWithName(name)
		if err != nil {
			ret = nil
		} else {
			ret = q.VarBase()
		}
	})
	return ret
}
func NewWindowMaxer(name string) gmi.Marker {
	var ret gmi.Marker
	gm.RemoteCall(func() {
		w, err := gm.NewWindowMaxer(name)
		if err != nil {
			ret = nil
		} else {
			ret = w.VarBase()
		}
	})
	return ret
}

func NewPercentile() interface {
	Push(v gm.Mark)
} {
	return gm.NewPercentile()
}
