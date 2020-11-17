// Package gomark runs an HTTP server for markable variables.
//
// A variable is a gmi.Marker that provides a single kind of monitor and expose their markings through HTTP
// request: http://ip:port/vars.
//
// After variable is created, user may call Mark() to push a number. The definition of number depends
// what kind of variable you create.
//
// If you want to destroy variable, call Cancel(), after which variable can not be used.
//
// Both Mark() and Cancel() are sync call and are safe cross routines.
//
// gomark is actually a go verison of bvar, see:
//    https://github.com/apache/incubator-brpc
// for more information.
package gomark

import (
	"github.com/forrestjgq/gomark/gmi"
	"github.com/forrestjgq/gomark/internal/gm"
	"github.com/forrestjgq/gomark/internal/httpsrv"
)

// StartHTTPServer will create an http server for gomark.
func StartHTTPServer(port int) {
	if port == 0 {
		panic("gomark does not take a zero port")
	}
	httpsrv.Start(port)
}
// NewLatencyRecorder create a latency recorder.
func NewLatencyRecorder(name string) gmi.Marker {
	var ret gmi.Marker
	gm.RemoteCall(func() {
		lr, err := gm.NewLatencyRecorder(name)
		if err == nil {
			ret = lr.VarBase()
		}
	})
	return ret
}
// NewAdder create an adder.
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
// NewCounter provide a passive status for counter.
// I prefer you use NewAdder instead.
func NewCounter(name string) gmi.Marker {
	var ret gmi.Marker
	gm.RemoteCall(func() {
		c, err := gm.NewCounterWithName(name)
		if err == nil {
			ret = c.VarBase()
		}
	})
	return ret
}
// NewQPS provide QPS statistics.
func NewQPS(name string) gmi.Marker {
	var ret gmi.Marker
	gm.RemoteCall(func() {
		q, err := gm.NewQPSWithName(name)
		if err == nil {
			ret = q.VarBase()
		}
	})
	return ret
}
// NewMaxer provide maximum value collecting.
func NewMaxer(name string) gmi.Marker {
	var ret gmi.Marker
	gm.RemoteCall(func() {
		w, err := gm.NewMaxerWithName(name)
		if err == nil {
			ret = w.VarBase()
		}
	})
	return ret
}
// NewWindowMaxer collects max values in each period.
func NewWindowMaxer(name string) gmi.Marker {
	var ret gmi.Marker
	gm.RemoteCall(func() {
		w, err := gm.NewWindowMaxer(name)
		if err == nil {
			ret = w.VarBase()
		}
	})
	return ret
}

// NewPercentile create a percentile collector.
func NewPercentile() interface {
	Push(v gm.Mark)
	Dispose()
} {
	return gm.NewPercentile()
}
