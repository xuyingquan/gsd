package util

import (
	"testing"
	"time"
	"fmt"
)

func TestPutAndGet(t *testing.T){
	ttl := 3
	cache := NewCache(ttl)
	cache.Put("a", "a")
	value, ok := cache.Get("a")
	if !ok || value.(string) != "a" {
		fmt.Println("1")
		t.FailNow()
	}
	time.Sleep(time.Duration(1 * 1000 * 1000 * 1000))
	value, ok = cache.Get("a")
	if !ok {
		fmt.Println("2")
		t.FailNow()
	}
	time.Sleep(time.Duration((ttl) * 1000 * 1000 * 1000))
	value, ok = cache.Get("a")
	if ok {
		fmt.Println("3")
		t.FailNow()
	}
}

func TestMemoryUse(t *testing.T) {
	ttl := 10
	cache := NewCache(ttl)
	for i := 0;i < 1000000;i ++ {
		cache.Put(i, i)
	}
}
