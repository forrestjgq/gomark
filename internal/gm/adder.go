package gm

type Adder struct {
	id     Identity
	name   string
	total  int64
	series *IntSeries
}

/********************************************************************************
                       Implementation of gomark.Marker
********************************************************************************/

// Mark a value
func (a *Adder) Mark(n int32) {
	s := makeStub(cmdMark, a.id, Mark(n))
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

func (a *Adder) Identity() Identity {
	return a.id
}

func (a *Adder) Push(n Mark) {
	a.total += int64(n)
}

// NewAdder create an adder
func NewAdder(name string) *Adder {
	a := &Adder{
		id:   0,
		name: name,
	}
	a.id = AddVariable(a)
	return a
}
