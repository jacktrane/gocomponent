package gopool

import (
	"io"
	"sync"

	"github.com/jacktrane/gocomponent/errmsg"
	"github.com/jacktrane/gocomponent/logger"
)

type ConnPool struct {
	// container *sync.Pool
	m              *sync.Mutex               // 保证多个goroutine访问时候，closed的线程安全
	conn           chan io.Closer            //连接存储的chan
	factory        func() (io.Closer, error) //新建连接的工厂方法
	closed         bool                      //连接池关闭标志
	maxConnNum     int                       // 最大连接数
	maxIdleConnNum int                       // 最大闲置连接数
	desc           string                    // 描述
}

func NewConnPool() *ConnPool {
	return nil
}

// 获取链接
func (cp *ConnPool) GetConn() (io.Closer, error) {
	select {
	case r, ok := <-cp.conn:
		logger.Debug(cp.desc, "get a conn")
		if !ok {
			return nil, errmsg.ErrConnPoolClosed
		}
		return r, nil
	default:
		logger.Debug(cp.desc, "new init conn")
		return cp.factory()
	}
}

// 还掉链接
func (cp *ConnPool) PutConn(r io.Closer) {
	//保证该操作和Close方法的操作是安全的
	cp.m.Lock()
	defer cp.m.Unlock()

	//资源池都关闭了，就省这一个没有释放的资源了，释放即可
	if cp.closed {
		r.Close()
		return
	}

	select {
	case cp.conn <- r:
		logger.Debug(cp.desc, "put conn in pool")
	default:
		logger.Debug(cp.desc, "conn pool full")
		r.Close()
	}

}

// 关闭链接
func (cp *ConnPool) Close() {
	cp.m.Lock()
	defer cp.m.Unlock()

	if cp.closed {
		return
	}

	cp.closed = true

	//关闭通道，不让写入了
	close(cp.conn)

	//关闭通道里的资源
	for r := range cp.conn {
		r.Close()
	}
}
