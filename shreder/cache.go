package shreder

import (
	"container/list"
	"github.com/rs/zerolog/log"
	"sync"
	"time"
)

type CacheItem struct {
	Value      string
	TimeToLive time.Time
}

type entry struct {
	key   string
	value CacheItem
}

type Cache struct {
	mu       sync.RWMutex
	Items    map[string]*list.Element
	eviction *list.List
	capacity int
}

func (c *Cache) Get(key string) (string, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	item, found := c.Items[key]
	if !found || time.Now().Local().After(item.Value.(entry).value.TimeToLive) {
		log.Info().Msgf("Cache miss for key %v", key)
		if found {
			c.eviction.Remove(item)
			delete(c.Items, key)
		}

		return "", false
	}

	c.eviction.MoveToFront(item)

	return item.Value.(entry).value.Value, true
}

func (c *Cache) Set(key string, value string, ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if item, found := c.Items[key]; found {
		c.eviction.Remove(item)
		delete(c.Items, key)
	}

	if c.eviction.Len() >= c.capacity {
		log.Info().Msg("Cache is full")
		c.evictLRU()
	}

	log.Info().Msgf("cache capacity  %d ", c.capacity)
	log.Info().Msgf("cache size  %d ", c.eviction.Len())

	item := CacheItem{
		Value:      value,
		TimeToLive: time.Now().Local().Add(ttl),
	}

	elem := c.eviction.PushFront(entry{key, item})
	c.Items[key] = elem

}

func (c *Cache) startEvictionWorker(d time.Duration) {
	ticker := time.NewTicker(d)
	go func() {
		for range ticker.C {
			c.evictExpiredItems()
		}
	}()
}

func (c *Cache) evictExpiredItems() {
	c.mu.Lock()
	defer c.mu.Unlock()
	now := time.Now().Local()

	for key, elem := range c.Items {
		if now.After(elem.Value.(entry).value.TimeToLive) {
			c.eviction.Remove(elem)
			delete(c.Items, key)
		}
	}

}

func (c *Cache) evictLRU() {
	log.Info().Msg("Evicting LRU item")
	elem := c.eviction.Back()
	if elem != nil {
		c.eviction.Remove(elem)
		kv := elem.Value.(entry)
		delete(c.Items, kv.key)

	}
}

func NewCache(capacity int) *Cache {
	c := &Cache{
		Items:    make(map[string]*list.Element),
		capacity: capacity,
		eviction: list.New(),
	}
	c.startEvictionWorker(5 * time.Second)
	return c
}
