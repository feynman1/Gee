package geecache

import (
	"fmt"
	"log"
	"sync"
)

// A Getter loads data for a key.
type Getter interface {
	Get(key string) ([]byte, error)
}

// A GetterFunc implements Getter with a function.
type GetterFunc func(key string) ([]byte, error)

// Get implements Getter interface function
func (f GetterFunc) Get(key string) ([]byte, error) {
	return f(key)
}

//A Group is a cache namespace and associated data loaded spread over
type Group struct {
	name      string
	getter    Getter //getter is a callback func which loads data for a key from db
	mainCache cache  //cache stores data which has a special namespace
}

var (
	mu     sync.RWMutex
	groups = make(map[string]*Group)
)

// NewGroup create a new instance of Group
func NewGroup(name string, cacheBytes int64, getter Getter) *Group {
	if getter == nil {
		panic("nil Getter")
	}
	mu.Lock()
	defer mu.Unlock()
	g := &Group{
		name:      name,
		getter:    getter,
		mainCache: cache{cacheBytes: cacheBytes},
	}
	groups[name] = g
	return g
}

// GetGroup returns the named group previously created with NewGroup, or
// nil if there's no such group.
func GetGroup(name string) *Group {
	mu.RLock()
	defer mu.RUnlock()
	g := groups[name]
	return g
}

func (group *Group) Get(key string) (ByteView, error) {
	if key == "" {
		return ByteView{}, fmt.Errorf("key is required")
	}
	if v, ok := group.mainCache.get(key); ok {
		log.Println("[GeeCache] hit")
		return v, nil
	}
	return group.load(key)
}

func (group *Group) load(key string) (ByteView, error) {
	//remote...

	//local
	return group.getLocally(key)
}

func (group *Group) getLocally(key string) (ByteView, error) {
	bytes, err := group.getter.Get(key)
	if err != nil {
		return ByteView{}, err
	}
	value := ByteView{cloneByte(bytes)}
	group.populateCache(key, value)
	return value, nil
}

func (group *Group) populateCache(key string, value ByteView) {
	group.mainCache.add(key, value)
}