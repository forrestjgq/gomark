package gomark_test

import (
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/forrestjgq/gomark"
)

const (
	port = 7777
)

func getMetrics(t *testing.T) string {
	c := http.Client{}
	rsp, err := c.Get("http://127.0.0.1:" + strconv.Itoa(port) + "/metrics")
	if err != nil {
		t.Fatalf(err.Error())
	}

	if rsp.StatusCode != 200 {
		t.Fatalf("status code %d", rsp.StatusCode)
	}

	if !strings.HasPrefix(rsp.Header.Get("Content-Type"), "text/plain") {
		t.Fatalf("invalid content type %s", rsp.Header.Get("Content-Type"))
	}

	b, err := ioutil.ReadAll(rsp.Body)
	if err != nil {
		t.Fatalf(err.Error())
	}

	return string(b)
}
func TestMetrics(t *testing.T) {
	wg := sync.WaitGroup{}

	lr := gomark.NewLatencyRecorder("t_latency_l_model_facedetect_inference")
	wg.Add(1)
	go func() {
		for i := 0; i < 100; i++ {
			m := (i+11)*2 + 13
			lr.Mark(int32(m))
			time.Sleep(40 * time.Millisecond)
		}
		wg.Done()
	}()

	ad := gomark.NewAdder("t_gauge_l_model_facedetect_inference")
	wg.Add(1)
	go func() {
		for i := 0; i < 100; i++ {
			ad.Mark(int32(i))
			time.Sleep(40 * time.Millisecond)
		}
		wg.Done()
	}()

	wg.Wait()

	c := gomark.NewCounter("test_counter_mark")
	wg.Add(1)
	go func() {
		for i := 0; i < 100; i++ {
			c.Mark(int32(i))
			time.Sleep(40 * time.Millisecond)
		}
		wg.Done()
	}()

	max := gomark.NewMaxer("test_maxer_mark")
	wg.Add(1)
	go func() {
		for i := 0; i < 100; i++ {
			max.Mark(int32(i))
			time.Sleep(40 * time.Millisecond)
		}
		wg.Done()
	}()

	ad1 := gomark.NewAdder("t_gauge_l_model_vehicledetect_inference")
	wg.Add(1)
	go func() {
		for i := 0; i < 100; i++ {
			ad1.Mark(int32(i))
			time.Sleep(40 * time.Millisecond)
		}
		wg.Done()
	}()

	st := gomark.NewStatus("teststatus")
	wg.Add(1)
	go func() {
		for i := 0; i < 100; i++ {
			st.Mark(int32(i))
			time.Sleep(40 * time.Millisecond)
		}
		wg.Done()
	}()

	tst := gomark.NewStatus("t_teststatus")
	wg.Add(1)
	go func() {
		for i := 0; i < 100; i++ {
			tst.Mark(int32(i))
			time.Sleep(40 * time.Millisecond)
		}
		wg.Done()
	}()

	wg.Wait()
	time.Sleep(1 * time.Second)
	body := getMetrics(t)
	t.Logf("Metrics\n%s", body)

}

func TestMain(m *testing.M) {
	gomark.StartHTTPServer(port)
	time.Sleep(3 * time.Second)
	m.Run()
}
