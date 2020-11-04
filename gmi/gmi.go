package gmi

type Marker interface {
	Mark(n int32)
	Cancel()
}
