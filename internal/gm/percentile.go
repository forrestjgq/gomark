package gm

import (
	"fmt"
	"io"
	"math"
	"reflect"
	"sort"

	"github.com/golang/glog"

	fr "github.com/valyala/fastrand"
)

// IntSlice attaches the methods of Interface to []int, sorting in increasing order.
type UIntSlice []uint32

func (p UIntSlice) Len() int           { return len(p) }
func (p UIntSlice) Less(i, j int) bool { return p[i] < p[j] }
func (p UIntSlice) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }

type PercentileInterval struct {
	sampleSize int
	numAdded   uint32
	numSamples uint16
	samples    []uint32
	sorted     bool
}

func (pi *PercentileInterval) CopyFrom(rhs *PercentileInterval) {
	pi.sampleSize = rhs.sampleSize
	pi.numAdded = rhs.numAdded
	pi.numSamples = rhs.numSamples
	pi.sorted = rhs.sorted
	pi.samples = make([]uint32, len(rhs.samples))
	copy(pi.samples, rhs.samples)
}
func (pi *PercentileInterval) SampleAt(idx int) uint32 {
	saved := int(pi.numSamples)
	if idx > saved {
		if saved == 0 {
			return 0
		}
		idx = saved - 1
	}

	if !pi.sorted {
		sort.Sort(UIntSlice(pi.samples[0:saved]))
		pi.sorted = true
	}

	if saved != int(pi.numSamples) {
		panic("You must call get_number() on a unchanging PercentileInterval")
	}
	return pi.samples[idx]
}
func roundOfExpectation(a, b uint) uint {
	if b == 0 {
		return 0
	}
	z := uint(0)
	r := uint(fr.Uint32n(uint32(b))) % b
	if r < a%b {
		z = 1
	}
	return a/b + z
}
func (pi *PercentileInterval) Merge(rhs *PercentileInterval) {
	if rhs.numAdded == 0 {
		return
	}
	if pi.sampleSize < rhs.sampleSize {
		panic("must merge small interval into larger one currently")
	}

	if int(rhs.numSamples) != int(rhs.numAdded) {
		panic("rhs num sample != num added")
	}

	if int(pi.numAdded+rhs.numAdded) <= pi.sampleSize {
		if int(pi.numSamples) != int(pi.numAdded) {
			glog.Fatalf("numAdded %v rhs numAdded %v numSamples %v rhs numSamples %v sampleSize %v rhs sampleSize %v",
				pi.numAdded, rhs.numAdded, pi.numSamples, rhs.numSamples, pi.sampleSize, rhs.sampleSize)
		}
		copy(pi.samples[int(pi.numSamples):int(pi.numSamples+rhs.numSamples)], rhs.samples[0:rhs.numSamples])
		pi.numSamples += rhs.numSamples
	} else {
		numRemain := roundOfExpectation(uint(pi.numAdded)*uint(pi.sampleSize), uint(pi.numAdded)+uint(rhs.numAdded))
		if numRemain > uint(pi.numSamples) {
			glog.Fatalf("remain: %v samples %v\n", numRemain, pi.numSamples)
		}
		for i := uint(pi.numSamples); i > numRemain; i-- {
			r := fr.Uint32n(uint32(i))
			pi.samples[int(r)] = pi.samples[int(i-1)]
		}

		numRemainFromRhs := uint(pi.sampleSize) - numRemain
		if numRemainFromRhs > uint(rhs.numSamples) {
			glog.Fatalf("remian from rhs %v num samples %v\n", numRemainFromRhs, rhs.numSamples)
		}

		tmp := make([]uint32, rhs.numSamples)
		copy(tmp, rhs.samples[0:rhs.numSamples])
		for i := uint(0); i < numRemainFromRhs; i++ {
			idx := fr.Uint32n(uint32(rhs.numSamples) - uint32(i))
			pi.samples[numRemain] = tmp[idx]
			numRemain++
			tmp[idx] = tmp[uint(rhs.numSamples)-i-1]
		}
		pi.numSamples = uint16(numRemain)
		if int(pi.numSamples) != pi.sampleSize {
			glog.Fatalf("numSamples %v sampleSize %v\n", pi.numSamples, pi.sampleSize)
		}
	}
	pi.numAdded += rhs.numAdded
}
func (pi *PercentileInterval) MergeWithExpectation(rhs *PercentileInterval, n uint16) {
	if n > rhs.numSamples {
		glog.Fatalf("n %v rhs.numSamples %v\n", n, rhs.numSamples)
	}
	pi.numAdded += rhs.numAdded
	if pi.numSamples+n <= uint16(pi.sampleSize) && n == rhs.numSamples {
		copy(pi.samples[pi.numSamples:pi.numSamples+n], rhs.samples[0:n])
		pi.numSamples += n
		return
	}

	for i := uint16(0); i < n; i++ {
		idx := fr.Uint32n(uint32(rhs.numSamples - i))
		if int(pi.numSamples) < pi.sampleSize {
			pi.samples[pi.numSamples] = rhs.samples[idx]
			pi.numSamples++
		} else {
			pi.samples[fr.Uint32n(uint32(pi.numSamples))] = rhs.samples[idx]
		}
		where := rhs.numSamples - i - 1
		rhs.samples[idx], rhs.samples[where] = rhs.samples[where], rhs.samples[idx] // swap
	}
}

