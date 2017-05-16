package util

import (
	"sync"
	"time"
)

type CacheItem struct {
	value interface{}
	updateTimeNano int64
}

type Cache struct {
	timeout   int64 // nanosecond
	items     map[interface{}]CacheItem
	lock      sync.RWMutex
}

func NewCache(timeoutSecond int) *Cache {
	cache := &Cache{ timeout: int64(timeoutSecond * 1000 * 1000 * 1000), items: map[interface{}]CacheItem{} }
	return cache
}

func (c *Cache) Put(key, value interface{}) {
	c.lock.Lock()
	defer c.lock.Unlock()

	now := time.Now().UnixNano()
	c.items[key] = CacheItem{value: value, updateTimeNano: now}
}

func (c *Cache) Get(key interface{})(value interface{},ok bool) {
	c.lock.RLock()
	defer c.lock.RUnlock()

	item, ok := c.items[key]
	if ok {
		now := time.Now().UnixNano()
		if now - item.updateTimeNano > c.timeout {
			ok = false
		}else{
			value = item.value
		}
	}
	return
}

func (c *Cache) GetEntries(key interface{}) (value interface{}, ok bool) {
	c.lock.RLock()
	defer c.lock.RUnlock()

	item, ok := c.items[key]
	if ok {
		now := time.Now().UnixNano()
		item.updateTimeNano = now
		value = item.value
		return
	} else {
		return nil, false
	}
}

func (c *Cache) PutEntries(key interface{}, value interface{}) {
	c.lock.Lock()
	defer c.lock.Unlock()

	now := time.Now().UnixNano()
	c.items[key] = CacheItem{value: value, updateTimeNano: now}
}