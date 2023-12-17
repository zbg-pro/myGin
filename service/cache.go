package service

import (
	"fmt"
	"sync"
	"time"
)

type CacheItem struct {
	value      interface{}
	expiration time.Time
}

type CacheMap struct {
	mu           sync.Mutex
	cache        map[string]CacheItem
	defaultTTL   time.Duration
	autoExpire   bool
	supplyMethod func(key string) interface{}
	maxKeys      int
}

func NewCacheMap(defaultTTL time.Duration, autoExpire bool, supplyMethod func(key string) interface{}, maxKeys int) *CacheMap {
	return &CacheMap{
		cache:        make(map[string]CacheItem),
		defaultTTL:   defaultTTL,
		autoExpire:   autoExpire,
		supplyMethod: supplyMethod,
		maxKeys:      maxKeys,
	}
}

func (c *CacheMap) Get(key string) interface{} {
	c.mu.Lock()
	defer c.mu.Unlock()

	item, found := c.cache[key]
	if !found {
		if c.supplyMethod != nil {
			value := c.supplyMethod(key)
			if len(c.cache) >= c.maxKeys {
				c.evictOldest()
			}
			c.cache[key] = CacheItem{value: value, expiration: time.Now().Add(c.defaultTTL)}
			return value
		}
		return nil
	}

	if c.autoExpire && time.Now().After(item.expiration) {
		delete(c.cache, key)
		if c.supplyMethod != nil {
			value := c.supplyMethod(key)
			if len(c.cache) >= c.maxKeys {
				c.evictOldest()
			}
			c.cache[key] = CacheItem{value: value, expiration: time.Now().Add(c.defaultTTL)}
			return value
		}
		return nil
	}

	return item.value
}

func (c *CacheMap) evictOldest() {
	oldestKey := ""
	oldestExpiration := time.Now().Add(c.defaultTTL)

	for key, item := range c.cache {
		if item.expiration.Before(oldestExpiration) {
			oldestKey = key
			oldestExpiration = item.expiration
		}
	}

	delete(c.cache, oldestKey)
}

func main() {
	// Example usage
	cache := NewCacheMap(10*time.Second, true, func(key string) interface{} {
		// Supply method for generating a new value
		fmt.Println("Generating value for key:", key)
		return key + "_value"
	}, 3) // Set maxKeys to 3 for this example

	fmt.Println(cache.Get("key1"))
	fmt.Println(cache.Get("key2"))
	fmt.Println(cache.Get("key3"))
	fmt.Println(cache.Get("key4"))
	fmt.Println(cache.Get("key1")) // This will return the cached value for key1
}

//在这个示例中，maxKeys 参数限制了缓存键的数量，当缓存键数量达到上限时，会通过 evictOldest 方法替换最老的值。你可以根据实际需求调整 maxKeys 的值。
