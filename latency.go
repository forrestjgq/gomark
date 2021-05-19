package gomark

import (
	"math"
	"time"

	"github.com/forrestjgq/gomark/gmi"
	"github.com/forrestjgq/glog"
)

// Latency provide a convenient way to record a single latency.
// It is bound with a latency recorder which is provided at creation
type Latency struct {
	marker gmi.Marker
	start  time.Time
	latency int32
}

// Mark calculate how many ms has passed since creation of this Latency
// and mark it.
func (l *Latency) Mark() {
	du := time.Since(l.start).Milliseconds()
	if du > math.MaxInt32 {
		glog.Warningf("latency overflow: %d ms", du)
		du = math.MaxInt32
	}
	l.latency = int32(du)
	if l.marker != nil {
		l.marker.Mark(l.latency)
	}
}

// Cancel will clear this latency and it will not be able to use any more.
func (l *Latency) Cancel() {
	l.start = time.Time{}
	l.marker = nil
}
func (l *Latency) Latency() int32 {
	return l.latency
}

// NewLatency create a latency for a Latency Recorder created by NewLatencyRecorder.
func NewLatency(marker gmi.Marker) *Latency {
	return &Latency{
		marker: marker,
		start:  time.Now(),
	}
}
