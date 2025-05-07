package cache

import (
	"strings"
	"sync"
)

type Cache struct {
	c map[string]struct{}
	m *sync.RWMutex
}

func New() *Cache {
	return &Cache{
		c: make(map[string]struct{}),
		m: new(sync.RWMutex),
	}
}

func (c *Cache) GetAll() []string {
	var ret = make([]string, 0, len(c.c))

	c.m.RLock()
	defer c.m.RUnlock()
	for k := range c.c {
		ret = append(ret, k)
	}

	return ret
}

func (c *Cache) Add(s string) {
	c.m.Lock()
	defer c.m.Unlock()
	c.c[strings.ToUpper(s)] = struct{}{}
}

func (c *Cache) Remove(s string) {
	c.m.Lock()
	defer c.m.Unlock()
	delete(c.c, strings.ToUpper(s))
}

func (c *Cache) IsPresent(s string) bool {
	c.m.RLock()
	defer c.m.RUnlock()
	_, ok := c.c[strings.ToUpper(s)]

	return ok
}
