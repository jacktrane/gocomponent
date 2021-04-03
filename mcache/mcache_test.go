package mcache

import (
	"fmt"
	"testing"
	"time"

	"github.com/jacktrane/gocomponent/time_cost"
)

func BenchmarkCache(b *testing.B) {
	cache := NewMcache()
	ilen := b.N
	h := time_cost.NewTimeCost()
	for i := 0; i < ilen; i++ {
		if i%10 == 0 {
			cache.Set(i, i, time.Now().Unix()+10)
		} else {
			cache.Set(i, i)
		}
	}
	h.AddPoint("Set")
	for i := 0; i < ilen; i++ {
		fmt.Println(cache.Get(i))
	}
	h.AddPoint("Get")
	fmt.Println(h.OutputCostStack(), ilen)
}
func TestCache(t *testing.T) {
	cache := NewMcache()
	ilen := 1000
	ch := make(chan struct{}, ilen)
	for i := 0; i < ilen; i++ {
		go func(i int) {
			if i%10 == 0 {
				cache.Set(i, i, time.Now().Unix()+10)
			} else {
				cache.Set(i, i)
			}
		}(i)
	}

	for i := 0; i < ilen; i++ {
		go func(i int) {
			fmt.Println(cache.Get(i))
		}(i)
	}
	for i := 0; i < ilen; i++ {
		<-ch
	}
}
