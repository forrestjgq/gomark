package gomark
type command uint8
type identity uint16
type mark int32
const (
	cmdLock = iota
	cmdNew
	cmdCancel
	cmdMark
)


type stub int64
func (st stub) cmd() command  {
	// todo
	return 0
}
func (st stub) send() {
	srv.c <- st
}
func makeStub(cmd command, id identity, m mark) stub  {
	// todo
	return 0
}
