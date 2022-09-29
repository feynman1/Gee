package singleflight

import "sync"

type call struct {
	wg  sync.WaitGroup
	val interface{}
	err error
}

type Group struct {
	mu sync.Mutex //protect m
	m  map[string]*call
}

func (g *Group) Do(key string, fn func() (interface{}, error)) (interface{}, error) {
	g.mu.Lock()
	if g.m == nil {
		g.m = make(map[string]*call)
	}
	if c, ok := g.m[key]; ok {
		g.mu.Unlock()
		c.wg.Wait()
		return c.val, c.err
	}
	newcall := new(call)
	newcall.wg.Add(1)
	g.m[key] = newcall
	g.mu.Unlock()

	newcall.val, newcall.err = fn()
	newcall.wg.Done()

	g.mu.Lock()
	delete(g.m, key)
	g.mu.Unlock()

	return newcall.val, newcall.err
}
