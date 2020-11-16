package main

import (
	"flag"
	"math"
	"math/rand"
	"time"

	"github.com/forrestjgq/gomark/internal/gm"

	"github.com/golang/glog"

	"github.com/forrestjgq/gomark"
)

func testWindowMaxer(total int) {
	glog.Info("Window Maxer")
	wm := gomark.NewWindowMaxer("window_maxer_t")
	for i := 0; i < total; i++{
		v := rand.Int31n(100) + 1
		wm.Mark(v)
		//glog.Infof("window maxer mark %d", v)
		time.Sleep(100 * time.Millisecond)
	}
	wm.Cancel()
	gm.MakeSureEmpty()
}
func testAdder(total int) {
	glog.Info("Adder")
	ad := gomark.NewAdder("hello")
	for i := 0; i < total; i++ {
		v := rand.Int31n(10) + 1
		ad.Mark(v)
		//glog.Infof("mark %d", v)
		time.Sleep(100 * time.Millisecond)
	}
	ad.Cancel()
	gm.MakeSureEmpty()
}
func testCounter(total int) {
	glog.Info("Counter")
	cnt := gomark.NewCounter("hello")
	for i := 0; i < total; i++ {
		v := rand.Int31n(10) + 1
		cnt.Mark(v)
		//glog.Infof("mark %d", v)
		time.Sleep(100 * time.Millisecond)
	}
	cnt.Cancel()
	gm.MakeSureEmpty()
}
func testQPS(total int) {
	glog.Info("QPS")

	qps := gomark.NewQPS("hello")
	for i := 0; i < total; i++ {
		//v := rand.Int31n(10) + 1
		qps.Mark(6)
		//glog.Infof("mark %d", v)
		//time.Sleep(time.Duration(rand.Intn(28)*3+17) * time.Millisecond)
		time.Sleep(100 * time.Millisecond)
	}
	qps.Cancel()
	gm.MakeSureEmpty()
}
func testLatencyRecorder(total int) {
	glog.Info("Latency Recorder")

	lr := gomark.NewLatencyRecorder("hello")
	for i := 0; i < total; i++{
		v := rand.Int31n(100) + 1
		//lr.Mark(70)
		lr.Mark(v)
		//glog.Infof("mark %d", v)
		time.Sleep(time.Duration(rand.Intn(28)*3+17) * time.Millisecond)
	}

	lr.Cancel()
	gm.MakeSureEmpty()
}
func testPecentile(total int) {
	glog.Info("Percentile")

	m := gomark.NewPercentile()
	for i := 0; i < total; i++ {
		v := rand.Int31n(100) + 1
		m.Push(gm.Mark(v))
		//glog.Infof("mark %d", v)
		time.Sleep(time.Duration(rand.Intn(28)*3+17) * time.Millisecond)
	}

	m.Dispose()
	gm.MakeSureEmpty()
}

func main() {
	total := 0
	port := 0
	flag.IntVar(&total, "n", math.MaxInt64 - 1, "how many time for each var to run, not present for infinite")
	flag.IntVar(&port, "p", 7777, "http port, default 7777")
	flag.Parse()
	gomark.StartHTTPServer(port)

	testLatencyRecorder(total)
	testWindowMaxer(total)
	testAdder(total)
	testCounter(total)
	testQPS(total)
	testPecentile(total)
	glog.Info("exit")

	gm.MakeSureEmpty()
}
