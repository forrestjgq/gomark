package main

import (
	"flag"
	"math"
	"math/rand"
	"time"

	"github.com/nsf/termbox-go"

	"github.com/forrestjgq/gomark/internal/gm"

	"github.com/golang/glog"

	"github.com/forrestjgq/gomark"
)

func init() {
	if err := termbox.Init(); err != nil {
		panic(err)
	}
	termbox.SetCursor(0, 0)
	termbox.HideCursor()
}
func testMaxWindow(total int) {
	glog.Info("Maxer window")
	ad := gomark.NewWindowMaxer("hello")
	for i := 0; i < total; i++ {
		v := rand.Int31n(10) + 1
		ad.Mark(v)
		//glog.Infof("mark %d", v)
		time.Sleep(100 * time.Millisecond)
	}
	ad.Cancel()
	gm.MakeSureEmpty()
}
func testMaxer(total int) {
	glog.Info("Maxer")
	ad := gomark.NewMaxer("hello")
	for i := 0; i < total; i++ {
		v := rand.Int31n(10) + 1
		ad.Mark(v)
		//glog.Infof("mark %d", v)
		time.Sleep(100 * time.Millisecond)
	}
	ad.Cancel()
	gm.MakeSureEmpty()
}
func testAdder(total int) {
	glog.Info("Adder")
	ad := gomark.NewAdder("hello")
	for i := 0; i < total; i++ {
		v := rand.Int31n(11) - 5
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
	for i := 0; i < total; i++ {
		v := rand.Int31n(100) + 1
		//lr.Mark(70)
		lr.Mark(v)
		glog.Infof("mark %d", v)
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

var stop = false

func wait(test string) {
	if !stop {
		return
	}

	glog.Info("press any key to start ", test)
Loop:
	for {
		switch ev := termbox.PollEvent(); ev.Type {
		case termbox.EventKey:
			break Loop
		}
	}
}
func main() {
	total := 0
	port := 0
	flag.BoolVar(&stop, "s", false, "set to true to test step by step")
	flag.IntVar(&total, "n", math.MaxInt64-1, "how many time for each var to run, not present for infinite")
	flag.IntVar(&port, "p", 7770, "http port, default 7770")
	flag.Parse()
	gomark.StartHTTPServer(port)

	wait("testMaxWindow")
	testMaxWindow(total)
	wait("testMaxer")
	testMaxer(total)
	wait("testAdder")
	testAdder(total)
	wait("testLatencyRecorder")
	testLatencyRecorder(total)
	wait("testQPS")
	testQPS(total)
	wait("testCounter")
	testCounter(total)
	wait("testPercentile")
	testPecentile(total)
	glog.Info("exit")

	gm.MakeSureEmpty()
}
