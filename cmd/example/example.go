package main

import (
	"github.com/forrestjgq/gomark"
	"github.com/forrestjgq/gomark/internal/gm"
	"math/rand"
	"sync"
	"time"
)

func main() {

	gomark.StartHTTPServer(7777)

	wg := sync.WaitGroup{}
	wg.Add(2) // for adder and latency waiting

	// adder
	go func() {
		adder := gomark.NewAdder("test_adder")
		for i := 0; i < 10000; i++ {
			adder.Mark(rand.Int31n(100))
			time.Sleep(100 * time.Millisecond)
		}
		adder.Cancel()
		wg.Done()
	}()

	// latency recorder
	go func() {
		lr := gomark.NewLatencyRecorder("test_latency")
		for i := 0; i < 10000; i++ {
			lr.Mark(rand.Int31n(100))
			time.Sleep(100 * time.Millisecond)
		}
		lr.Cancel()
		wg.Done()
	}()

	go func() {
		time.Sleep(time.Second * 10)
		gm.DisableServer()
		time.Sleep(time.Second * 10)
		gm.EnableServer()

	}()

	// wait and see monitor on http://127.0.0.1:7777/vars
	wg.Wait()
}
