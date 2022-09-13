package main

import (
	"log"
	"time"
)

type Cache struct {
	cache   map[string]*CacheItem
	ticker  *time.Ticker
	options *CacheOptions
}

type Now func() time.Time

type CacheItem struct {
	Item       *PageMeta
	Expiration time.Time
}

type CacheOptions struct {
	DefaultExpiration    time.Duration
	InvalidationInterval time.Duration
	Now                  Now
}

func DefaultOptions() *CacheOptions {
	return &CacheOptions{
		DefaultExpiration:    24 * time.Hour,
		InvalidationInterval: 1 * time.Hour,
		Now:                  time.Now,
	}
}

func NewCache(options *CacheOptions) *Cache {
	cache := &Cache{
		cache:   map[string]*CacheItem{},
		ticker:  time.NewTicker(options.InvalidationInterval),
		options: options,
	}

	go cache.invalidate()

	return cache
}

func (c *Cache) get(key *string) (meta *PageMeta, ok bool) {
	item, ok := c.cache[*key]
	if !ok {
		return nil, false
	}

	if c.options.Now().After(item.Expiration) {
		delete(c.cache, *key)
		return nil, false
	}

	return item.Item, true
}

func (c *Cache) set(key *string, meta *PageMeta) {
	c.cache[*key] = &CacheItem{
		Item:       meta,
		Expiration: c.options.Now().Add(c.options.DefaultExpiration),
	}
}

func (c *Cache) invalidate() {
	for {
		<-c.ticker.C
		log.Printf("Cache invalidation started")
		now := c.options.Now()
		for key, value := range c.cache {
			if now.After(value.Expiration) {
				delete(c.cache, key)
			}
		}

		log.Printf("Cache invalidation finished")
	}
}
