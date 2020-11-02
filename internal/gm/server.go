package gm

import "sync"

const (
	sizeOfQ = 10000
)

type server struct {
	seq Identity
	c   chan stub
	wg  sync.WaitGroup
	all map[Identity]Variable
}

var srv *server

func init() {
	srv = &server{
		c:   make(chan stub, sizeOfQ),
		wg:  sync.WaitGroup{},
		all: make(map[Identity]Variable),
	}
	go srv.run()
}
func (s *server) unlock() {
	s.wg.Done()
}
func (s *server) lock() interface {
	unlock()
} {
	s.wg.Add(1)
	s.c <- makeStub(cmdLock, 0, 0)
	return s
}

func (s *server) run() {
	for rx := range s.c {
		switch rx.cmd() {
		case cmdLock:
			s.wg.Wait()
		case cmdNew:
			s.newStub(rx)
		case cmdCancel:
			s.removeStub(rx)
		case cmdMark:
			s.markStub(rx)
		}
	}
}
func (s *server) newStub(stub stub) {
}
func (s *server) removeStub(stub stub) {
}
func (s *server) markStub(stub stub) {
}

func AddVariable(v Variable) Identity {
	defer srv.lock().unlock()

	var id Identity
	for {
		srv.seq++
		if _, ok := srv.all[srv.seq]; !ok && srv.seq != 0 {
			id = srv.seq
			break
		}
	}
	srv.all[id] = v
	return id
}
func RemoveVariable(id Identity) {
	defer srv.lock().unlock()

	delete(srv.all, id)
}

func PushStub(s stub) {
	srv.c <- s
}
