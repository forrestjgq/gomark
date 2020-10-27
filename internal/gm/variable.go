package gm

type Variable interface {
	Name() string
	Identity() identity
	Push(v mark)
}
