package gm

import (
	"time"

	"github.com/golang/glog"
)

type sample struct {
	value Value
	ts    time.Time
}
type sampleInRange struct {
	value Value
	du    time.Duration
}
type sampler interface {
	takeSample()
	destroy()
}
type reduceable interface {
	Reset() Value
	GetValue() Value
	Operators() (op Operator, invOp Operator)
}

type sampleQueue struct {
	q          []sample
	end, start int // filled from start to end-1
	window     int
}

func (q *sampleQueue) inc(n int) int {
	return (n + 1) % len(q.q)
}
func (q *sampleQueue) dec(n int) int {
	return (n + len(q.q) - 1) % len(q.q)
}
func (q *sampleQueue) push(s sample) {
	if q.window+1 > len(q.q) {
		if q.window == 0 {
			q.window = 1
		}

		newlen := len(q.q) * 2
		if q.window+1 > newlen {
			newlen = q.window + 1
		}
		tq := make([]sample, newlen)
		sz := q.size()
		if q.end > q.start {
			copy(tq[0:sz], q.q[q.start:q.end])
		} else if q.end < q.start {
			seg := sz - q.start
			copy(tq[0:seg], q.q[q.start:])
			copy(tq[seg:sz], q.q[0:q.end])
		}
		q.start, q.end = 0, sz
	}

	if q.full() {
		_ = q.pop()
	}
	q.q[q.end] = s
	q.end = q.inc(q.end)
}
func (q *sampleQueue) pop() sample {
	if q.empty() {
		panic("queue is empty")
	}
	s := q.q[q.start]
	q.start = q.inc(q.start)
	if q.empty() {
		q.end, q.start = 0, 0
	}
	return s
}
func (q *sampleQueue) top() sample {
	if q.empty() {
		panic("queue is empty")
	}
	return q.q[q.dec(q.end)]
}
func (q *sampleQueue) latest() sample {
	return q.top()
}
func (q *sampleQueue) oldest() sample {
	return q.bottom()
}
func (q *sampleQueue) bottom() sample {
	if q.empty() {
		panic("queue is empty")
	}
	return q.q[q.start]
}
func (q *sampleQueue) oldestIn(n int) sample {
	if q.empty() {
		panic("queue is empty")
	}
	if n <= 0 {
		return q.top()
	}
	if n >= q.size() {
		return q.bottom()
	}

	idx := (n + len(q.q) - 1) % len(q.q)
	return q.q[idx]
}
func (q *sampleQueue) size() int {
	if q.start == q.end {
		return 0
	}
	return (q.end + len(q.q) - q.start) % len(q.q)
}
func (q *sampleQueue) full() bool {
	return q.size() >= len(q.q)
}
func (q *sampleQueue) empty() bool {
	return q.start == q.end
}
func (q *sampleQueue) setWindow(window int) {
	if window > q.window {
		q.window = window
	}
}

type ReducerSampler struct {
	r reduceable
	q sampleQueue
}

func (rs *ReducerSampler) SetWindow(window int) {
	rs.q.setWindow(window)
}
func (rs *ReducerSampler) takeSample() {
	var s sample
	if _, invOp := rs.r.Operators(); invOp != nil {
		s.value = rs.r.Reset()
	} else {
		s.value = rs.r.GetValue()
	}
	s.ts = time.Now()
	rs.q.push(s)
}
func (rs *ReducerSampler) ValueInWindow(window int) sampleInRange {
	var s sampleInRange

	if window <= 0 {
		glog.Fatal("invalid window size ", window)
		return s
	}

	if rs.q.size() <= 1 {
		return s
	}

	oldest := rs.q.oldestIn(window)
	latest := rs.q.latest()
	op, inv := rs.r.Operators()
	if inv == nil {
		s.value = latest.value
		for i := 1; true; i++ {
			e := rs.q.oldestIn(i)
			if e.ts == oldest.ts {
				break
			}
			s.value = op(s.value, e.value)
		}
	} else {
		s.value = inv(latest.value, oldest.value)
	}
	s.du = latest.ts.Sub(oldest.ts)
	return s
}
func (rs *ReducerSampler) SamplesInWindow(window int) (ret []Value) {
	if window <= 0 {
		glog.Fatal("invalid window size ", window)
		return
	}

	if rs.q.size() <= 1 {
		return
	}

	oldest := rs.q.oldestIn(window)
	for i := 1; true; i++ {
		e := rs.q.oldestIn(i)
		if e.ts == oldest.ts {
			break
		}
		ret = append(ret, e.value)
	}
	return
}

func NewReducerSampler(r reduceable) *ReducerSampler {
	s := &ReducerSampler{
		r: r,
	}
	s.SetWindow(1)
	return s
}
