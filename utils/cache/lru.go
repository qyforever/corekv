package cache

import (
	"container/list"
	"fmt"
)

type windowLRU struct {
	data map[uint64]*list.Element //
	cap  int                      //
	list *list.List
}

type storeItem struct {
	stage    int
	key      uint64
	conflict uint64 //辅助冲突判断
	value    interface{}
}

func newWindowLRU(size int, data map[uint64]*list.Element) *windowLRU {
	return &windowLRU{
		data: data,
		cap:  size,
		list: list.New(),
	}
}

func (lru *windowLRU) add(newitem storeItem) (eitem storeItem, evicted bool) {
	//implement me here!!!
	if lru.list.Len() < lru.cap {
		lru.data[newitem.key] = lru.list.PushFront(&newitem) // 节点插到头部
		return storeItem{}, false
	}

	evictItem := lru.list.Back() //尾部元素
	item := evictItem.Value.(*storeItem)

	delete(lru.data, item.key) // 删除尾部元素
	eitem, *item = *item, newitem
	lru.data[item.key] = evictItem
	lru.list.MoveToFront(evictItem) //把新数据放到头部
	return eitem, true
}

func (lru *windowLRU) get(v *list.Element) {
	//implement me here!!!
	lru.list.MoveToFront(v) //移动到链表最前端
}
func (lru *windowLRU) String() string {
	var s string
	for e := lru.list.Front(); e != nil; e = e.Next() {
		s += fmt.Sprintf("%v,", e.Value.(*storeItem).value)
	}
	return s
}
