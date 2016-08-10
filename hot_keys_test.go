package main

import (
	"container/heap"
	"testing"
)

func stringsEqual(first []string, second []string) bool {
	if len(first) != len(second) {
		return false
	}
	for i := range first {
		if first[i] != second[i] {
			return false
		}
	}
	return true
}

func TestHotKeys(t *testing.T) {
	h := NewHotKeyPool()
	h.Add([]string{"foo", "foo", "baz", "baz", "bar", "baz"})
	top_keys := h.GetTopKeys()
	expected_keys := []Key{
		Key{"baz", 3},
		Key{"foo", 2},
		Key{"bar", 1},
	}
	for _, key := range expected_keys {
		popped_key := heap.Pop(top_keys).(*Key)
		if key.Name != popped_key.Name {
			t.Errorf("Expected top key %v, got %v\n", key.Name, popped_key.Name)
		}
		if key.Hits != popped_key.Hits {
			t.Errorf("Expected key %s to have %d hits, got %vdn",
				key.Name, key.Hits, popped_key.Hits)
		}
	}
}

func TestHotKeysClone(t *testing.T) {
	h := NewHotKeyPool()
	h.Add([]string{"foo", "foo", "bar", "baz", "baz", "baz"})

	rotated := h.Rotate()

	// Validate old top keys
	top_keys := rotated.GetTopKeys()
	expected_keys := []string{"baz", "foo", "bar"}
	for _, key := range expected_keys {
		popped_key := heap.Pop(top_keys).(*Key)
		if key != popped_key.Name {
			t.Errorf("Expected top key %v, got %v\n", key, popped_key)
		}
	}
}
