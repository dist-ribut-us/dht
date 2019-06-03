package dht

import (
  "sync"
)

type blacklist struct {
	Map map[string]bool
	sync.RWMutex
}

func newblacklist() *blacklist {
	return &blacklist{
		Map: make(map[string]bool),
	}
}

func (t *blacklist) get(key string) (bool, bool) {
	t.RLock()
	k, b := t.Map[key]
	t.RUnlock()
	return k, b
}

func (t *blacklist) set(key string, val bool) {
	t.Lock()
	t.Map[key] = val
	t.Unlock()
}

func (t *blacklist) delete(keys ...string) {
	t.Lock()
	for _, key := range keys {
		delete(t.Map, key)
	}
	t.Unlock()
}


