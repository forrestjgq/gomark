package gm

import (
	"testing"
	"time"
)

func TestAdder(t *testing.T) {
	adder, err := NewAdder("v")
	if err != nil {
		t.Fatalf(err.Error())
	}

	adder.VarBase().Mark(1)
	adder.VarBase().Mark(2)
	adder.VarBase().Mark(4)

	time.Sleep(10 * time.Millisecond)

	if adder.GetValue() != 7 {
		t.Fatalf("invalid value: %d", adder.GetValue())
	}

	adder.VarBase().Cancel()
	MakeSureEmpty()

	adder, _ = NewAdderNoExpose()
	adder.Push(-9)
	adder.Push(2)
	adder.Push(4)
	adder.Push(0)

	if adder.GetValue() != -3 {
		t.Fatalf("invalid value: %d", adder.GetValue())
	}

	MakeSureEmpty()
}
