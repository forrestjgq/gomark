package gm

import (
	"io"
)

type DisplayFilter int

const (
	DisplayOnNothing DisplayFilter = iota
	DisplayOnHTML
	DisplayOnPlainText
	DisplayOnAll
)

type SeriesOption struct {
	FixedLength, TestOnly bool
}

func NewSeriesOption() *SeriesOption {
	return &SeriesOption{
		FixedLength: true,
		TestOnly:    false,
	}
}

type Dumper interface {
	Dump(name, desc string) bool
}
type DumpOption struct {
	QuoteString    bool
	QuestionMark   byte
	DisplayFilter  DisplayFilter
	WhiteWildcards string
	BlackWildcards string
}
type Pushable interface {
	Push(v Mark)
}
type VarBase struct {
	name          string
	id            Identity
	displayFilter DisplayFilter
	child         []Identity
	disposer  []disposer
}

func (vb *VarBase) Mark(n int32) {
	if vb != nil && vb.Valid() {
		s := makeStub(vb.ID(), Mark(n))
		PushStub(s)
	}
}

func (vb *VarBase) Cancel() {
	if vb != nil && vb.Valid() {
		RemoveVariable(vb.ID())
	}
}

func (vb *VarBase) Name() string {
	return vb.name
}
func (vb *VarBase) ID() Identity {
	return vb.id
}
func (vb *VarBase) AddDisposer(d disposer) {
	vb.disposer = append(vb.disposer, d)
}
func (vb *VarBase) AddChild(id ...Identity) {
	vb.child = append(vb.child, id...)
}
func (vb *VarBase) Valid() bool {
	return vb.id != 0
}
func (vb *VarBase) GetDisplayFilter() DisplayFilter {
	return vb.displayFilter
}

type vbs []*VarBase

func (v vbs) Len() int           { return len(v) }
func (v vbs) Less(i, j int) bool { return v[i].name < v[j].name }
func (v vbs) Swap(i, j int)      { v[i], v[j] = v[j], v[i] }

type Variable interface {
	VarBase() *VarBase
	// Called internally inside server's routine to mark variable
	Push(v Mark)
	// Notify variable to dispose, and return all child variables that need to be removed
	Dispose()
	Describe(w io.StringWriter, quote bool)
	DescribeSeries(w io.StringWriter, opt *SeriesOption) error
}
