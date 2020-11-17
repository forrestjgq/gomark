package gm

import (
	"bytes"
	"math"
	"sync"
	"testing"
	"time"

	"github.com/golang/glog"
)

func TestIntRecorder_Sanity(t *testing.T) {
	ir, err := NewIntRecorderWithName("var")
	if err != nil {
		t.Fatalf("fail to create int recorder")
	}
	for i := 0; i < 100; i++ {
		ir.VarBase().Mark(2)
	}

	// wait for processing
	time.Sleep(100 * time.Millisecond)

	if ir.IntAverage() != 2 {
		t.Fatalf("average failure")
	}

	buf := &bytes.Buffer{}
	ir.Describe(buf, false)
	if buf.String() != "2" {
		t.Fatalf("describe fail")
	}

	ir.VarBase().Cancel()
	MakeSureEmpty()
}
func TestIntRecorder_Negative(t *testing.T) {
	ir, err := NewIntRecorderNoExpose()
	if err != nil {
		t.Fatalf("fail to create int recorder")
	}
	for i := 0; i < 100; i++ {
		ir.Push(-2)
	}

	if ir.IntAverage() != -2 {
		t.Fatalf("average failure")
	}
	MakeSureEmpty()
}
func TestIntRecorder_PositiveOverflow(t *testing.T) {
	ir1, err := NewIntRecorderNoExpose()
	if err != nil {
		t.Fatalf("fail to create int recorder")
	}
	for i := 0; i < math.MaxInt32*100; i++ {
		ir1.Push(math.MaxInt32)
		v := ir1.value
		if v.x < 0 || v.y < 0 {
			t.Fatalf("overflow")
		}
		if v.x == 0 {
			if v.y != 0 {
				t.Fatalf("overflow but now reset")
			}
			break
		}
		if ir1.IntAverage() != math.MaxInt32 {
			t.Fatalf("average failure")
		}
	}

	MakeSureEmpty()
}
func TestIntRecorder_NegativeOverflow(t *testing.T) {
	ir1, err := NewIntRecorderNoExpose()
	if err != nil {
		t.Fatalf("fail to create int recorder")
	}
	for i := 0; i < math.MaxInt32*100; i++ {
		ir1.Push(math.MinInt32)
		v := ir1.value
		if v.x > 0 || v.y < 0 {
			t.Fatalf("overflow")
		}
		if v.x == 0 {
			if v.y != 0 {
				t.Fatalf("overflow but now reset")
			}
			break
		}
		if ir1.IntAverage() != math.MinInt32 {
			t.Fatalf("average failure")
		}
	}

	MakeSureEmpty()
}
func TestIntRecorder_Perf(t *testing.T) {
	ir1, err := NewIntRecorderWithName("var")
	if err != nil {
		t.Fatalf("fail to create int recorder")
	}
	n := 20
	qps := 20000000
	wg := sync.WaitGroup{}
	wg.Add(n)
	start := time.Now()
	for i := 0; i < n; i++ {
		go func() {
			for k := 0; k < qps; k++ {
				ir1.VarBase().Mark(int32(k))
			}
			RemoteCall(func() {
				wg.Done()
			})
		}()
	}
	wg.Wait()
	du := time.Since(start)
	glog.Infof("%d routines, each push %d value, takes %v", n, qps, du)

	ir1.VarBase().Cancel()
	MakeSureEmpty()
}
