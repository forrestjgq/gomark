package main

import (
	"bufio"
	"flag"
	"fmt"
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

var mode = ""

func wait(test string) {
	if mode != "step" {
		return
	}
	glog.Info("Press 'Enter' to continue ", test)
	_, _ = bufio.NewReader(os.Stdin).ReadBytes('\n')
}
func pause() {
	if mode != "perf" {
		time.Sleep(time.Duration(rand.Intn(28)*3+17) * time.Millisecond)
		//time.Sleep(time.Second)
	}
}

func reportPerf(name string, start time.Time, total int) {
	if mode == "perf" {
		du := time.Since(start).Seconds()
		qps := int(float64(total) / du)
		glog.Infof("%s qps: %d", name, qps)
	}
}

func testMaxWindow(name string, total int) {
	wait(name)
	ad := gomark.NewWindowMaxer(name)
	start := time.Now()
	for i := 0; i < total; i++ {
		v := rand.Int31n(10) + 1
		ad.Mark(v)
		pause()
	}
	reportPerf(name, start, total)
	ad.Cancel()
	gm.MakeSureEmpty()
}
func testMaxer(name string, total int) {
	wait(name)
	ad := gomark.NewMaxer(name)
	start := time.Now()
	for i := 0; i < total; i++ {
		v := rand.Int31n(10) + 1
		ad.Mark(v)
		//glog.Infof("mark %d", v)
		pause()
	}
	reportPerf(name, start, total)
	ad.Cancel()
	gm.MakeSureEmpty()
}
func testAdder(name string, total int) {
	wait(name)
	ad := gomark.NewAdder(name)
	start := time.Now()
	for i := 0; i < total; i++ {
		v := rand.Int31n(11) - 5
		ad.Mark(v)
		//glog.Infof("mark %d", v)
		pause()
	}
	reportPerf(name, start, total)
	ad.Cancel()
	gm.MakeSureEmpty()
}
func testAdderPerSecond(name string, total int) {
	wait(name)
	ad := gomark.NewAdderPerSecond(name)
	start := time.Now()
	for i := 0; i < total; i++ {
		//v := rand.Int31n(11) - 5
		ad.Mark(10)
		//glog.Infof("mark %d", v)
		pause()
	}
	reportPerf(name, start, total)
	ad.Cancel()
	gm.MakeSureEmpty()
}
func testStatus(name string, total int) {
	wait(name)
	ad := gomark.NewStatus(name)
	start := time.Now()
	for i := 0; i < total; i++ {
		//v := rand.Int31n(11) - 5
		ad.Mark(int32(i % 10))
		//glog.Infof("mark %d", v)
		pause()
	}
	reportPerf(name, start, total)
	ad.Cancel()
	gm.MakeSureEmpty()
}
func testCounter(name string, total int) {
	wait(name)
	cnt := gomark.NewCounter(name)
	start := time.Now()
	for i := 0; i < total; i++ {
		v := rand.Int31n(10) + 1
		cnt.Mark(v)
		//glog.Infof("mark %d", v)
		pause()
	}
	reportPerf(name, start, total)
	cnt.Cancel()
	gm.MakeSureEmpty()
}
func testQPS(name string, total int) {
	wait(name)

	qps := gomark.NewQPS(name)
	start := time.Now()
	for i := 0; i < total; i++ {
		//v := rand.Int31n(10) + 1
		qps.Mark(6)
		//glog.Infof("mark %d", v)
		//time.Sleep(time.Duration(rand.Intn(28)*3+17) * time.Millisecond)
		pause()
	}
	reportPerf(name, start, total)
	qps.Cancel()
	gm.MakeSureEmpty()
}
func testLatencyRecorder(name string, total int) {
	wait(name)

	lr := gomark.NewLatencyRecorder(name)
	start := time.Now()
	for i := 0; i < total; i++ {
		v := rand.Int31n(100) + 1
		//lr.Mark(70)
		lr.Mark(v)
		//glog.Infof("mark %d", v)
		pause()
	}

	reportPerf(name, start, total)
	lr.Cancel()
	gm.MakeSureEmpty()
}
func testPecentile(total int) {
	wait("percentile")

	m := gomark.NewPercentile()
	for i := 0; i < total; i++ {
		v := rand.Int31n(100) + 1
		m.Push(gm.Mark(v))
		//glog.Infof("mark %d", v)
		pause()
	}

	m.Dispose()
	gm.MakeSureEmpty()
}

func usage() {
	str := `
gomark -p <port> -m <mode> -n <number>
<port>: http server port
<mode>: test mode
        - walk: test one by one, each runs <number> marks, and sleep a litter while after each mark.
        - step: just like walk, but it pauses on starting of each test
        - perf: runs <number> marks and do not sleep after mark. Then calculate QPS of this variable.
        - forever: start <number> goroutines and each runs all supported variables forever.
`
	fmt.Println(str)
}
func main() {
	total := 0
	port := 0
	flag.StringVar(&mode, "m", "", "test mode: walk,perf,step,forever")
	flag.IntVar(&total, "n", 100, "how many time for each var to run, or how many batches in forever mode")
	flag.IntVar(&port, "p", 7770, "http port, default 7770")
	flag.Parse()

	gomark.StartHTTPServer(port)

	switch mode {
	case "walk":
	case "step":
	case "perf":
	case "forever":
		glog.Info("")
		wg := sync.WaitGroup{}
		wg.Add(1)
		gm.EnableInternalVariables()

		for i := 0; i < total; i++ {
			go testStatus("testStatus_"+strconv.Itoa(i), math.MaxInt64)
			go testAdderPerSecond("testAdderPerSecond_"+strconv.Itoa(i), math.MaxInt64)
			go testMaxWindow("max_win_"+strconv.Itoa(i), math.MaxInt64)
			go testMaxer("max_"+strconv.Itoa(i), math.MaxInt64)
			go testAdder("adder_"+strconv.Itoa(i), math.MaxInt64)
			go testLatencyRecorder("latency_recorder_"+strconv.Itoa(i), math.MaxInt64)
			go testQPS("qps_"+strconv.Itoa(i), math.MaxInt64)
			go testCounter("counter_"+strconv.Itoa(i), math.MaxInt64)
		}

		wg.Wait()
	default:
		glog.Info("unknown mode: ", mode)
		usage()
		os.Exit(1)
	}

	if total == 0 {
		total = 100
	}

	glog.Infof("mode: %s total %d", mode, total)

	testAdderPerSecond("testAdderPerSecond", total)
	testStatus("testStatus", total)
	testMaxWindow("testMaxWindow", total)
	testMaxer("testMaxer", total)
	testAdder("testAdder", total)
	testLatencyRecorder("testLatencyRecorder", total)
	testQPS("testQPS", total)
	testCounter("testCounter", total)
	testPecentile(total)
	gm.MakeSureEmpty()

	glog.Info("exit")
}
