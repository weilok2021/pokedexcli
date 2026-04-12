package pokecache

import (
	"time"
	"sync"
)

type Cache struct {
	entry map[string]cacheEntry
	mu sync.Mutex
	interval time.Duration
}

type cacheEntry struct {
	createdAt time.Time
	val []byte 
}

func NewCache(interval time.Duration) *Cache {
	c := Cache{}
	c.entry = make(map[string]cacheEntry)
	c.interval = interval

	ticker := time.NewTicker(interval)

	// Goroutine that running continuosly in this program to clean up old cache entry
	go func () {
		for range ticker.C {
			// Iterate over all cache entries
			c.mu.Lock()
			for key, entry := range c.entry {
				if time.Since(entry.createdAt) > interval {
					delete(c.entry, key)
				}
			}
			c.mu.Unlock()
		}
	}()
	return &c
}

func (c *Cache) Add(key string, val []byte) {
	newVal := make([]byte, len(val))
	copy(newVal, val)
	value := cacheEntry{createdAt: time.Now(), val: newVal}
	c.mu.Lock()
	c.entry[key] = value
	c.mu.Unlock()
}

func (c *Cache) Get(key string) ([]byte, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if entry, ok := c.entry[key]; ok {
		return entry.val, true
	}
	return nil, false
}
