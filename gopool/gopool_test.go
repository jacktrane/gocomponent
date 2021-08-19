package gopool

import (
	"fmt"
	"testing"
	"time"
)

func TestGetElasticAddGoFunc(t *testing.T) {
	pool := NewGoPool(100)
	defer pool.Close()
	for i := 0; i < 100000; i++ {
		pool.ElasticAddGoFunc(f1, "1")
	}
	time.Sleep(10 * time.Second)
}

func BenchmarkAddGoFunc(b *testing.B) {
	b.Run("test gopool", func(b *testing.B) {
		pool := NewGoPool(100)
		defer pool.Close()
		for i := 0; i < 100000; i++ {
			pool.ElasticAddGoFunc(f1, "1")
		}
		time.Sleep(10 * time.Second)
	})
}

func TestGetWaitNum(t *testing.T) {
	pool := NewGoPool(100)
	defer pool.Close()
	for i := 0; i < 1000; i++ {
		fmt.Println(pool.WaitFuncNum())
		if i == 200 {
			pool.ChangeWorkerNum(200)
		}
		pool.AddGoFunc(f1, "1")
	}
	time.Sleep(10 * time.Second)
}

func TestGoPool(t *testing.T) {
	pool := NewGoPool(2)
	defer pool.Close()
	for i := 0; i < 2; i++ {
		p := Para{
			a: i,
			b: "A",
			c: float32(i),
			d: true,
		}

		pool.AddGoFunc(f, p)
	}

	for i := 2; i < 4; i++ {
		p := Para{
			a: i,
			b: "B",
			c: float32(i),
			d: false,
		}

		pool.AddGoFunc(f, p)
	}

	for i := 4; i < 6; i++ {
		p := Para{
			a: i,
			b: "C",
			c: float32(i),
			d: true,
		}

		pool.AddGoFunc(f, p)
	}
}

type Para struct {
	a int
	b string
	c float32
	d bool
}

func f(p interface{}) {
	para := p.(Para)
	fmt.Println(time.Now().UnixNano(), para.a, para.b, para.c, para.d)
}

func f1(p interface{}) {
	time.Sleep(2 * time.Second)
}
