package mcache

import (
	"container/list"
	"errors"
	"sync"

	"github.com/jacktrane/gocomponent/time_format"
)

// 优先做一个LRU吧，使用双向链表
// TODO LFU
type mcacheDO struct {
	omap *sync.Map // 元数据字典
	list *list.List
}

type ValueDO struct {
	oValue interface{}
	expire int64
	ele    *list.Element
}

func NewMcache() *mcacheDO {
	return &mcacheDO{
		omap: &sync.Map{},
		list: list.New(),
	}
}

func (m *mcacheDO) Set(key, val interface{}, expire ...int64) {
	var v ValueDO
	v.oValue = val
	if len(expire) != 0 {
		v.expire = expire[0]
	}

	value, existed := m.omap.Load(key)
	if existed {
		actValue := value.(ValueDO)
		actValue.oValue = val
		m.omap.Store(key, actValue)
		m.list.MoveToFront(value.(ValueDO).ele)
		return
	}

	v.ele = m.list.PushFront(val)
	m.omap.Store(key, v)
}

func (m *mcacheDO) Get(key interface{}) (interface{}, error) {
	value, ok := m.omap.Load(key)
	if !ok {
		return nil, errors.New("value no existed")
	}

	// 清过期的值
	if value.(ValueDO).expire < time_format.GetTimestamp() {
		m.omap.Delete(key)
		m.list.Remove(value.(ValueDO).ele)
		return nil, errors.New("value expire")
	}

	m.list.MoveToFront(value.(ValueDO).ele)
	return value.(ValueDO).oValue, nil
}

func (m *mcacheDO) Delete(key interface{}) {
	value, ok := m.omap.Load(key)
	if !ok {
		return
	}
	m.omap.Delete(key)
	m.list.Remove(value.(ValueDO).ele)
}
