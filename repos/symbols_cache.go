package repos

import (
	"strings"
	"sync"
)

type SymbolCache struct {
	c map[string]struct{}
	m *sync.RWMutex
}

func NewSymbolCache() *SymbolCache {
	return &SymbolCache{
		c: make(map[string]struct{}),
		m: new(sync.RWMutex),
	}
}

func (sc *SymbolCache) GetAll() []string {
	var ret = make([]string, 0, len(sc.c))

	sc.m.RLock()
	defer sc.m.RUnlock()
	for k := range sc.c {
		ret = append(ret, k)
	}

	return ret
}

func (sc *SymbolCache) Add(s string) {
	sc.m.Lock()
	defer sc.m.Unlock()
	sc.c[strings.ToUpper(s)] = struct{}{}
}

func (sc *SymbolCache) Remove(s string) {
	sc.m.Lock()
	defer sc.m.Unlock()
	delete(sc.c, strings.ToUpper(s))
}

func (sc *SymbolCache) IsPresent(s string) bool {
	sc.m.RLock()
	defer sc.m.RUnlock()
	_, ok := sc.c[strings.ToUpper(s)]
	return ok
}
