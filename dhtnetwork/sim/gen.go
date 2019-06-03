package sim

import (
  "sync"
)

type waiting struct {
	Map map[string]seekResponseHandler
	sync.RWMutex
}

func newwaiting() *waiting {
	return &waiting{
		Map: make(map[string]seekResponseHandler),
	}
}

func (t *waiting) get(key string) (seekResponseHandler, bool) {
	t.RLock()
	k, b := t.Map[key]
	t.RUnlock()
	return k, b
}

func (t *waiting) set(key string, val seekResponseHandler) {
	t.Lock()
	t.Map[key] = val
	t.Unlock()
}

func (t *waiting) delete(keys ...string) {
	t.Lock()
	for _, key := range keys {
		delete(t.Map, key)
	}
	t.Unlock()
}


