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

/*
struct DumpOptions {
    // Constructed with default options.
    DumpOptions();

    // If this is true, string-type values will be quoted.
    bool quote_string;

    // The ? in wildcards. Wildcards in URL need to use another character
    // because ? is reserved.
    char question_mark;

    // Dump variables with matched display_filter
    DisplayFilter display_filter;

    // Name matched by these wildcards (or exact names) are kept.
    std::string white_wildcards;

    // Name matched by these wildcards (or exact names) are skipped.
    std::string black_wildcards;
};
*/
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
type VarBase struct {
	name          string
	id            Identity
	displayFilter DisplayFilter
	child         []Identity
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
	Push(v Mark)
	// Notify variable to dispose, and return all child variables that need to be removed
	Dispose() []Identity
	Describe(w io.StringWriter, quote bool)
	DescribeSeries(w io.StringWriter, opt *SeriesOption) error
}
