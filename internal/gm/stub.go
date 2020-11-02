package gm

type Command uint8
type Identity uint16
type Mark int32

const (
	cmdLock Command = iota
	cmdNew
	cmdCancel
	cmdMark
)

type stub struct {
	c   Command
	who Identity
	m   Mark
}

func (st stub) cmd() Command {
	return st.c
}
func (st stub) identity() Identity {
	return st.who
}
func (st stub) mark() Mark {
	return st.m
}
func makeStub(cmd Command, id Identity, m Mark) stub {
	return stub{
		c:   cmd,
		who: id,
		m:   m,
	}
}
