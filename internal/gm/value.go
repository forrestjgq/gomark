package gm

type Value struct {
	x, y int64
}

func (v *Value) Reset() {
	v.x, v.y = 0, 0
}

func (v *Value) Sub(rhs *Value) Value {
	return Value{
		x: v.x - rhs.x,
		y: v.y - rhs.y,
	}
}
func (v *Value) Add(rhs *Value) Value {
	return Value{
		x: v.x + rhs.x,
		y: v.y + rhs.y,
	}
}
func (v *Value) Reduce(rhs *Value) {
	v.x -= rhs.x
	v.y -= rhs.y
}
func (v *Value) Append(rhs *Value) {
	v.x += rhs.x
	v.y += rhs.y
}

func OneValue(x int64) Value {
	return Value{
		x: x,
		y: 0,
	}
}
func ValueOf(x, y int64) Value {
	return Value{
		x: x,
		y: y,
	}
}