func (pi *PercentileInterval) Add32(x uint32) bool {
	if int(pi.numSamples) >= pi.sampleSize {
		glog.Error("this interval was full")
		return false
	}
	pi.numAdded++
	pi.samples[pi.numSamples] = x
	pi.numSamples++
	return true
}
func (pi *PercentileInterval) Add64(x int64) bool {
	if x >= 0 {
		return pi.Add32(uint32(x))
	}
	return false
}
func (pi *PercentileInterval) Clear() {
	pi.numAdded = 0
	pi.sorted = false
	pi.numSamples = 0
}
func (pi *PercentileInterval) Full() bool {
	return int(pi.numSamples) == pi.sampleSize
}
func (pi *PercentileInterval) Empty() bool {
	return pi.numSamples == 0
}

func (pi *PercentileInterval) AddedCount() uint32 {
	return pi.numAdded
}
func (pi *PercentileInterval) SampleCount() uint16 {
	return pi.numSamples
}
func (pi *PercentileInterval) SameOf(rhs *PercentileInterval) bool {
	return pi.numAdded == rhs.numAdded &&
		pi.numSamples == rhs.numSamples &&
		reflect.DeepEqual(pi.samples[0:pi.numSamples], rhs.samples[0:rhs.numSamples])
}

func (pi *PercentileInterval) check() bool {

	n := pi.samples[0]
	for i := 1; i < int(pi.numSamples); i++ {
		if pi.samples[i] != n+uint32(i) {
			return false
		}
	}
	return true
}
func (pi *PercentileInterval) Describe(w io.StringWriter) {
	_, _ = w.WriteString(fmt.Sprintf("(num_added=%d)[", pi.AddedCount()))
	for i := 0; i < int(pi.numSamples); i++ {
		_, _ = w.WriteString(fmt.Sprintf(" %d", pi.samples[i]))
	}
	_, _ = w.WriteString("]\n")
}

func newPercentileInterval(sampleSize int) *PercentileInterval {
	return &PercentileInterval{
		sampleSize: sampleSize,
		samples:    make([]uint32, sampleSize),
	}
}

////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////

const (
	NumIntervals = 32
)

type PercentileSamples struct {
	sampleSize int
	numAdded   int
	intervals  []*PercentileInterval
}

