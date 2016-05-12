package mcsauna

import "fmt"

type Hotness struct {
	// The overall position relative to other keys
	position int
	// The number of hits the key has had in the current time period
	hits int
}

type HotKeyPool struct {
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

func (h *HotKeyPool) Add(keys []string) {
	for _, key := range keys {

		// Update hits
		if val, ok := h.hotness_by_key[key]; ok {
			val.hits += 1

			// If the key is already the most hot, continue
			if val.position == 0 {
				continue
			}

			// Update position
			higher_key := h.keys_by_position[val.position-1]
			if val.hits > h.hotness_by_key[higher_key].hits {
				fmt.Printf("key: %v, higher_key: %v, keys_by_position: %d, val.position: %d\n", key, higher_key, len(h.keys_by_position), val.position)
				h.keys_by_position[val.position-1], h.keys_by_position[val.position] = h.keys_by_position[val.position], h.keys_by_position[val.position-1]
				h.hotness_by_key[higher_key].position += 1
				h.hotness_by_key[key].position -= 1
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
	return h.keys_by_position
}

/*

KEY_LIMIT and GROWTH_LIMIT


channel -> N size pool -> main pool

Messages get put into the first channel, pulled off and counted into N size
pool.  This pool is exactly the same as the main pool.  When it reaches X size,
it gets merged with main pool.

For clearing out the pools, there is a simple lock that returns the old data

*/
