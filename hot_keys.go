package main

import (
	//	"fmt"
	"sync"
)

type Hotness struct {
	// The overall position relative to other keys
	position int
	// The number of hits the key has had in the current time period
	hits int
}

type HotKeyPool struct {
	Lock             sync.Mutex
	items            int
	keys_by_position []string
	hotness_by_key   map[string]*Hotness
}

func NewHotKeyPool() *HotKeyPool {
	h := &HotKeyPool{}
	h.items = 0
	h.keys_by_position = []string{}
	h.hotness_by_key = map[string]*Hotness{}
	return h
}

// NewHotKeyPoolFromExisting clones an existing HotKeyPool with a new lock.
// The caller should handle the locking on the pool that is to be cloned.
func NewHotKeyPoolFromExisting(existing *HotKeyPool) *HotKeyPool {
	h := &HotKeyPool{}
	h.items = existing.items
	h.keys_by_position = existing.keys_by_position
	h.hotness_by_key = existing.hotness_by_key
	return h
}

// Add adds a new key, incrementing its hit counter and updating its position
// in the top keys list.
func (h *HotKeyPool) Add(keys []string) {
	h.Lock.Lock()
	defer h.Lock.Unlock()

	for _, key := range keys {

		// Update hits
		if val, ok := h.hotness_by_key[key]; ok {
			val.hits += 1

		PositionUpdateLoop:
			for {

				// If the key is already the most hot, continue
				if val.position == 0 {
					break PositionUpdateLoop
				}

				// Keep moving the key up until it's reached another key that
				// has more hits
				higher_key := h.keys_by_position[val.position-1]
				if val.hits > h.hotness_by_key[higher_key].hits {
					h.keys_by_position[val.position-1], h.keys_by_position[val.position] = h.keys_by_position[val.position], h.keys_by_position[val.position-1]
					h.hotness_by_key[higher_key].position += 1
					h.hotness_by_key[key].position -= 1
				} else {
					break PositionUpdateLoop
				}

				val, _ = h.hotness_by_key[key]
			}

		} else {
			h.items += 1
			h.keys_by_position = append(h.keys_by_position, key)
			h.hotness_by_key[key] = &Hotness{h.items - 1, 1}
		}

	}
}

// GetTopKeys returns a list keys, ordered by number of hits, descending.
func (h *HotKeyPool) GetTopKeys() []string {
	h.Lock.Lock()
	defer h.Lock.Unlock()
	return h.keys_by_position
}

func (h *HotKeyPool) GetHits(key string) int {
	h.Lock.Lock()
	defer h.Lock.Unlock()
	return h.hotness_by_key[key].hits
}

// Rotate clears the data on the existing HotKeyPool, returning a new pool
// containing the old data.
func (h *HotKeyPool) Rotate() *HotKeyPool {
	h.Lock.Lock()
	defer h.Lock.Unlock()
	new_hot_key_pool := NewHotKeyPoolFromExisting(h)
	h.items = 0
	h.keys_by_position = []string{}
	h.hotness_by_key = map[string]*Hotness{}
	return new_hot_key_pool
}

/*

KEY_LIMIT and GROWTH_LIMIT

channel -> N size pool -> main pool

Messages get put into the first channel, pulled off and counted into N size
pool.  This pool is exactly the same as the main pool.  When it reaches X size,
it gets merged with main pool.

*/
