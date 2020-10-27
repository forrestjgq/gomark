package gm

type command uint8
type identity uint16
type mark int32

const (
	cmdLock command = iota
	cmdNew
	cmdCancel
	cmdMark
)

type stub struct {
	c   command
	who identity
	m   mark
}

func (st stub) cmd() command {
	return st.c
}
func (st stub) identity() identity {
	return st.who
}
func (st stub) mark() mark {
	return st.m
}
func makeStub(cmd command, id identity, m mark) stub {
	return stub{
		c:   cmd,
		who: id,
		m:   m,
	}
}
