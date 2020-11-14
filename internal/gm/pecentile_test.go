package gm

import (
	"bytes"
	"testing"
)

func TestPercentileAdd(t *testing.T) {
	p := NewPercentile()
	for j := 0; j < 10; j++ {
		for i := 0; i < 10000; i++ {
			p.Push(Mark(i + 1))
		}
		b := p.Reset()
		//if j == 9 {
		buf := &bytes.Buffer{}
		b.Describe(buf)
		t.Log(buf.String())
		//_ = ioutil.WriteFile(t.Name()+"_out.txt", buf.Bytes(), os.ModePerm)
		//}
		lastValue := uint32(0)
		for k := 1; k < 10; k++ {
			value := b.GetNumber(float64(k) / 10.0)
			if value < lastValue {
				t.Fatalf("value: %d lastValue %d", value, lastValue)
			} else {
				t.Logf("value: %d lastValue %d", value, lastValue)
			}
			lastValue = value
			if value <= uint32(k*1000-500) {
				t.Fatalf("f1: k: %d", k)
			}
			if value >= uint32(k*1000+500) {
				t.Fatalf("f2: k: %d", k)
			}
		}

		t.Logf("99%%: %v 99.9%%: %v 99.99%%: %v",
			b.GetNumber(0.99), b.GetNumber(0.999), b.GetNumber(0.9999))

	}
}
