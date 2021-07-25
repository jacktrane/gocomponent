package gopool

import (
	"fmt"
	"sync"
	"sync/atomic"
)

// TODO 不能一直占着茅坑不拉屎，得想办法缩减协程

// GoPool 协程池
type GoPool struct {
	workerNum int
	workers   chan GoFunc
	l         sync.Locker
	reloadNum int32
}

// GoFunc 协程池函数
type GoFunc struct {
	f    func(para interface{})
	para interface{}
}

// NewGoPool 初始化协程池
// workerNum 协程池大小
func NewGoPool(workerNum int) *GoPool {
	return newGP(workerNum)
}

func newGP(wokerNum int) *GoPool {
	pool := GoPool{
		workerNum: wokerNum,
		workers:   make(chan GoFunc, wokerNum),
		l:         &sync.Mutex{},
	}

	for i := 0; i < wokerNum; i++ {
		go pool.worker()
	}

	return &pool
}

func (gp *GoPool) worker() {
	for worker := range gp.workers {
		worker.f(worker.para)
	}
}

// AddGoFunc 添加函数
func (gp *GoPool) AddGoFunc(f func(para interface{}), para interface{}) {
	gp.workers <- GoFunc{
		f:    f,
		para: para,
	}
}

// ElasticAddGoFunc 弹性扩容
func (gp *GoPool) ElasticAddGoFunc(f func(para interface{}), para interface{}) {
	fmt.Println("gp reloadNum", gp.reloadNum, gp.workerNum)
	if gp.WaitFuncNum() == gp.workerNum {
		num := gp.workerNum + gp.workerNum/(int(gp.reloadNum)+1)
		atomic.AddInt32(&gp.reloadNum, 1)
		gp.ChangeWorkerNum(num)
	}

	gp.workers <- GoFunc{
		f:    f,
		para: para,
	}
}

// Close 关闭协程池
func (gp *GoPool) Close() {
	close(gp.workers)
}

// WaitFuncNum 排队数量
func (gp *GoPool) WaitFuncNum() int {
	return len(gp.workers)
}

// ChangeWorkerNum 修改worker的数量
func (gp *GoPool) ChangeWorkerNum(num int) {
	gp.l.Lock()
	defer gp.l.Unlock()

	chgCh := make(chan GoFunc, num)
	tmpCh := gp.workers
	gp.workers = chgCh
	gp.workerNum = num
	close(tmpCh)

	for i := 0; i < num; i++ {
		go gp.worker()
	}
}