func (ps *PercentileSamples) DupFrom(rhs *PercentileSamples) {
	ps.numAdded = rhs.numAdded
	ps.sampleSize = rhs.sampleSize
	if ps.intervals == nil {
		ps.intervals = make([]*PercentileInterval, NumIntervals)
	}
	copy(rhs.intervals, ps.intervals)
}
func (ps *PercentileSamples) CopyFrom(rhs *PercentileSamples) {
	ps.numAdded = rhs.numAdded
	ps.sampleSize = rhs.sampleSize
	for i, v := range rhs.intervals {
		if v != nil && !v.Empty() {
			ps.intervals[i] = &PercentileInterval{}
			ps.intervals[i].CopyFrom(v)
		} else {
			ps.intervals[i] = nil
		}
	}
}
func (ps *PercentileSamples) TakeFrom(rhs *PercentileSamples) {
	ps.numAdded = rhs.numAdded
	for i, v := range rhs.intervals {
		if v != nil && !v.Empty() {
			ps.intervals[i] = &PercentileInterval{}
			ps.intervals[i].CopyFrom(v)
		} else {
			ps.intervals[i].Clear()
		}
	}
}
func (ps *PercentileSamples) Clear() {
	ps.numAdded = 0
	for i := range ps.intervals {
		ps.intervals[i] = nil
	}
}
func (ps *PercentileSamples) Dup() *PercentileSamples {
	rhs := &PercentileSamples{
		sampleSize: ps.sampleSize,
		numAdded:   ps.numAdded,
		intervals:  make([]*PercentileInterval, NumIntervals),
	}
	copy(rhs.intervals, ps.intervals)
	return rhs
}

func (ps *PercentileSamples) GetNumber(ratio float64) uint32 {
	n := int(math.Ceil(ratio * float64(ps.numAdded)))
	if n > ps.numAdded {
		n = ps.numAdded
	} else if n == 0 {
		return 0
	}
	for _, v := range ps.intervals {
		if v == nil {
			continue
		}
		if n <= int(v.AddedCount()) {
			samplen := n * int(v.SampleCount()) / int(v.AddedCount())
			sampleIdx := 0
			if samplen != 0 {
				sampleIdx = samplen - 1
			}
			//glog.Infof("get sample from interval %d sample size %d at idx %d", k, v.sampleSize,  sampleIdx)
			return v.SampleAt(sampleIdx)
		}
		n -= int(v.AddedCount())
	}
	panic("can not reach here")
}
func (ps *PercentileSamples) Merge(rhs *PercentileSamples) {
	ps.numAdded += rhs.numAdded
	for i, v := range rhs.intervals {
		if v != nil && !v.Empty() {
			if ps.intervals[i] == nil {
				ps.intervals[i] = newPercentileInterval(ps.sampleSize)
			}
			ps.intervals[i].Merge(v)
		}
	}
}
func (ps *PercentileSamples) IntervalOf(idx int) *PercentileInterval {
	if ps.intervals[idx] == nil {
		ps.intervals[idx] = newPercentileInterval(ps.sampleSize)
	}
	return ps.intervals[idx]
}
func (ps *PercentileSamples) CombineOf(many []*PercentileSamples) {
	if ps.numAdded != 0 {
		for _, v := range ps.intervals {
			v.Clear()
		}
		ps.numAdded = 0
	}

	for _, v := range many {
		ps.numAdded += v.numAdded
	}

	for i := 0; i < NumIntervals; i++ {
		total, totalSample := 0, 0
		for _, v := range many {
			if v.intervals[i] != nil {
				total += int(v.intervals[i].AddedCount())
				totalSample += int(v.intervals[i].SampleCount())
			}
		}
		if total == 0 {
			continue
		}

		for _, v := range many {
			vi := v.intervals[i]
			if vi == nil || vi.Empty() {
				continue
			}

			invl := &PercentileInterval{}
			invl.CopyFrom(vi)
			if total <= ps.sampleSize {
				ps.IntervalOf(i).MergeWithExpectation(invl, invl.SampleCount())
				continue
			}

			b := int(invl.AddedCount())
			remain := roundOfExpectation(uint(b*ps.sampleSize), uint(total))
			if remain > uint(invl.sampleSize) {
				remain = uint(invl.sampleSize)
			}
			ps.IntervalOf(i).MergeWithExpectation(invl, uint16(remain))
		}
	}
}
func (ps *PercentileSamples) Describe(w io.StringWriter) {
	_, _ = w.WriteString(fmt.Sprintf("{num_added=%d", ps.numAdded))
	for i := 0; i < NumIntervals; i++ {
		if ps.intervals[i] != nil && !ps.intervals[i].Empty() {
			_, _ = w.WriteString(fmt.Sprintf(" interval[%d]=", i))
			ps.intervals[i].Describe(w)
		}
	}
	_, _ = w.WriteString("}")
}
func NewPercentileSamples(sampleSize int) *PercentileSamples {
	return &PercentileSamples{
		sampleSize: sampleSize,
		numAdded:   0,
		intervals:  make([]*PercentileInterval, NumIntervals),
	}
}

