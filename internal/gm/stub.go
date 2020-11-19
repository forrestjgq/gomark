package gm

type Identity uint32

func (i Identity) Mark(n int32) {
	if i != 0 {
		s := makeStub(i, Mark(n))
		PushStub(s)
	}
}

func (i Identity) Cancel() {
	if i != 0 {
		RemoveVariable(i)
	}
}

type Mark int32

type stub struct {
	who Identity
	m   Mark
}

func (st stub) identity() Identity {
	return st.who
}
func (st stub) mark() Mark {
	return st.m
}
func makeStub(id Identity, m Mark) stub {
	return stub{
		who: id,
		m:   m,
	}
}
