package gopool

type GoPool struct {
	goNum   int
	workers chan GoFunc
}

type GoFunc struct {
	f    func(para interface{})
	para interface{}
}

// goNum 协程池的内woker数量，必填
// chNum 管道大小
func NewGoPool(workerNum, chNum int) *GoPool {
	pool := GoPool{
		goNum:   workerNum,
		workers: make(chan GoFunc, chNum),
	}

	for i := 0; i < workerNum; i++ {
		go pool.worker()
	}

	return &pool
}

func (gp *GoPool) worker() {
	for worker := range gp.workers {
		worker.f(worker.para)
	}
}

func (gp *GoPool) AddGoFunc(f func(para interface{}), para interface{}) {
	gp.workers <- GoFunc{
		f:    f,
		para: para,
	}
}

func (gp *GoPool) Close() {
	close(gp.workers)
}
