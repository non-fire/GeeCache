package geecache

import (
	"fmt"
	"log"
	"sync"
)

/*
	var f Getter = GetterFunc(func(key string) ([]byte, error) {
		return []byte(key), nil
	})
	v, ok := f.Get("key")

	define a func struct(GetterFunc) and implement a method(Get) of an interface(Getter):
	=> turn a function into an interface
*/
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

// a cache namespace
type Group struct{
	name string
	getter Getter // the callback func when cache miss
	mainCache cache // concurrent cache in cache.go
	peers PeerPicker // inject HttpPool(implement PeerPicker) into Group
}

var (
	mu sync.RWMutex
	groups = make(map[string]*Group)
)

// create a new group(namespace)
func NewGroup(name string, cacheBytes int64, getter Getter) *Group {
	if getter == nil {
		panic("getter is nil")
	}
	mu.Lock()
	defer mu.Unlock()
	g := &Group{
		name: name,
		getter: getter,
		mainCache: cache{cacheBytes: cacheBytes},
	}
	groups[name] = g // set mapping
	return g
}

func GetGroup(name string) *Group {
	mu.RLock() // read lock is ok
	g := groups[name]
	mu.RUnlock()
	return g
}

func (g *Group) RegisterPeers(peers PeerPicker) {
	if (g.peers != nil) {
		panic("Reregister PeerPicker!")
	}
	g.peers = peers
}

func (g *Group) Get(key string) (ByteView, error) {
	if key == "" { // empty key
		return ByteView{}, fmt.Errorf("key is needed")
	}
	// mainCache HIT
	if value, ok := g.mainCache.get(key); ok {
		log.Println("Geecache Hit")
		return value, nil
	}
	// mainCache MISS -> load locally / load from peer node (distributed situation)
	return g.load(key)
}

func (g *Group) load(key string) (value ByteView, err error) {
	if g.peers != nil {
		// not self node
		if peerGetter, ok := g.peers.PickPeer(key); ok {
			if value, err = g.getFromPeer(key, peerGetter); err == nil {
				return value, nil
			}
			log.Println("[GeeCache] Failed to get from peer", err)
		}
	}
	// self node / fail
	return g.getLocally(key)
}

// get cache value from remote peer by httpGetter(implement PeerGetter)
func (g *Group) getFromPeer(key string, peerGetter PeerGetter) (ByteView, error) {
	bytes, err := peerGetter.Get(g.name, key)
	if err == nil {
		return ByteView{}, err
	}
	return ByteView{bytes}, nil
}

func (g *Group) getLocally(key string) (ByteView, error) {
	bytes, err := g.getter.Get(key)
	if err != nil {
		return ByteView{}, err
	}
	// add to mainCache
	value := ByteView{b: cloneBytes(bytes)}
	g.populateCache(key, value)
	return value, nil
}

func (g *Group) populateCache(key string, value ByteView) {
	g.mainCache.add(key, value)
}