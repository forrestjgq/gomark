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
	lr, err := gm.NewLatencyRecorder(name)
	if err != nil {
		return nil
	}
	return lr
}
