package algorithm

import "container/list"

type LFU struct {
	cache    map[string]*list.Element
	k2FMap   map[string]int     //key->frequency
	f2KMap   map[int]*list.List //frequency->list of key
	minFreq  int
	capacity int64 //maxBytes
	uBytes   int64
	// optional and executed when an entry is purged.
	OnEvicted func(key string, value Value)
	len       int
}

//-------------------------utils-------------------------

func (lfu *LFU) remove() {
	list := lfu.f2KMap[lfu.minFreq]
	ele := list.Back()
	if ele != nil {
		//update cache , kf , fk
		list.Remove(ele)
		if list.Len() == 0 {
			delete(lfu.f2KMap, lfu.minFreq)
		}
		ey := ele.Value.(*entry)
		delete(lfu.cache, ey.key)
		delete(lfu.k2FMap, ey.key)
		//update len, uBytes , minFreq
		lfu.uBytes -= int64(len(ey.key)) + int64(ey.value.Len())
		lfu.len--
		i := 0
		fq := 0
		for freq, _ := range lfu.f2KMap {
			if i == 0 {
				fq = freq
			}
			if freq < fq {
				fq = freq
			}
			i++
		}
		lfu.minFreq = fq
		return
	}
}

func (lfu *LFU) Len() int {
	return lfu.len
}

func (lfu *LFU) UpdateFreq(key string, ele *list.Element) {
	oldFre := lfu.k2FMap[key]
	//update k2f , f2k and lfu.minFreq
	lfu.k2FMap[key]++
	lfu.f2KMap[oldFre].Remove(ele)
	if lfu.f2KMap[oldFre].Len() == 0 {
		delete(lfu.f2KMap, oldFre)
		if oldFre == lfu.minFreq {
			lfu.minFreq = oldFre + 1
		}
	}
	//insert ele into new list
	if _, ok := lfu.f2KMap[oldFre+1]; !ok {
		lfu.f2KMap[oldFre+1] = &list.List{}
	}
	lfu.f2KMap[oldFre+1].PushFront(ele)
}

//-------------------------service-------------------------

func NewLFU(capacity int64, onEvicted func(string, Value)) *LFU {
	return &LFU{
		cache:     make(map[string]*list.Element),
		k2FMap:    make(map[string]int),
		f2KMap:    make(map[int]*list.List),
		minFreq:   0,
		capacity:  capacity,
		uBytes:    0,
		OnEvicted: onEvicted,
		len:       0,
	}
}

func (lfu *LFU) Get(key string) (value Value, ok bool) {
	if ele, ok := lfu.cache[key]; ok {
		entry := ele.Value.(*entry)
		lfu.UpdateFreq(key, ele)
		return entry.value, true
	}
	return
}

func (lfu *LFU) Add(key string, value Value) {
	if ele, ok := lfu.cache[key]; ok {
		lfu.UpdateFreq(key, ele)
		ey := ele.Value.(*entry)
		lfu.uBytes += int64(value.Len()) - int64(ey.value.Len())
		ey.value = value
	} else {
		//insert
		newEntry := &entry{
			key:   key,
			value: value,
		}
		lfu.k2FMap[key] = 1
		if _, ok := lfu.f2KMap[1]; !ok {
			lfu.f2KMap[1] = &list.List{}
		}
		newEle := lfu.f2KMap[1].PushFront(newEntry)
		lfu.cache[key] = newEle
		//update len , minFreq , uBytes
		lfu.uBytes += int64(len(key)) + int64(value.Len())
		lfu.minFreq = 1
		lfu.len++
	}
	for lfu.capacity != 0 && lfu.capacity < lfu.uBytes {
		lfu.remove()
	}
}
