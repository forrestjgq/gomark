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
	all   map[Identity]Variable
}

var srv *server

func init() {
	srv = &server{
		stubc: make(chan stub, sizeOfQ),
		callc: make(chan func()),
		all:   make(map[Identity]Variable),
	}
	go srv.run()
}

func (s *server) run() {
	for {
		select {
		case f := <-s.callc:
			f()
		case rx := <-s.stubc:
			s.markStub(rx)
		}
	}
}
func (s *server) markStub(stub stub) {
	if v, ok := srv.all[stub.identity()]; ok {
		v.Push(stub.mark())
	}
}
func (s *server) remove(id Identity) {
	if v, ok := srv.all[id]; ok {
		ids := v.Dispose()
		for _, sub := range ids {
			s.remove(sub)
		}
		delete(srv.all, id)
	}
}
func (s *server) add(v Variable) Identity {
	for {
		srv.seq++
		if _, ok := srv.all[srv.seq]; !ok && srv.seq != 0 {
			srv.all[srv.seq] = v
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

func PushStub(s stub) {
	srv.stubc <- s
}
