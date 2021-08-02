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
	inputL     sync.Locker
	chL        sync.Locker
	reloadNum  int32
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
		inputL:     &sync.Mutex{},
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
		if policyNum > waitFuncNum && gp.workerNum > gp.oWorkerNum {
			fmt.Println("flex", policyNum)
			gp.ChangeWorkerNum(policyNum)
		}

		time.Sleep(1000 * time.Millisecond)
	}
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
	if gp.WaitFuncNum() == gp.workerNum {
		fmt.Println("gp reloadNum", gp.reloadNum, gp.workerNum)
		num := gp.policy(true)
		atomic.AddInt32(&gp.reloadNum, 1)
		gp.ChangeWorkerNum(num)
	}
	// TODO 这里有问题，串行意义不大
	gp.chL.Lock()
	defer gp.chL.Unlock()
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
// 弹性过程中事实上有时候会出现double倍协程
func (gp *GoPool) ChangeWorkerNum(num int) {
	gp.chL.Lock()
	chgCh := make(chan GoFunc, num)
	tmpCh := gp.workers
	gp.workers = chgCh
	gp.workerNum = num
	close(tmpCh)
	gp.chL.Unlock()

	for i := 0; i < num; i++ {
		go gp.worker()
	}
}

func (gp *GoPool) policy(needExpand bool) int {
	meta := gp.workerNum / (int(gp.reloadNum) + 1)
	return basic.IfElseInt(needExpand, gp.workerNum+meta, gp.workerNum-meta)
}

func emptyFunc(emptyPara interface{}) {}
