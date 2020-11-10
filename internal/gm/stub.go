package gm

type Identity uint32
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
