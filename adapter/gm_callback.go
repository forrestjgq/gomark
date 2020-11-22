package adapter

/*
#include "gmhook.h"
*/
import "C"
import (
	"reflect"

	"github.com/forrestjgq/gomark"
	"github.com/forrestjgq/gomark/gmi"
)

//export gmCreate
func gmCreate(typ int, name *C.char) C.int {
	var m gmi.Marker
	switch typ {
	case C.VAR_LATENCY_RECORDER:
		m = gomark.NewLatencyRecorder(C.GoString(name))
	case C.VAR_ADDER:
		m = gomark.NewAdder(C.GoString(name))
	case C.VAR_MAXER:
		m = gomark.NewAdder(C.GoString(name))
	case C.VAR_STATUS:
		m = gomark.NewStatus(C.GoString(name))
	case C.VAR_PERSECOND_ADDER:
		m = gomark.NewAdderPerSecond(C.GoString(name))
	default:
		return 0
	}

	v := reflect.ValueOf(m).Uint()
	return C.int(v)
}

//export gmMark
func gmMark(id, value C.int) {
	m := gomark.IntToMarker(int(id))
	m.Mark(int32(value))
}

//export gmCancel
func gmCancel(id C.int) {
	m := gomark.IntToMarker(int(id))
	m.Cancel()
}
