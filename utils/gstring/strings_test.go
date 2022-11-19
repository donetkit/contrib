package gstring

import "testing"

func TestStringLen(t *testing.T) {
	var str = "Test"
	v := Len(str)
	if v != 4 {
		t.Error("error")
	}
	str = "你好HL"
	v = Len(str)
	if v != 4 {
		t.Error("error")
	}
	str = "你好@,"
	v = Len(str)
	if v != 4 {
		t.Error("error")
	}
	str = "你1好?"
	v = Len(str)
	if v != 4 {
		t.Error("error")
	}
}

func TestLowercaseFirst(t *testing.T) {
	var str = "TestLowercaseFirst"
	v := LowercaseFirst(str)
	if v != "testLowercaseFirst" {
		t.Error("error")
	}
}

func TestUppercaseFirst(t *testing.T) {
	var str = "uppercaseFirst"
	v := UppercaseFirst(str)
	t.Log(v)
	if v != "UppercaseFirst" {
		t.Error("error")
	}
}

func TestContains(t *testing.T) {
	d := []string{"A", "B", "C"}
	v := Contains(d, "A")
	if !v {
		t.Error("error")
	}
	v = Contains(d, "b")
	if v {
		t.Error("error")
	}
	v = Contains(d, "C")
	if !v {
		t.Error("error")
	}
	v = Contains(d, "D")
	if v {
		t.Error("error")
	}
}

func TestContainsString(t *testing.T) {
	d := []string{"A", "B", "C"}
	v := ContainsString(d, "A")
	if !v {
		t.Error("error")
	}
	v = ContainsString(d, "a")
	t.Log(v)
	if v {
		t.Error("error")
	}
}

func TestSub(t *testing.T) {
	d := "ABCDEFGHIJKLMNOP"
	v := Sub(d, 6, 6)
	t.Log(v)
	if v != "GHIJKL" {
		t.Error("error")
	}
}

func TestRandomString(t *testing.T) {
	v := RandomString(16)
	t.Log(v)
	if len(v) != 16 {
		t.Error("error")
	}
}

func TestJoin(t *testing.T) {
	d := []string{"A", "B", "C"}
	v := Join(d, ",")
	t.Log(v)
	if v != "A,B,C" {
		t.Error("error")
	}

}

func TestNewLine(t *testing.T) {
	v := NewLine()
	t.Log(v)
	if v != "\r\n" {
		t.Error("error")
	}

}
