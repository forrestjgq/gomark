package util

import "testing"

func TestWildMatcher(t *testing.T) {
	m := NewWildcardMatcher("*decode*,*hello*", '?', true)
	b := m.Match("this_is_a_decode_name")
	if !b {
		t.Fatalf("fail match")
	}
	b = m.Match("this_is_a_hello_name")
	if !b {
		t.Fatalf("fail match")
	}
	b = m.Match("this_is_a_name")
	if b {
		t.Fatalf("fail match")
	}
	b = m.Match("")
	if b {
		t.Fatalf("fail match")
	}
	b = m.Match("decode")
	if !b {
		t.Fatalf("fail match")
	}
	b = m.Match("hello")
	if !b {
		t.Fatalf("fail match")
	}
	b = m.Match("hellodecode")
	if !b {
		t.Fatalf("fail match")
	}
	b = m.Match("")
	if b {
		t.Fatalf("fail match")
	}
}
func TestWildMatcherHead(t *testing.T) {
	m := NewWildcardMatcher("*decode", '?', true)
	b := m.Match("decode_name")
	if b {
		t.Fatalf("fail match")
	}
	b = m.Match("jgq_decode")
	if !b {
		t.Fatalf("fail match")
	}
}
func TestWildMatcherTail(t *testing.T) {
	m := NewWildcardMatcher("decode*", '?', true)
	b := m.Match("decode_name")
	if !b {
		t.Fatalf("fail match")
	}
	b = m.Match("jgq_decode")
	if b {
		t.Fatalf("fail match")
	}
}
func TestEmptyMatcher(t *testing.T) {
	m := NewWildcardMatcher("", '?', false)
	b := m.Match("")
	if b {
		t.Fatalf("fail match")
	}
	b = m.Match("hello")
	if b {
		t.Fatalf("fail match")
	}
	m = NewWildcardMatcher("", '?', true)
	b = m.Match("")
	if !b {
		t.Fatalf("fail match")
	}
}
