package gm

import (
	"fmt"
	"math"
	"reflect"
	"strconv"
	"unsafe"

	"github.com/forrestjgq/glog"
)

type Value struct {
	x, y int64
}

type ValueSerializer func(v Value) string

var XValueSerializer ValueSerializer = func(v Value) string {
	return strconv.Itoa(int(v.x))
}
var YValueSerializer ValueSerializer = func(v Value) string {
	return strconv.Itoa(int(v.y))
}
var VectorValueSerializer ValueSerializer = func(v Value) string {
	slice := DivideToSliceU32(v)
	quote := ""
	if FlagQuoteVector {
		quote = "\""
	}
	return fmt.Sprintf("%s[%d,%d,%d,%d]%s", quote, slice[0], slice[1], slice[2], slice[3], quote)
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
func (v *Value) AverageInt() int64 {
	if v.y == 0 {
		return 0
	}
	return v.x / v.y
}
func (v *Value) AverageFloat() float64 {
	if v.y == 0 {
		return 0
	}
	return float64(v.x) / float64(v.y)
}
func (v *Value) GetU32(idx int) uint32 {
	return *(*uint32)(unsafe.Pointer(uintptr(unsafe.Pointer(v)) + uintptr(idx*4)))
}
func (v *Value) SetU32(idx int, val uint32) {
	*(*uint32)(unsafe.Pointer(uintptr(unsafe.Pointer(v)) + uintptr(idx*4))) = val
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
func CombineToValue(s []int64) Value {
	var v Value
	tmp := reflect.SliceHeader{
		Data: uintptr(unsafe.Pointer(&v)),
		Len:  4,
		Cap:  4,
	}
	src := *(*[]int32)(unsafe.Pointer(&tmp))
	for i, v := range s {
		if v >= math.MaxInt32 {
			glog.Error("overflow")
			src[i] = math.MaxInt32
		} else {
			src[i] = int32(v)
		}
	}
	return v
}
func DivideToSlice(v Value) []int32 {
	dst := make([]int32, 4)
	tmp := reflect.SliceHeader{
		Data: uintptr(unsafe.Pointer(&v)),
		Len:  4,
		Cap:  4,
	}
	src := *(*[]int32)(unsafe.Pointer(&tmp))
	copy(dst, src)
	return dst
}
func CombineToValueU32(s []uint32) Value {
	var v Value
	tmp := reflect.SliceHeader{
		Data: uintptr(unsafe.Pointer(&v)),
		Len:  4,
		Cap:  4,
	}
	dst := *(*[]uint32)(unsafe.Pointer(&tmp))
	copy(dst, s)
	return v
}
func DivideToSliceU32(v Value) []uint32 {
	dst := make([]uint32, 4)
	tmp := reflect.SliceHeader{
		Data: uintptr(unsafe.Pointer(&v)),
		Len:  4,
		Cap:  4,
	}
	src := *(*[]uint32)(unsafe.Pointer(&tmp))
	copy(dst, src)
	return dst
}

var statOperatorInt OperatorInt = func(left Value, right int) Value {
	if right == 0 {
		return left
	}
	left.x /= int64(right)
	left.y /= int64(right)
	return left
}
