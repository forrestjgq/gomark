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

// sampler is an entity that can be scheduled to take samples.
// AddSampler
type sampler interface {
	takeSample()
}

// callback to be called on variable removing
type disposer func()

type server struct {
	seq      Identity
	smp      Identity
	stubc    chan stub
	callc    chan func()
	all      map[Identity]Variable
	samplers map[Identity]sampler
	tk       *time.Ticker

	// internal variables
	sampleLatency     *LatencyRecorder
	markLatency       *LatencyRecorder
	remoteCallLatency *LatencyRecorder
	exposeAdder       *Adder
	samplerAdder      *Adder
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

	//srv.markLatency.vb.EnablePerf()
}

func (s *server) run() {
	for {
		select {
		case f := <-s.callc:
			if srv.remoteCallLatency != nil {
				srv.remoteCallLatency.doUs(f)
			} else {
				f()
			}
		case rx := <-s.stubc:
			if s.markLatency != nil {
				srv.markLatency.doUs(func() {
					s.markStub(rx)
				})
			} else {
				s.markStub(rx)
			}
		case <-s.tk.C:
			if srv.sampleLatency != nil {
				srv.sampleLatency.doUs(func() {
					s.sample()
				})
			} else {
				s.sample()
			}
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
		//glog.Infof("Take sample for %T", smp)
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
		for _, d := range v.VarBase().disposer {
			if d != nil {
				d()
			}
		}
		delete(s.all, id)
		if s.exposeAdder != nil {
			s.exposeAdder.Push(-1)
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
	if option == nil {
		option = &DumpOption{
			QuoteString:    true,
			QuestionMark:   '?',
			DisplayFilter:  DisplayOnPlainText,
			WhiteWildcards: "",
			BlackWildcards: "",
		}
	}
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
	for _, v := range srv.all {
		if v.VarBase().name == vb.name {
			return nil, fmt.Errorf("duplicate variable name %s", vb.name)
		}
	}
	srv.all[vb.id] = v
	if srv.exposeAdder != nil {
		srv.exposeAdder.Push(1)
	}

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
	if srv.samplerAdder != nil {
		srv.samplerAdder.Push(1)
	}
	return func() {
		srv.removeSample(id)
		if srv.samplerAdder != nil {
			srv.samplerAdder.Push(-1)
		}
	}
}

func RemoteCall(call func()) {
	callFunc(call)
}

func MakeSureEmpty() {
	fail := false
	if len(srv.samplers) > 0 {
		glog.Errorf("sampler not cleared:")
		for _, s := range srv.samplers {
			glog.Infof("    %T", s)
		}

		fail = true
	}

	if len(srv.all) > 0 {
		glog.Errorf("variables not cleared:")
		for _, v := range srv.all {
			glog.Infof("    name: %s type: %T", v.VarBase().Name(), v)
		}
		fail = true
	}

	if fail {
		glog.Fatalf("gomark check fail")
	}
}
func DisableInternalVariables() {
	t := srv.sampleLatency
	if t != nil {
		srv.sampleLatency = nil
		t.vb.Cancel()
	}

	t = srv.markLatency
	if t != nil {
		srv.markLatency = nil
		t.vb.Cancel()
	}

	t = srv.remoteCallLatency
	if t != nil {
		srv.remoteCallLatency = nil
		t.vb.Cancel()
	}

	a := srv.exposeAdder
	if a != nil {
		srv.exposeAdder = nil
		a.vb.Cancel()
	}

	a = srv.samplerAdder
	if a != nil {
		srv.samplerAdder = nil
		a.vb.Cancel()
	}
}
func EnableInternalVariables() {
	if srv.sampleLatency == nil {
		srv.sampleLatency, _ = NewLatencyRecorder("internal_sample_us")
	}
	if srv.markLatency == nil {
		srv.markLatency, _ = NewLatencyRecorder("internal_mark_us")
	}
	if srv.remoteCallLatency == nil {
		srv.remoteCallLatency, _ = NewLatencyRecorder("internal_remote_call_us")
	}
	if srv.exposeAdder == nil {
		srv.exposeAdder, _ = NewAdder("internal_variable")
		srv.exposeAdder.Push(Mark(len(srv.all)))
	}
	if srv.samplerAdder == nil {
		srv.samplerAdder, _ = NewAdder("internal_sampler")
		srv.exposeAdder.Push(Mark(len(srv.samplers)))
	}
}
