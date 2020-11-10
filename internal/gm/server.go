package gm

import (
	"strconv"
	"sync"
	"unicode"
)

const (
	sizeOfQ = 10000
)

type server struct {
	seq   Identity
	stubc chan stub
	callc chan func()
	wg    sync.WaitGroup
	all   map[Identity]Variable
}

var srv *server

func init() {
	srv = &server{
		stubc: make(chan stub, sizeOfQ),
		callc: make(chan func()),
		wg:    sync.WaitGroup{},
		all:   make(map[Identity]Variable),
	}
	go srv.run()
}
func (s *server) Unlock1() {
	s.wg.Done()
}
func Lock() interface {
	Unlock1()
} {
	srv.wg.Add(1)
	srv.stubc <- makeStub(cmdLock, 0, 0)
	return srv
}

func (s *server) run() {
	for rx := range s.stubc {
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
func (s *server) remove(id Identity) {

}
func (s *server) add(v Variable) Identity {
	for {
		srv.seq++
		if _, ok := srv.all[srv.seq]; !ok && srv.seq != 0 {
			return srv.seq
		}
	}
}

func toUnderScore(name, src string) string {
	var last rune
	for i, c := range src {
		if unicode.IsLetter(c) {
			if unicode.IsUpper(c) {
				if i != 0 && last == 0 && name[len(name)-1] != '_' {
					name += "_"
				}
				name += strconv.QuoteRune(unicode.ToLower(c))
			} else {
				name += strconv.QuoteRune(c)
			}
		} else if unicode.IsDigit(c) {
			name += strconv.QuoteRune(c)
		} else if len(name) == 0 || name[len(name)-1] != '_' {
			name += "_"
		}

		last = c
	}
	return name
}
func Expose(prefix, name string, displayFilter DisplayFilter, v Variable) error {
	vb := &VarBase{}
	realName := ""
	if len(prefix) > 0 {
		realName = toUnderScore(realName, prefix)
		if len(realName) > 0 && realName[len(realName)-1] != '_' {
			realName += "_"
		}
	}
	vb.name = toUnderScore(realName, name)
	for {
		srv.seq++
		if _, ok := srv.all[srv.seq]; !ok && srv.seq != 0 {
			vb.id = srv.seq
			break
		}
	}
	vb.displayFilter = displayFilter
	return v.OnExpose(vb)
}
func callFunc(call func()) {
	var wg sync.WaitGroup
	wg.Add(1)
	srv.callc <- func() {
		call()
		wg.Done()
	}
	wg.Wait()
}
func AddVariable(prefix, name string, displayFilter DisplayFilter, v Variable) error {
	var err error
	f := func() {
		err = Expose(prefix, name, displayFilter, v)
	}
	callFunc(f)
	return err
}
func RemoveVariable(id Identity) {
	f := func() {
		srv.remove(id)
	}
	callFunc(f)
}
func RemoveVariables(ids ...Identity) {
	f := func() {
		for _, id := range ids {
			srv.remove(id)
		}
	}
	callFunc(f)
}

func PushStub(s stub) {
	srv.stubc <- s
}
