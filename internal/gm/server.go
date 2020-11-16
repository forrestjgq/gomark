package gm

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"sort"
	"sync"
	"time"
	"unicode"

	"github.com/forrestjgq/gomark/internal/util"
	"github.com/golang/glog"
)

const (
	sizeOfQ = 10000
)

type sampler interface {
	takeSample()
}

type disposer func()

type server struct {
	seq      Identity
	smp      Identity
	stubc    chan stub
	callc    chan func()
	all      map[Identity]Variable
	samplers map[Identity]sampler
	tk       *time.Ticker
}

var srv *server

func init() {
	srv = &server{
		stubc:    make(chan stub, sizeOfQ),
		callc:    make(chan func()),
		all:      make(map[Identity]Variable),
		samplers: make(map[Identity]sampler),
		tk:       time.NewTicker(time.Second),
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
		case <-s.tk.C:
			s.sample()
		}
	}
}
func (s *server) removeSample(id Identity) {
	delete(s.samplers, id)
}
func (s *server) addSample(smp sampler) Identity {
	for {
		s.smp++
		if _, ok := s.samplers[s.smp]; !ok && s.smp != 0 {
			s.samplers[s.smp] = smp
			return s.smp
		}
	}
}
func (s *server) sample() {
	for _, smp := range s.samplers {
		smp.takeSample()
	}
}
func (s *server) markStub(stub stub) {
	if v, ok := s.all[stub.identity()]; ok {
		v.Push(stub.mark())
	}
}
func (s *server) remove(id Identity) {
	if v, ok := s.all[id]; ok {
		v.Dispose()
		ids := v.VarBase().child
		for _, sub := range ids {
			s.remove(sub)
		}
		delete(s.all, id)
	}
}
func (s *server) add(v Variable) Identity {
	for {
		s.seq++
		if _, ok := s.all[s.seq]; !ok && s.seq != 0 {
			s.all[s.seq] = v
			return s.seq
		}
	}
}
func (s *server) describe(v Variable, w io.StringWriter, quoteString bool, filter DisplayFilter) error {
	if filter&v.VarBase().displayFilter == 0 {
		return errors.New(v.VarBase().name + " do not match filter")
	}

	v.Describe(w, quoteString)
	return nil
}
func (s *server) describeID(id Identity, w io.StringWriter, quoteString bool, filter DisplayFilter) error {
	if vb, ok := s.all[id]; ok {
		return s.describe(vb, w, quoteString, filter)
	}
	return fmt.Errorf("var not exist: %d", id)
}
func (s *server) describeVar(name string, w io.StringWriter, quoteString bool, filter DisplayFilter) error {
	for _, vb := range s.all {
		if vb.VarBase().name == name {
			return s.describe(vb, w, quoteString, filter)
		}
	}
	return fmt.Errorf("var not exist: %s", name)
}
func (s *server) listVars(filter DisplayFilter) []*VarBase {
	ret := []*VarBase{}
	for _, vb := range s.all {
		if vb.VarBase().displayFilter&filter != 0 {
			ret = append(ret, vb.VarBase())
		}
	}
	return ret
}

func (s *server) describeSeries(name string, w io.StringWriter, option *SeriesOption) error {
	for _, vb := range s.all {
		if vb.VarBase().name == name {
			return vb.DescribeSeries(w, option)
		}
	}
	return fmt.Errorf("var not exist: %s", name)

}
func (s *server) dump(dumper Dumper, option *DumpOption) (int, error) {
	black := util.NewWildcardMatcher(option.BlackWildcards, option.QuestionMark, false)
	white := util.NewWildcardMatcher(option.WhiteWildcards, option.QuestionMark, true)

	logDump := flagLogDumped
	logBuf := &bytes.Buffer{}

	buf := &bytes.Buffer{}
	count := 0
	if len(white.Wildcards()) == 0 && len(white.ExactNames()) > 0 {
		for _, name := range white.ExactNames() {
			if !black.Match(name) {
				err := s.describeVar(name, buf, option.QuoteString, option.DisplayFilter)
				if err != nil {
					continue
				}
				if logDump {
					logBuf.WriteString("\n")
					logBuf.WriteString(name)
					logBuf.WriteString(buf.String())
				}
				if !dumper.Dump(name, buf.String()) {
					return -1, fmt.Errorf("dump %s fail", name)
				}
				buf.Reset()
				count++
			}
		}
	} else {
		varnames := s.listVars(option.DisplayFilter)
		sort.Sort(vbs(varnames))
		for _, v := range varnames {
			name := v.name
			if white.Match(name) && !black.Match(name) {
				err := s.describeID(v.id, buf, option.QuoteString, option.DisplayFilter)
				if err != nil {
					continue
				}
				if logDump {
					logBuf.WriteString("\n")
					logBuf.WriteString(name)
					logBuf.WriteString(buf.String())
				}
				if !dumper.Dump(name, buf.String()) {
					return -1, fmt.Errorf("dump %s fail", name)
				}
				buf.Reset()
				count++
			}
		}
	}
	if logDump {
		glog.Info("Dumpped variables:", logBuf.String())
	}

	return count, nil
}

func toUnderScore(name, src string) string {
	var last rune
	for i, c := range src {
		if unicode.IsLetter(c) {
			if unicode.IsUpper(c) {
				if i != 0 && last == 0 && name[len(name)-1] != '_' {
					name += "_"
				}
				name += string(unicode.ToLower(c))
			} else {
				name += string(c)
			}
		} else if unicode.IsDigit(c) {
			name += string(c)
		} else if len(name) == 0 || name[len(name)-1] != '_' {
			name += "_"
		}

		last = c
	}
	return name
}
func Expose(prefix, name string, displayFilter DisplayFilter, v Variable) (*VarBase, error) {
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
	srv.all[vb.id] = v
	// todo: check dupliacted name
	return vb, nil
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
func RemoveVariable(id Identity) {
	f := func() {
		srv.remove(id)
	}
	callFunc(f)
}
func DescribeVarSeries(varName string, w io.StringWriter, option SeriesOption) error {
	var err error
	f := func() {
		err = srv.describeSeries(varName, w, &option)
	}
	callFunc(f)
	return err
}
func DescribeVarSeriesSafe(varName string, w io.StringWriter, option SeriesOption) error {
	return srv.describeSeries(varName, w, &option)
}

func Dump(dumper Dumper, option *DumpOption) (int, error) {
	var ret int
	var err error
	f := func() {
		ret, err = srv.dump(dumper, option)
	}
	callFunc(f)
	return ret, err
}

func PushStub(s stub) {
	srv.stubc <- s
}

func AddSampler(s sampler) disposer {
	id := srv.addSample(s)
	return func() {
		srv.removeSample(id)
	}
}

func RemoteCall(call func()) {
	callFunc(call)
}
