package util

type ErrorMerge struct {
	err error
}

func (e *ErrorMerge) Failed() bool {
	return e.err != nil
}
func (e *ErrorMerge) Merge(err error) *ErrorMerge {
	if e.err == nil {
		e.err = err
	}
	return e
}

func NewErrorMerge() *ErrorMerge {
	return &ErrorMerge{}
}
