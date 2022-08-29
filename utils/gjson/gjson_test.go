package gjson

import "testing"

type Test struct {
	Name string
	Age  int64
}

func TestJson(t *testing.T) {
	t.Log(MarshalIndent(Test{"Hijacker", 19}))
	t.Log(Marshal(Test{"Hijacker", 19}))
}
