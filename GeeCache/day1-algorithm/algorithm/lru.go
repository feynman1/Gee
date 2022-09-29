package algorithm

import (
	"container/list"
)

type LRU struct {
	capacity int64
	uBytes   int64
	ll       *list.List
	cache    map[string]*list.Element
	// optional and executed when an entry is purged.
	OnEvicted func(key string, value Value)
}

//-------------------------utils-------------------------

func (lru *LRU) removeOldest() {
	ele := lru.ll.Back()
	if ele != nil {
		lru.ll.Remove(ele)
		entry := ele.Value.(*entry)
		delete(lru.cache, entry.key)
		lru.uBytes -= int64(len(entry.key)) + int64(entry.value.Len())
		return
	}
}

func (c *LRU) Len() int {
	return c.ll.Len()
}

//-------------------------service-------------------------

func New(capacity int64, onEvicted func(string, Value)) *LRU {
	return &LRU{
		capacity:  capacity,
		ll:        list.New(),
		cache:     make(map[string]*list.Element),
		OnEvicted: onEvicted,
	}
}

func (lru *LRU) Get(key string) (value Value, ok bool) {
	if ele, ok := lru.cache[key]; ok {
		lru.ll.MoveToFront(ele)
		entry := ele.Value.(*entry)
		return entry.value, true
	}
	return
}

func (lru *LRU) Add(key string, value Value) {
	if ele, ok := lru.cache[key]; ok {
		//update
		lru.ll.MoveToFront(ele)
		ey := ele.Value.(*entry)
		lru.uBytes += int64(value.Len()) - int64(ey.value.Len())
		ey.value = value
	} else {
		//insert
		newEntry := &entry{
			key:   key,
			value: value,
		}
		newEle := lru.ll.PushFront(newEntry)
		lru.cache[key] = newEle
		lru.uBytes += int64(len(key)) + int64(value.Len())
	}
	for lru.capacity != 0 && lru.capacity < lru.uBytes {
		lru.removeOldest()
	}
}
