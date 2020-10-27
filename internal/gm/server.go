package gm

import "sync"

const (
	sizeOfQ = 10000
)

type server struct {
	seq identity
	c   chan stub
	wg  sync.WaitGroup
	all map[identity]Variable
}

var srv *server

func init() {
	srv = &server{
		c:   make(chan stub, sizeOfQ),
		wg:  sync.WaitGroup{},
		all: make(map[identity]Variable),
	}
	go srv.run()
}
func (s *server) unlock() {
	s.wg.Done()
}
func (s *server) run() {
	for rx := range s.c {
		if rx.cmd() == cmdLock {
			s.wg.Wait()
			continue
		}
	}
}
func (s *server) newIdentity() identity {
	for {
		s.seq++
		if _, ok := srv.all[s.seq]; !ok && s.seq != 0 {
			return s.seq
		}
	}
}
func AddVariable(v Variable) identity {
	defer lock().unlock()

	id := srv.newIdentity()
	srv.all[id] = v
	return id
}
func RemoveVariable(id identity) {
	defer lock().unlock()

	delete(srv.all, id)
}

func PushStub(s stub) {
	srv.c <- s
}

func lock() interface {
	unlock()
} {
	srv.wg.Add(1)
	srv.c <- makeStub(cmdLock, 0, 0)
	return srv
}
