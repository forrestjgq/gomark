package gomark

import (
	"github.com/forrestjgq/gomark/gmi"
	"github.com/forrestjgq/gomark/internal/gm"
	"github.com/forrestjgq/gomark/internal/httpsrv"
)

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
		add, err := gm.NewAdder(name)
		if err == nil {
			ret = add
		}
	})
	return ret
}
func NewCounter(name string) gmi.Marker {
	var ret gmi.Marker
	gm.RemoteCall(func() {
		ret = gm.NewCounter(name)
	})
	return ret
}
func NewQPS(name string) gmi.Marker {
	var ret gmi.Marker
	gm.RemoteCall(func() {
		ret = gm.NewQPS(name)
	})
	return ret
}
func NewWindowMaxer(name string) gmi.Marker {
	var ret gmi.Marker
	gm.RemoteCall(func() {
		ret = gm.NewWindowMaxer(name)
	})
	return ret
}
