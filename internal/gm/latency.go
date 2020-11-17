package gm

import (
	"math"
	"time"

	"github.com/forrestjgq/gomark/gmi"
	"github.com/golang/glog"
)

// Latency provide a convenient way to record a single latency.
// It is bound with a latency recorder which is provided at creation
type Latency struct {
	marker gmi.Marker
	start  time.Time
}

// Mark calculate how many ms has passed since creation of this Latency
// and mark it.
func (l *Latency) Mark() {
	if l.marker == nil {
		glog.Warningf("latency is cancelled")
		return
	}

	du := time.Since(l.start).Milliseconds()
	if du > math.MaxInt32 {
		glog.Warningf("latency overflow: %d ms", du)
		du = math.MaxInt32
	}
	l.marker.Mark(int32(du))
}

// Cancel will clear this latency and it will not be able to use any more.
func (l *Latency) Cancel() {
	l.start = time.Time{}
	l.marker = nil
}

// NewLatency create a latency for a Latency Recorder created by NewLatencyRecorder.
func NewLatency(marker gmi.Marker) *Latency {
	return &Latency{
		marker: marker,
		start:  time.Now(),
	}
}
