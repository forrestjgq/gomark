package main

import (
	"flag"
	"log"
	"math/rand"
	"time"

	"github.com/forrestjgq/gomark/internal/gm"

	"github.com/golang/glog"

	"github.com/forrestjgq/gomark"
)

func testWindowMaxer() {
	wm := gomark.NewWindowMaxer("window_maxer_t")
	for {
		v := rand.Int31n(100) + 1
		wm.Mark(v)
		//glog.Infof("window maxer mark %d", v)
		time.Sleep(100 * time.Millisecond)
	}
}
func testAdder() {
	ad := gomark.NewAdder("hello")
	for {
		v := rand.Int31n(10) + 1
		ad.Mark(v)
		glog.Infof("mark %d", v)
		time.Sleep(1000 * time.Millisecond)
	}
}
func testCounter() {
	cnt := gomark.NewCounter("hello")
	for {
		v := rand.Int31n(10) + 1
		cnt.Mark(v)
		//glog.Infof("mark %d", v)
		time.Sleep(1000 * time.Millisecond)
	}
}
func testQPS() {

	cnt := gomark.NewQPS("hello")
	for {
		//v := rand.Int31n(10) + 1
		cnt.Mark(6)
		//glog.Infof("mark %d", v)
		//time.Sleep(time.Duration(rand.Intn(28)*3+17) * time.Millisecond)
		time.Sleep(100 * time.Millisecond)
	}
}
func testLatencyRecorder() {

	lr := gomark.NewLatencyRecorder("hello")
	for {
		v := rand.Int31n(100) + 1
		//lr.Mark(70)
		lr.Mark(v)
		//glog.Infof("mark %d", v)
		time.Sleep(time.Duration(rand.Intn(28)*3+17) * time.Millisecond)
	}
}
func testPecentile() {
	m := gomark.NewPercentile()
	for {
		v := rand.Int31n(100) + 1
		m.Push(gm.Mark(v))
		//glog.Infof("mark %d", v)
		time.Sleep(time.Duration(rand.Intn(28)*3+17) * time.Millisecond)
	}
}

func main() {
	flag.Parse()
	gomark.StartHTTPServer(7777)

	testLatencyRecorder()
	log.Print("exit")
}
