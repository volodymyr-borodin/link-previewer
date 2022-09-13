package main

type Cache struct {
	cache map[string]*PageMeta
}

func NewCache() *Cache {
	return &Cache{cache: map[string]*PageMeta{}}
}

func (c *Cache) get(key *string) (meta *PageMeta, ok bool) {
	meta, ok = c.cache[*key]
	return
}

func (c *Cache) set(key *string, meta *PageMeta) {
	c.cache[*key] = meta
}
