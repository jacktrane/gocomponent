package gopool

import (
	"fmt"
	"testing"
	"time"
)

func TestGoPool(t *testing.T) {
	pool := NewPool(2, 0)
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
