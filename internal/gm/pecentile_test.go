package gm

import (
	"math"
	"testing"
)

func TestPercentileAdd(t *testing.T) {
	p := NewPercentile()
	for j := 0; j < 10; j++ {
		for i := 0; i < 10000; i++ {
			p.Push(Mark(i + 1))
		}
		b := p.Reset()
		//buf := &bytes.Buffer{}
		//buf.WriteString("before get num:")
		//b.Describe(buf)
		//t.Log(buf.String())
		lastValue := uint32(0)
		for k := 1; k < 10; k++ {
			value := b.GetNumber(float64(k) / 10.0)
			if value < lastValue {
				t.Fatalf("value: %d lastValue %d", value, lastValue)
			}
			lastValue = value
			if value <= uint32(k*1000-500) {
				//tb := &bytes.Buffer{}
				//tb.WriteString("after 1 get num:")
				//b.Describe(tb)
				//t.Log(tb.String())
				t.Fatalf("f1: k: %d", k)
			}
			if value >= uint32(k*1000+500) {
				//tb := &bytes.Buffer{}
				//tb.WriteString("after 2 get num:")
				//b.Describe(tb)
				//t.Log(tb.String())
				t.Fatalf("f2: k: %d", k)
			}
		}

		t.Logf("99%%: %v 99.9%%: %v 99.99%%: %v",
			b.GetNumber(0.99), b.GetNumber(0.999), b.GetNumber(0.9999))

	}
}
func TestPercentileMerge1(t *testing.T) {
	const (
		N = 1000
		SampleSize = 32
	)
	belongToB1, belongToB2 := 0, 0
	for repeat := 0; repeat < 100; repeat++ {
		b0 := newPercentileInterval(SampleSize*3)
		b1 := newPercentileInterval(SampleSize)
		for i := 0; i < N; i++ {
			if b1.Full() {
				b0.Merge(b1)
				b1.Clear()
			}
			if !b1.Add32(uint32(i)) {
				t.Fatalf("1 repeat: %d i %d", repeat, i)
			}
		}
		b0.Merge(b1)
		b2 := newPercentileInterval(SampleSize*2)
		for i := 0; i < N*2; i++ {
			if b2.Full() {
				b0.Merge(b2)
				b2.Clear()
			}
			if !b2.Add32(uint32(i+N)) {
				t.Fatalf("2 repeat: %d i %d", repeat, i)
			}
		}
		b0.Merge(b2)
		for i := 0; i < int(b0.numSamples); i++ {
			if int(b0.samples[i]) < N {
				belongToB1++
			} else {
				belongToB2++
			}
		}
	}
	f := float64(belongToB1) / float64(belongToB2) - 0.5
	if math.Abs(f) >= 0.2 {
		t.Fatalf("belong b1 : %d , b2: %d", belongToB1, belongToB2)
	}
}

func TestPercentileMerge2(t *testing.T) {
	const (
		n1 = 1000
		n2 = 400
	)
	belongToB1, belongToB2 := 0, 0

	for repeat := 0; repeat < 100; repeat++ {
		b0 := newPercentileInterval(64)
		b1 := newPercentileInterval(64)
		for i := 0; i < n1; i++ {
			if b1.Full() {
				b0.Merge(b1)
				b1.Clear()
			}
			if !b1.Add32(uint32(i)) {
				t.Fatalf("1 repeat: %d i %d", repeat, i)
			}
		}
		b0.Merge(b1)
		b2 := newPercentileInterval(64)
		for i := 0; i < n2; i++ {
			if b2.Full() {
				b0.Merge(b2)
				b2.Clear()
			}
			if !b2.Add32(uint32(i+n1)) {
				t.Fatalf("2 repeat: %d i %d", repeat, i)
			}
		}
		b0.Merge(b2)
		for i := 0; i < int(b0.numSamples); i++ {
			if int(b0.samples[i]) < n1 {
				belongToB1++
			} else {
				belongToB2++
			}
		}
	}
	f := float64(belongToB1) / float64(belongToB2) - float64(n1) / float64(n2)
	if math.Abs(f) >= 0.2 {
		t.Fatalf("belong b1 : %d , b2: %d", belongToB1, belongToB2)
	}
}

func TestPercentileSamples_CombineOf(t *testing.T) {
	const (
		numSamplers = 10
		base = uint32(1<<30 + 1)
		N = 1000
	)

	belongs := make([]int, numSamplers)
	total := 0

	for repeat := 0; repeat < 100; repeat++ {
		p := make([]*Percentile, numSamplers)
		result := make([]*PercentileSamples, numSamplers)
		for i := 0; i < numSamplers; i++ {
			p[i] = NewPercentile()
			for j := 0; j < N * (i+1); j++ {
				p[i].Push(Mark(base + uint32(i *(i+1) * N/2+j)))
			}
			result[i] = p[i].GetValue()
		}
		g := NewPercentileSamples(510)
		g.CombineOf(result)
		for i := 0; i < NumIntervals; i++ {
			if g.intervals[i] == nil {
				continue
			}
			p := g.intervals[i]
			total += int(p.numSamples)
			for j:=0; j < int(p.numSamples); j++ {
				for k:=0; k < int(numSamplers) ; k++ {
					if int(p.samples[j] - base) /N < (k+1) * (k+2) / 2 {
						belongs[k]++
						break
					}
				}
			}
		}
	}

	for i := 0; i < numSamplers; i++ {
		expectRatio := float64(i+1) * 2.0 / float64(numSamplers*(numSamplers+1))
		actualRatio := float64(belongs[i]) / float64(total)
		if math.Abs(expectRatio - actualRatio) - 1.0 >= 0.2 {
			t.Fatalf("ratio expect %v actual %v", expectRatio, actualRatio)
		}
	}
}