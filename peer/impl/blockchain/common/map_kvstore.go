package common

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"sort"
	"sync"
)

// mapKVStore implements KVStore
type mapKVStore[TV comparable] struct {
	mu sync.Mutex
	m  map[string]TV
}

func NewKVStore[TV comparable]() KVStore[TV] {
	return &mapKVStore[TV]{
		mu: sync.Mutex{},
		m:  make(map[string]TV),
	}
}

func (m *mapKVStore[TV]) Len() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.m)
}

func (m *mapKVStore[TV]) Get(key string) (value TV, ok bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	value, ok = m.m[key]
	return
}

func (m *mapKVStore[TV]) Set(key string, value TV) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.m[key] = value
	return
}

func (m *mapKVStore[TV]) Delete(key string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.m, key)
	return
}

func (m *mapKVStore[TV]) Copy() KVStore[TV] {
	m.mu.Lock()
	defer m.mu.Unlock()

	cp := NewKVStore[TV]()
	for k, v := range m.m {
		cp.Set(k, v)
	}

	return cp
}

func (m *mapKVStore[TV]) Hash() []byte {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Sort the keys
	keys := make([]string, 0, len(m.m))
	for k, _ := range m.m {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	h := sha256.New()
	for _, k := range keys {
		v, _ := m.m[k]
		kv, _ := json.Marshal(v)
		h.Write([]byte(k))
		h.Write(kv)
	}

	return h.Sum(nil)
}

func (m *mapKVStore[TV]) HashCode() string {
	return hex.EncodeToString(m.Hash())
}

func (m *mapKVStore[TV]) ForEach(fn func(key string, value TV) bool) bool {
	//TODO implement me
	panic("implement me")
}

func (m *mapKVStore[TV]) GetSimpleMap() map[string]TV {
	m.mu.Lock()
	defer m.mu.Unlock()

	m2 := make(map[string]TV)

	for k, v := range m.m {
		m2[k] = v
	}
	return m2
}

func (m *mapKVStore[TV]) Keys() []string {
	m.mu.Lock()
	defer m.mu.Unlock()
	keys := make([]string, 0, len(m.m))
	for k := range m.m {
		keys = append(keys, k)
	}
	return keys
}

func (m *mapKVStore[TV]) Equal(other *KVStore[TV]) bool {
	m.mu.Lock()
	defer m.mu.Unlock()

	if len(m.m) != (*other).Len() {
		return false
	}

	keys := make([]string, 0, len(m.m))
	for k := range m.m {
		keys = append(keys, k)
	}

	for _, k := range keys {
		v, ok := (*other).Get(k)
		if !ok || v != m.m[k] {
			return false
		}
	}

	return true
}