const (
	GlobalPercentileSamplesSize      = 254
	ThreadLocalPercentileSamplesSize = 30
)

type Percentile struct {
	sampler *PercentileReducerSampler
	value   *PercentileSamples
	local   *PercentileSamples
}

func ones32(x uint32) uint32 {
	/* 32-bit recursive reduction using SWAR...
	 * but first step is mapping 2-bit values
	 * into sum of 2 1-bit values in sneaky way
	 */
	x -= (x >> 1) & 0x55555555
	x = ((x >> 2) & 0x33333333) + (x & 0x33333333)
	x = ((x >> 4) + x) & 0x0f0f0f0f
	x += x >> 8
	x += x >> 16
	return x & 0x0000003f
}

func log2(x uint32) uint32 {
	y := int32(x & (x - 1))
	y |= -y
	y >>= 31
	x |= x >> 1
	x |= x >> 2
	x |= x >> 4
	x |= x >> 8
	x |= x >> 16
	return ones32(x) - 1 - uint32(y)
}

func (p *Percentile) intervalIdx(v Mark) (latency int64, idx int) {
	latency = int64(v)
	idx = 0
	if v < 0 {
		latency = 0
		return
	}

	if v <= 2 {
		return
	}
	idx = int(log2(uint32(latency)) - 1)
	return
}
func (p *Percentile) Push(v Mark) {
	latency, idx := p.intervalIdx(v)
	//glog.Infof("latency %v idx: %v", latency, idx)
	interval := p.local.IntervalOf(idx)
	if interval.Full() {
		g := p.value.IntervalOf(idx)
		g.Merge(interval)
		p.value.numAdded += int(interval.numAdded)
		p.local.numAdded -= int(interval.AddedCount())
		//if flagPecentileLog && !g.check() {
		//	buf := &bytes.Buffer{}
		//	buf.WriteString(fmt.Sprintf("mark: %d\n", v))
		//	g.Describe(buf)
		//	glog.Info(buf.String())
		//}
		interval.Clear()
	}
	interval.Add64(latency)
	p.local.numAdded++

	//if flagPecentileLog && !interval.check() {
	//	buf := &bytes.Buffer{}
	//	buf.WriteString(fmt.Sprintf("mark: %d\n", v))
	//	interval.Describe(buf)
	//	glog.Info(buf.String())
	//}
}
func (p *Percentile) Reset() *PercentileSamples {
	ret := p.value.Dup()
	ret.Merge(p.local)
	p.value.Clear()
	p.local.Clear()
	return ret
}

func (p *Percentile) GetValue() *PercentileSamples {
	ret := p.value.Dup()
	ret.Merge(p.local)
	return ret
}

func (p *Percentile) Operators() (op PercentileOperator, invOp PercentileOperator) {
	op = func(left, right *PercentileSamples) {
		left.Merge(right)
	}
	invOp = nil
	return
}

func (p *Percentile) GetWindowSampler() PercentileWinSampler {
	if p.sampler == nil {
		p.sampler = NewPercentileReducerSampler(p)
	}
	return p.sampler
}

func NewPercentile() *Percentile {
	p := &Percentile{
		sampler: nil,
		value:   NewPercentileSamples(GlobalPercentileSamplesSize),
		local:   NewPercentileSamples(ThreadLocalPercentileSamplesSize),
	}
	return p
}
