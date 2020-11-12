package main

import (
	"log"
	"math/rand"
	"time"

	"github.com/forrestjgq/gomark"
)

func main() {
	gomark.StartHTTPServer(7777)
	lr := gomark.NewLatencyRecorder("hello")
	for {
		v := rand.Int31n(100) + 1
		lr.Mark(v)
		//fmt.Printf("mark %d\n", v)
		time.Sleep(1000 * time.Microsecond)
	}
	log.Print("exit")
}
