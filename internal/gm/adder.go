package gm

import "github.com/forrestjgq/gomark"

type Adder struct {
	id    identity
	name  string
	total int64
}

/********************************************************************************
                       Implementation of gomark.Marker
********************************************************************************/

// Mark a value
func (a *Adder) Mark(n int32) {
	s := makeStub(cmdMark, a.id, mark(n))
	PushStub(s)
}
func (a *Adder) Cancel() {
	RemoveVariable(a.id)
}

/********************************************************************************
                       Implementation of Variable
********************************************************************************/
func (a *Adder) Name() string {
	return a.name
}

func (a *Adder) Identity() identity {
	return a.id
}

func (a *Adder) Push(n mark) {
	a.total += int64(n)
}

// NewAdder create an adder
func NewAdder(name string) gomark.Marker {
	a := &Adder{
		id:   0,
		name: name,
	}
	a.id = AddVariable(a)
	return a
}
