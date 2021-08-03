package gopool

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/jacktrane/gocomponent/basic"
)

// GoPool 协程池
type GoPool struct {
	workerNum  int
	oWorkerNum int // 最原始的worker数量
	workers    chan GoFunc
	chL        sync.Locker
	reloadNum  int32
	reloading  uint32 // 原子操作
}

// GoFunc 协程池函数
type GoFunc struct {
	f    func(para interface{})
	para interface{}
}

// NewGoPool 初始化协程池
// workerNum 协程池大小
func NewGoPool(workerNum int) *GoPool {
	pool := newGP(workerNum)
	go pool.watch()
	return pool
}

func newGP(workerNum int) *GoPool {
	pool := GoPool{
		workerNum:  workerNum,
		oWorkerNum: workerNum,
		workers:    make(chan GoFunc, workerNum),
		chL:        &sync.Mutex{},
	}

	for i := 0; i < workerNum; i++ {
		go pool.worker()
	}

	return &pool
}

// 监控协程池变化，太多闲置就进行缩减
func (gp *GoPool) watch() {
	for {
		// 如果比策略调整前的数据还小就要缩容
		waitFuncNum := gp.WaitFuncNum()
		policyNum := gp.policy(false)
		if policyNum > waitFuncNum && gp.workerNum > gp.oWorkerNum && policyNum > gp.oWorkerNum {
			fmt.Println("flex", policyNum)
			gp.ChangeWorkerNum(policyNum)
		}

		time.Sleep(500 * time.Millisecond)
	}
}

func (gp *GoPool) worker() {
	for worker := range gp.workers {
		worker.f(worker.para)
	}
}

// AddGoFunc 添加函数
func (gp *GoPool) AddGoFunc(f func(para interface{}), para interface{}) {
	goFounc := GoFunc{
		f:    f,
		para: para,
	}
	gp.workers <- goFounc
}

// ElasticAddGoFunc 弹性扩容
func (gp *GoPool) ElasticAddGoFunc(f func(para interface{}), para interface{}) {
	if gp.WaitFuncNum() == gp.workerNum {
		fmt.Println("gp reloadNum", gp.reloadNum, gp.workerNum)
		num := gp.policy(true)
		atomic.AddInt32(&gp.reloadNum, 1)
		gp.ChangeWorkerNum(num)
	}

	goFounc := GoFunc{
		f:    f,
		para: para,
	}

	for {
		if atomic.LoadUint32(&gp.reloading) == 0 {
			gp.workers <- goFounc
			break
		}
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
// 弹性过程中事实上有时候会出现double倍协程
func (gp *GoPool) ChangeWorkerNum(num int) {
	// 串行改变协程数量
	gp.chL.Lock()
	defer gp.chL.Unlock()

	// 用原子方式保证不往已关闭的channel塞数据
	if atomic.LoadUint32(&gp.reloading) != 0 {
		fmt.Println("gp.reloading")
		return
	}
	atomic.StoreUint32(&gp.reloading, 1)
	defer atomic.CompareAndSwapUint32(&gp.reloading, 1, 0)

	chgCh := make(chan GoFunc, num)
	tmpCh := gp.workers
	defer close(tmpCh)

	gp.workers = chgCh
	gp.workerNum = num
	for i := 0; i < num; i++ {
		go gp.worker()
	}
}

func (gp *GoPool) policy(needExpand bool) int {
	meta := gp.workerNum / (int(gp.reloadNum) + 1)
	return basic.IfElseInt(needExpand, gp.workerNum+meta, gp.workerNum-meta)
}

func emptyFunc(emptyPara interface{}) {}
