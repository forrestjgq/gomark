package main

import (
	"flag"
	"log"
	"math/rand"
	"time"

	"github.com/forrestjgq/gomark"
)

func main() {
	flag.Parse()
	gomark.StartHTTPServer(7777)
	//wm := gomark.NewWindowMaxer("window_maxer_t")
	//for {
	//	v := rand.Int31n(100) + 1
	//	wm.Mark(v)
	//	glog.Infof("window maxer mark %d", v)
	//	time.Sleep(100 * time.Millisecond)
	//}
	//ad := gomark.NewAdder("hello")
	//for {
	//	v := rand.Int31n(10) + 1
	//	ad.Mark(v)
	//	glog.Infof("mark %d", v)
	//	time.Sleep(1000 * time.Millisecond)
	//}

	//cnt := gomark.NewCounter("hello")
	//for {
	//	v := rand.Int31n(10) + 1
	//	cnt.Mark(v)
	//	//glog.Infof("mark %d", v)
	//	time.Sleep(1000 * time.Millisecond)
	//}

	lr := gomark.NewLatencyRecorder("hello")
	for {
		v := rand.Int31n(100) + 1
		lr.Mark(v)
		//glog.Infof("mark %d", v)
		time.Sleep(time.Duration(rand.Intn(28)*3+17) * time.Millisecond)
	}
	log.Print("exit")
}
