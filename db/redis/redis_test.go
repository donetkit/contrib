package redis

import (
	"testing"
)

func TestCacheGet(t *testing.T) {
	tc := New(WithAddr("127.0.0.1"), WithPort(6379), WithPassword(""))
	v := tc.WithDB(0).Get("test")
	if v != nil {
		t.Error("Found c when it should have been automatically deleted")
	}
	v = tc.WithDB(0).Set("test", "test", 0)
	if v != nil {
		t.Error("Found c when it should have been automatically deleted")
	}
	v = tc.WithDB(0).Get("test")
	if v == nil {
		t.Error("Found c when it should have been automatically deleted")
	}
	v1 := tc.WithDB(0).Delete("test")
	t.Log(v1)
}

func BenchmarkCacheGet(b *testing.B) {
	b.StopTimer()
	tc := New(WithAddr("127.0.0.1"), WithPort(6379), WithPassword(""))
	cache := tc.WithDB(0)
	cache.Set("test", "test", 0)
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		cache.Get("test")
	}
	cache.Delete("test")
}
