package mcache

import (
	"container/list"
	"sync"
)

// LRU
// TODO LFU
type mcacheDO struct {
	omap *sync.Map // 元数据字典
	list *list.List
}

type dataNode struct {
	dataName string
	expire   int64
	num      int // 使用次数
	preNode  *dataNode
	nextNode *dataNode
}

func NewMcache() *mcacheDO {
	return &mcacheDO{
		omap: &sync.Map{},
		list: list.New(),
	}
}

// TODO 过期时间
func (m *mcacheDO) Set(key string, val interface{}, expire ...int64) {
	m.omap.Store(key, val)

	// nd := &dataNode{
	// 	dataName: key,
	// }
}

func (m *mcacheDO) Get(key string) interface{} {
	return nil
}
