package gm

import (
	"fmt"
	"log"
	"reflect"
	"sort"

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
			log.Fatalf("numAdded %v rhs numAdded %v numSamples %v rhs numSamples %v sampleSize %v rhs sampleSize %v\n",
				pi.numAdded, rhs.numAdded, pi.numSamples, rhs.numSamples, pi.sampleSize, rhs.sampleSize)
		}
		copy(pi.samples[int(pi.numSamples):int(pi.numSamples+rhs.numSamples)], rhs.samples[0:rhs.numSamples])
		pi.numSamples += rhs.numSamples
	} else {
		numRemain := roundOfExpectation(uint(pi.numAdded)*uint(pi.sampleSize), uint(pi.numAdded)+uint(rhs.numAdded))
		if numRemain > uint(pi.numSamples) {
			log.Fatalf("remain: %v samples %v\n", numRemain, pi.numSamples)
		}
		for i := uint(pi.numSamples); i > numRemain; i-- {
			r := fr.Uint32n(uint32(i))
			pi.samples[int(r)] = pi.samples[int(i-1)]
		}

		numRemainFromRhs := uint(pi.sampleSize) - numRemain
		if numRemainFromRhs > uint(rhs.numSamples) {
			log.Fatalf("remian from rhs %v num samples %v\n", numRemainFromRhs, rhs.numSamples)
		}

		tmp := make([]uint32, rhs.numSamples)
		copy(tmp, rhs.samples[0:rhs.numSamples])
		for i := uint(0); i < numRemainFromRhs; i++ {
			idx := fr.Uint32n(uint32(rhs.numSamples) - uint32(i))
			pi.samples[numRemain] = tmp[idx]
			numRemain++
			tmp[idx] = tmp[uint(rhs.numSamples)-i-1]
		}
		pi.numSamples += uint16(numRemain)
		if int(pi.numSamples) != pi.sampleSize {
			log.Fatalf("numSamples %v sampleSize %v\n", pi.numSamples, pi.sampleSize)
		}
	}
	pi.numAdded += rhs.numAdded
}
func (pi *PercentileInterval) MergeWithExpectation(rhs *PercentileInterval, n uint16) {
	if n > rhs.numSamples {
		log.Fatalf("n %v rhs.numSamples %v\n", n, rhs.numSamples)
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
		fmt.Println("this interval was full")
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

const (
	NumIntervals = 32
)

func newPercentileInterval(sampleSize int) *PercentileInterval {
	return &PercentileInterval{
		sampleSize: sampleSize,
		samples:    make([]uint32, sampleSize),
	}
}

type Percentile struct {
}
