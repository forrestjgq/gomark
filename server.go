package gomark

import "sync"

const (
	sizeOfQ = 10000
)

type server struct {
	c chan stub
	wg sync.WaitGroup
}

var srv server

func init() {
	srv.c = make(chan stub, sizeOfQ)
}
func (s *server) lock() interface{
	unlock()
}  {
	makeStub(cmdLock, 0, 0).send()
	return s
}
func (s *server) unlock()  {

}
func (s *server) run()  {
	for rx := range s.c {
		if rx.cmd() == cmdLock {
			s.wg.Add(1)
			s.wg.Wait()
			continue
		}
	}
}
func StartServer(port int)  {
	go srv.run()
}