package main

import (
	"container/heap"
	"sync"
)

type Key struct {
	Name string
	Hits int
}

// KeyHeap keeps track of hot keys and pops them off ordered by hotness,
// greatest hotness first.
type KeyHeap []*Key

func (h KeyHeap) Len() int { return len(h) }

// Less sorts in reverse order so we will pop the hottest keys first
// (i.e. we use > rather than <).
func (h KeyHeap) Less(i, j int) bool { return h[i].Hits > h[j].Hits }

func (h KeyHeap) Swap(i, j int) { h[i], h[j] = h[j], h[i] }

func (h *KeyHeap) Push(x interface{}) {
	*h = append(*h, x.(*Key))
}

func (h *KeyHeap) Pop() interface{} {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[0 : n-1]
	return x
}

type HotKeyPool struct {
	Lock sync.Mutex

	// Map of keys to hits
	items map[string]int
}

func NewHotKeyPool() *HotKeyPool {
	h := &HotKeyPool{}
	h.items = make(map[string]int)
	return h
}

// Add adds a new key to the hit counter or increments the key's hit counter
// if it is already present.
func (h *HotKeyPool) Add(keys []string) {
	h.Lock.Lock()
	defer h.Lock.Unlock()

	for _, key := range keys {
		if _, ok := h.items[key]; ok {
			h.items[key] += 1
		} else {
			h.items[key] = 1
		}

	}
}

// GetTopKeys returns a KeyHeap object.  Keys can be popped from the
// resulting object and will be ordered by hits, descending.
func (h *HotKeyPool) GetTopKeys() *KeyHeap {
	h.Lock.Lock()
	defer h.Lock.Unlock()

	top_keys := &KeyHeap{}
	heap.Init(top_keys)

	for key, hits := range h.items {
		heap.Push(top_keys, &Key{key, hits})
	}
	return top_keys
}

func (h *HotKeyPool) GetHits(key string) int {
	h.Lock.Lock()
	defer h.Lock.Unlock()
	return h.items[key]
}

// Rotate clears the data on the existing HotKeyPool, returning a new pool
// containing the old data.  This allows sorting and reporting to happen in
// another goroutine, while counting can continue on new keys.
func (h *HotKeyPool) Rotate() *HotKeyPool {
	h.Lock.Lock()
	defer h.Lock.Unlock()

	// Clone existing
	new_hot_key_pool := NewHotKeyPool()
	new_hot_key_pool.items = h.items

	// Clear existing values
	h.items = make(map[string]int)
	return new_hot_key_pool
}
