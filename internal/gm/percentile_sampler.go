package gm

import (
	"time"

	"github.com/golang/glog"
)

type PercentileOperator func(left, right *PercentileSamples)
type PercentileOperatorInt func(left *PercentileSamples, right int)
type PercentileReducer interface {
	Reset() *PercentileSamples
	GetValue() *PercentileSamples
	Operators() (op PercentileOperator, invOp PercentileOperator)
}

type PercentileSample struct {
	value *PercentileSamples
	ts    time.Time
}
type PercentileSampleInRange struct {
	value *PercentileSamples
	du    time.Duration
}
type PercentileSampleQueue struct {
	q         []PercentileSample
	sz, start int
	window    int
}

func (q *PercentileSampleQueue) inc(n int) int {
	return (n + 1) % len(q.q)
}

//func (q *PercentileSampleQueue) dec(n int) int {
//	return (n + len(q.q) - 1) % len(q.q)
//}
func (q *PercentileSampleQueue) push(s PercentileSample) {
	if q.window+1 > len(q.q) {
		if q.window == 0 {
			q.window = 1
		}

		newlen := len(q.q) * 2
		if q.window+1 > newlen {
			newlen = q.window + 1
		}
		tq := make([]PercentileSample, newlen)
		sz := q.size()
		if sz > 0 {
			last := q.last()
			if last > q.start {
				copy(tq[0:sz], q.q[q.start:last+1])
			} else if last < q.start {
				seg := sz - q.start
				copy(tq[0:seg], q.q[q.start:])
				copy(tq[seg:sz], q.q[0:last+1])
			}
		}
		q.start = 0
		q.q = tq
	}

	if q.full() {
		_ = q.pop()
	}
	last := q.inc(q.last())
	q.q[last] = s
	q.sz++
}
func (q *PercentileSampleQueue) pop() PercentileSample {
	if q.empty() {
		panic("queue is empty")
	}
	s := q.q[q.start]
	q.start = q.inc(q.start)
	q.sz--
	if q.empty() {
		q.start = 0
	}
	return s
}
func (q *PercentileSampleQueue) top() PercentileSample {
	if q.empty() {
		panic("queue is empty")
	}
	return q.q[q.last()]
}
func (q *PercentileSampleQueue) latest() PercentileSample {
	return q.top()
}

//func (q *PercentileSampleQueue) oldest() PercentileSample {
//	return q.bottom()
//}
//func (q *PercentileSampleQueue) bottom() PercentileSample {
//	if q.empty() {
//		panic("queue is empty")
//	}
//	return q.q[q.start]
//}
func (q *PercentileSampleQueue) oldestIn(n int) PercentileSample {
	if q.empty() {
		panic("queue is empty")
	}
	if n < 0 {
		n = 0
	} else if n >= q.size() {
		n = q.size() - 1
	}
	return q.q[(q.start+n)%len(q.q)]
}
func (q *PercentileSampleQueue) size() int {
	return q.sz
}
func (q *PercentileSampleQueue) last() int {
	return (q.start + q.sz - 1) % len(q.q)
}
func (q *PercentileSampleQueue) full() bool {
	return q.size() >= len(q.q)
}
func (q *PercentileSampleQueue) empty() bool {
	return q.sz == 0
}
func (q *PercentileSampleQueue) setWindow(window int) {
	if window > q.window {
		q.window = window
	}
}

type PercentileReducerSampler struct {
	dis disposer
	r   PercentileReducer
	q   PercentileSampleQueue
}

func (rs *PercentileReducerSampler) Dispose() {
	if rs.dis != nil {
		rs.dis()
	}
}

func (rs *PercentileReducerSampler) SetWindow(window int) {
	rs.q.setWindow(window)
}
func (rs *PercentileReducerSampler) takeSample() {
	var s PercentileSample
	if _, invOp := rs.r.Operators(); invOp == nil {
		s.value = rs.r.Reset()
	} else {
		s.value = rs.r.GetValue()
	}
	s.ts = time.Now()
	rs.q.push(s)
}
func (rs *PercentileReducerSampler) ValueInWindow(window int) PercentileSampleInRange {
	var s PercentileSampleInRange
	if window <= 0 {
		glog.Fatal("invalid window size ", window)
		s.value = &PercentileSamples{}
		return s
	}

	if rs.q.size() <= 1 {
		s.value = &PercentileSamples{}
		return s
	}

	oldest := rs.q.oldestIn(window)
	latest := rs.q.latest()
	op, inv := rs.r.Operators()
	if inv == nil {
		s.value = latest.value.Dup()
		for i := 1; true; i++ {
			e := rs.q.oldestIn(i)
			if e.ts == oldest.ts {
				break
			}
			op(s.value, e.value)
		}
	} else {
		s.value = latest.value.Dup()
		inv(latest.value, oldest.value)
	}
	s.du = latest.ts.Sub(oldest.ts)
	return s
}
func (rs *PercentileReducerSampler) SamplesInWindow(window int) (ret []*PercentileSamples) {
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
		ret = append(ret, e.value.Dup())
	}
	return
}

func NewPercentileReducerSampler(r PercentileReducer) *PercentileReducerSampler {
	s := &PercentileReducerSampler{
		r: r,
	}
	s.SetWindow(1)
	s.dis = AddSampler(s)
	return s
}
