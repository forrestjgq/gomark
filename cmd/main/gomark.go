package main

import (
	"bufio"
	"flag"
	"math"
	"math/rand"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/forrestjgq/gomark"
	"github.com/forrestjgq/gomark/internal/gm"
	"github.com/golang/glog"
)

func testMaxWindow(name string, total int) {
	glog.Info("Maxer window")
	ad := gomark.NewWindowMaxer(name)
	for i := 0; i < total; i++ {
		v := rand.Int31n(10) + 1
		ad.Mark(v)
		//glog.Infof("mark %d", v)
		time.Sleep(100 * time.Millisecond)
	}
	ad.Cancel()
	gm.MakeSureEmpty()
}
func testMaxer(name string, total int) {
	glog.Info("Maxer")
	ad := gomark.NewMaxer(name)
	for i := 0; i < total; i++ {
		v := rand.Int31n(10) + 1
		ad.Mark(v)
		//glog.Infof("mark %d", v)
		time.Sleep(100 * time.Millisecond)
	}
	ad.Cancel()
	gm.MakeSureEmpty()
}
func testAdder(name string, total int) {
	glog.Info("Adder")
	ad := gomark.NewAdder(name)
	for i := 0; i < total; i++ {
		v := rand.Int31n(11) - 5
		ad.Mark(v)
		//glog.Infof("mark %d", v)
		time.Sleep(100 * time.Millisecond)
	}
	ad.Cancel()
	gm.MakeSureEmpty()
}
func testCounter(name string, total int) {
	glog.Info("Counter")
	cnt := gomark.NewCounter(name)
	for i := 0; i < total; i++ {
		v := rand.Int31n(10) + 1
		cnt.Mark(v)
		//glog.Infof("mark %d", v)
		time.Sleep(100 * time.Millisecond)
	}
	cnt.Cancel()
	gm.MakeSureEmpty()
}
func testQPS(name string, total int) {
	glog.Info("QPS")

	qps := gomark.NewQPS(name)
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
func testLatencyRecorder(name string, total int) {
	glog.Info("Latency Recorder")

	lr := gomark.NewLatencyRecorder(name)
	for i := 0; i < total; i++ {
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

var stop = false

func wait(test string) {
	if !stop {
		return
	}
	glog.Info("Press 'Enter' to continue ", test)
	_, _ = bufio.NewReader(os.Stdin).ReadBytes('\n')
}
func main() {
	total := 0
	port := 0
	flag.BoolVar(&stop, "s", false, "set to true to test step by step")
	flag.IntVar(&total, "n", math.MaxInt64-1, "how many time for each var to run, not present for infinite")
	flag.IntVar(&port, "p", 7770, "http port, default 7770")
	flag.Parse()
	gomark.StartHTTPServer(port)

	if total == 0 {
		total = math.MaxInt64 - 1
		wg := sync.WaitGroup{}
		wg.Add(1)
		gm.EnableInternalVariables()

		for i := 0; i < 1000; i++ {
			go testMaxWindow("max_win_"+strconv.Itoa(i), total)
			go testMaxer("max_"+strconv.Itoa(i), total)
			go testAdder("adder_"+strconv.Itoa(i), total)
			go testLatencyRecorder("latency_recorder_"+strconv.Itoa(i), total)
			go testQPS("qps_"+strconv.Itoa(i), total)
			go testCounter("counter_"+strconv.Itoa(i), total)
		}

		go testPecentile(total)
		wg.Wait()
	} else {
		wait("testMaxWindow")
		testMaxWindow("testMaxWindow", total)
		wait("testMaxer")
		testMaxer("testMaxer", total)
		wait("testAdder")
		testAdder("testAdder", total)
		wait("testLatencyRecorder")
		testLatencyRecorder("testLatencyRecorder", total)
		wait("testQPS")
		testQPS("testQPS", total)
		wait("testCounter")
		testCounter("testCounter", total)
		wait("testPercentile")
		testPecentile(total)
		glog.Info("exit")

		gm.MakeSureEmpty()
	}
}
