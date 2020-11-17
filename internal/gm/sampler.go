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
type reduceable interface {
	Reset() Value
	GetValue() Value
	Operators() (op Operator, invOp Operator)
}

type sampleQueue struct {
	q         []sample
	sz, start int // filled from start to end-1
	window    int
}

func (q *sampleQueue) inc(n int) int {
	return (n + 1) % len(q.q)
}

//func (q *sampleQueue) dec(n int) int {
//	return (n + len(q.q) - 1) % len(q.q)
//}
func (q *sampleQueue) push(s sample) {
	//glog.Info("Push ", s)
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
		if sz > 0 {
			last := q.last()
			if last > q.start {
				/* ---- [start][1][2][3]...[last][end] ---- */
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
	//glog.Infof("q: %+v", q)
}
func (q *sampleQueue) pop() sample {
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
func (q *sampleQueue) top() sample {
	if q.empty() {
		panic("queue is empty")
	}
	return q.q[q.last()]
}
func (q *sampleQueue) latest() sample {
	return q.top()
}

//func (q *sampleQueue) oldest() sample {
//	if q.empty() {
//		panic("queue is empty")
//	}
//	return q.q[q.start]
//}
func (q *sampleQueue) oldestIn(n int) sample {
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
func (q *sampleQueue) size() int {
	return q.sz
}
func (q *sampleQueue) last() int {
	return (q.start + q.sz - 1) % len(q.q)
}
func (q *sampleQueue) full() bool {
	return q.size() >= len(q.q)
}
func (q *sampleQueue) empty() bool {
	return q.sz == 0
}
func (q *sampleQueue) setWindow(window int) {
	if window > q.window {
		q.window = window
	}
}

type ReducerSampler struct {
	dis disposer
	r   reduceable
	q   sampleQueue
}

func (rs *ReducerSampler) dispose() {
	if rs.dis != nil {
		rs.dis()
	}
}

func (rs *ReducerSampler) SetWindow(window int) {
	rs.q.setWindow(window)
}
func (rs *ReducerSampler) takeSample() {
	var s sample
	if _, invOp := rs.r.Operators(); invOp == nil {
		s.value = rs.r.Reset()
	} else {
		s.value = rs.r.GetValue()
	}
	s.ts = time.Now()
	//glog.Infof("Push sample %v", s)
	rs.q.push(s)
}
func (rs *ReducerSampler) ValueInWindow(window int) sampleInRange {
	var s sampleInRange

	if window <= 0 {
		glog.Fatal("invalid window size ", window)
		return s
	}

	//glog.Info("window: ", window)
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
		//glog.Infof("Latest %v oldest %v now %v", latest, oldest, s.value)
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
	s.dis = AddSampler(s)
	return s
}
