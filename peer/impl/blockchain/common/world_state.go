package common

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"
	"sync"
)

type WorldState struct {
	mu sync.Mutex
	m  map[string]State
}

func NewWorldState() WorldState {
	return WorldState{
		mu: sync.Mutex{},
		m:  make(map[string]State),
	}
}

func (m *WorldState) Len() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.m)
}

func (m *WorldState) Get(key string) (value State, ok bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	value, ok = m.m[key]
	return
}

func (m *WorldState) Set(key string, value State) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.m[key] = value
	return
}

func (m *WorldState) Delete(key string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.m, key)
	return
}

func (m *WorldState) Copy() *WorldState {
	m.mu.Lock()
	defer m.mu.Unlock()

	cp := NewWorldState()
	for k, v := range m.m {
		cp.Set(k, v)
	}

	return &cp
}

func (m *WorldState) Hash() []byte {
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

func (m *WorldState) HashCode() string {
	return hex.EncodeToString(m.Hash())
}

func (m *WorldState) GetSimpleMap() map[string]State {
	m.mu.Lock()
	defer m.mu.Unlock()

	m2 := make(map[string]State)

	for k, v := range m.m {
		m2[k] = v
	}
	return m2
}

func (m *WorldState) Keys() []string {
	m.mu.Lock()
	defer m.mu.Unlock()
	keys := make([]string, 0, len(m.m))
	for k := range m.m {
		keys = append(keys, k)
	}
	return keys
}

func (m *WorldState) Equal(other *WorldState) bool {
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
		if !ok || !v.Equals(m.m[k]) {
			return false
		}
	}

	return true
}

func QuickWorldState(accounts int, balance int64) *WorldState {
	worldState := NewWorldState()
	for i := 0; i < accounts; i++ {
		worldState.Set(fmt.Sprintf("%d", i+1), State{
			Nonce:       0,
			Balance:     balance,
			CodeHash:    "",
			StorageRoot: "",
			Tasks:       make(map[string][2]string),
		})
	}
	return &worldState
}
