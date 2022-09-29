package geecache

//main thread

import (
	"GeeCache/day3-4-5-6-7/geecache/geecachepb"
	"GeeCache/day3-4-5-6-7/geecache/singleflight"
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
	getter    Getter     //getter is a callback func which loads data for a key from db
	mainCache cache      //cache stores data which has a special namespace
	peers     PeerPicker //pick a remote peer
	// use singleflight.Group to make sure that
	// each key is only fetched once
	loader *singleflight.Group
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
		loader:    &singleflight.Group{},
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

func (group *Group) load(key string) (value ByteView, err error) {
	//remote...
	viewi, err := group.loader.Do(key, func() (interface{}, error) {
		if group.peers != nil {
			if peer, ok := group.peers.PickPeer(key); ok {
				value, err := group.getFromPeer(peer, key)
				if err == nil {
					return value, nil
				}
				log.Println("[GeeCache] Failed to get from peer", err)
			}
		}
		//local:group is local
		fmt.Println("hitabc")
		return group.getLocally(key)
	})
	if err == nil {
		return viewi.(ByteView), nil
	}
	return
}

func (group *Group) getFromPeer(peer PeerGetter, key string) (ByteView, error) {
	res := &geecachepb.Request{Group: group.name, Key: key}
	reply := &geecachepb.Response{}
	err := peer.Get(res, reply)
	if err != nil {
		return ByteView{}, err
	}
	return ByteView{b: reply.Value}, nil
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

func (group *Group) RegisterPeers(peers PeerPicker) {
	if group.peers != nil {
		panic("RegisterPeer called more than once")
	}
	group.peers = peers
}
