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
	"github.com/golang/glog"
)

// StartHTTPServer will create an http server for gomark.
func StartHTTPServer(port int) {
	if port == 0 {
		panic("gomark does not take a zero port")
	}
	httpsrv.Start(port)
}

// Request will docker gomark to your own http server
// See http_server.go for detailed information, there Start() will start an http
// server based on mux, and it will call the same API as this calls.
func Request(req *gmi.Request) *gmi.Response {
	return httpsrv.RequestHTTP(req)
}

// NewLatencyRecorder create a latency recorder.
func NewLatencyRecorder(name string) gmi.Marker {
	var ret gmi.Marker
	gm.RemoteCall(func() {
		lr, err := gm.NewLatencyRecorder(name)
		if err == nil {
			ret = lr.VarBase().Marker()
		} else {
			glog.Errorf("create latency recorder(%s) fails, err: %v", name, err)
		}
	})
	return ret
}

// NewAdder create an adder with series.
func NewAdder(name string) gmi.Marker {
	var ret gmi.Marker
	gm.RemoteCall(func() {
		add, err := gm.NewAdder(name)
		if err == nil {
			ret = add.VarBase().Marker()
		} else {
			glog.Errorf("create adder(%s) fails, err: %v", name, err)
		}
	})
	return ret
}
func NewAdderPerSecond(name string) gmi.Marker {
	var ret gmi.Marker
	gm.RemoteCall(func() {
		add, err := gm.NewAdderPersecond(name)
		if err == nil {
			ret = add.VarBase().Marker()
		} else {
			glog.Errorf("create adder per second(%s) fails, err: %v", name, err)
		}
	})
	return ret
}
func NewStatus(name string) gmi.Marker {
	var ret gmi.Marker
	gm.RemoteCall(func() {
		add, err := gm.NewStatus(name)
		if err == nil {
			ret = add.VarBase().Marker()
		} else {
			glog.Errorf("create status(%s) fails, err: %v", name, err)
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
			ret = c.VarBase().Marker()
		} else {
			glog.Errorf("create counter(%s) fails, err: %v", name, err)
		}
	})
	return ret
}

// NewQPS provide QPS statistics.
func NewQPS(name string) gmi.Marker {
	var ret gmi.Marker
	gm.RemoteCall(func() {
		q, err := gm.NewQPS(name)
		if err == nil {
			ret = q.VarBase().Marker()
		} else {
			glog.Errorf("create qps(%s) fails, err: %v", name, err)
		}
	})
	return ret
}

// NewMaxer saves max value(no series)
func NewMaxer(name string) gmi.Marker {
	var ret gmi.Marker
	gm.RemoteCall(func() {
		w, err := gm.NewMaxer(name)
		if err == nil {
			ret = w.VarBase().Marker()
		} else {
			glog.Errorf("create maxer(%s) fails, err: %v", name, err)
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
			ret = w.VarBase().Marker()
		} else {
			glog.Errorf("create windown maxer(%s) fails, err: %v", name, err)
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

func IntToMarker(id int) gmi.Marker {
	t := gm.Identity(id)
	return t
}
