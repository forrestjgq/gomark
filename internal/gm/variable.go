package gm

import "io"

type DisplayFilter int

const (
	DisplayOnHTML DisplayFilter = iota + 1
	DisplayOnPlainText
	DisplayOnAll
	DisplayOnNothing
)

type SeriesOption struct {
	fixedLength, testOnly bool
}

func NewSeriesOption() *SeriesOption {
	return &SeriesOption{
		fixedLength: true,
		testOnly:    false,
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
type DumpOption struct {
	quoteString    bool
	questionMark   int8
	displayFilter  DisplayFilter
	whiteWildcards string
	blackWildcards string
}
type VarBase struct {
	name          string
	id            Identity
	displayFilter DisplayFilter
}

func (vb *VarBase) Name() string {
	return vb.name
}
func (vb *VarBase) ID() Identity {
	return vb.id
}
func (vb *VarBase) Valid() bool {
	return vb.id != 0
}
func (vb *VarBase) GetDisplayFilter() DisplayFilter {
	return vb.displayFilter
}

type Variable interface {
	VarBase() *VarBase
	Push(v Mark)
	// OnExpose is called inside server after variable is registered
	OnExpose(vb *VarBase) error
	//Dispose()
	OnSample()
	Describe(w io.StringWriter, quote bool)
	DescribeSeries(w io.StringWriter, opt *SeriesOption) error
}
