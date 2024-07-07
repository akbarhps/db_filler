package main

import (
	"fmt"
	"math/rand"
	"sync"
)

func GetCacheKey(key string, value string) string {
	return fmt.Sprintf("%s#%s", key, value)
}

type ConcurrentCache struct {
	sync.RWMutex
	items      map[string][]string
	itemsIndex map[string]uint
}

func NewConcurrentCache() *ConcurrentCache {
	return &ConcurrentCache{
		items:      make(map[string][]string),
		itemsIndex: make(map[string]uint),
	}
}

func (cc *ConcurrentCache) ItemExists(key string) bool {
	cc.Lock()
	defer cc.Unlock()

	_, ok := cc.items[key]
	return ok
}

func (cc *ConcurrentCache) Add(key string, value *string) {
	cc.Lock()
	defer cc.Unlock()

	_, ok := cc.items[key]
	if !ok {
		cc.items[key] = make([]string, 0)
		cc.itemsIndex[key] = 0
	}

	if value == nil {
		return
	}

	cc.items[key] = append(cc.items[key], *value)
}

func (cc *ConcurrentCache) ClearCacheIndex() {
	cc.Lock()
	defer cc.Unlock()

	for key := range cc.itemsIndex {
		cc.itemsIndex[key] = 0
	}
}

func (cc *ConcurrentCache) GetRandom(key string) string {
	cc.Lock()
	defer cc.Unlock()

	cache, ok := cc.items[key]
	if !ok {
		panic(fmt.Sprintln("Cache doesn't have key: ", key))
	}

	cacheSize := len(cache)
	if cacheSize == 0 {
		panic(fmt.Sprintln("Cache doesn't have value for key: ", key))
	}

	return cache[rand.Intn(cacheSize-1)]
}

// Pull return values from cache in order
func (cc *ConcurrentCache) Pull(key string) string {
	cc.Lock()
	defer cc.Unlock()

	cache, ok := cc.items[key]
	if !ok {
		panic(fmt.Sprintln("Cache doesn't have key: ", key))
	}

	cacheSize := len(cache)
	if cacheSize == 0 {
		panic(fmt.Sprintln("Cache doesn't have value for key: ", key))
	}

	index, _ := cc.itemsIndex[key]
	cc.itemsIndex[key] += 1
	return cache[index]
}
