package core

import (
	"fmt"
	"log"
	"sync"
)

var (
	mu     sync.RWMutex
	groups = make(map[string]*Group)
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

type Group struct {
	name      string
	getter    Getter
	mainCache cache
}

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

func GetGroup(name string) *Group {
	mu.RLock()
	g := groups[name]
	mu.RUnlock()
	return g
}

func (g *Group) Get(key string) (b ByteView, err error) {
	if key == "" {
		return ByteView{}, fmt.Errorf("key is required")
	}

	// 命中缓存
	if value, ok := g.mainCache.get(key); ok {
		log.Println("cache hit")
		return value, nil
	}

	return g.load(key)
}

func (g *Group) load(key string) (ByteView, error) {
	value, err := g.getter.Get(key)
	if err != nil {
		return ByteView{}, err
	}
	v := ByteView{b: cloneBytes(value)}
	g.mainCache.add(key, v)
	return v, nil
}
