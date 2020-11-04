package gm

import "io"

type DisplayFilter int

const (
	DisplayOnHTML DisplayFilter = iota + 1
	DisplayOnPlainText
	DisplayOnAll
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
type Variable interface {
	Name() string
	Identity() Identity
	Push(v Mark)
	Reset() int64
	Value() int64
	OnExpose()
	OnSample()
	Describe(w io.Writer, quote bool)
	DescribeSeries(w io.Writer, opt *SeriesOption) error
}
