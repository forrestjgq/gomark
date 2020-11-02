package gm

type Variable interface {
	Name() string
	Identity() Identity
	Push(v Mark)
}
