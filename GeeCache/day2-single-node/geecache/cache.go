package geecache

import (
	"GeeCache/day1-algorithm/algorithm"
	"sync"
)

type cache struct {
	lru        *algorithm.LRU
	mu         sync.Mutex
	cacheBytes int64
}

func (c *cache) add(key string, value ByteView) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.lru == nil {
		c.lru = algorithm.New(c.cacheBytes, nil)
	}
	c.lru.Add(key, value)
}

func (c *cache) get(key string) (value ByteView, ok bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.lru == nil {
		return //equal to "return value,ok"
	}
	if v, ok1 := c.lru.Get(key); ok1 {
		return v.(ByteView), ok1
	}
	return
}
