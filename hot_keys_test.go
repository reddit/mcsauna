package main

import (
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
	h.Add([]string{"foo", "foo", "bar", "baz", "baz", "baz", "car", "fish", "boat", "zoo", "zoo"})

	// Validate top keys
	top_keys := h.GetTopKeys()
	//                        3       2      2      1      1      1       1
	expected_keys := []string{"baz", "foo", "zoo", "bar", "car", "fish", "boat"}
	if !stringsEqual(top_keys, expected_keys) {
		t.Errorf("Expected top keys %v, got %v\n", expected_keys, top_keys)
	}

	// Validate hits
	if h.GetHits("foo") != 2 {
		t.Errorf("Expected key %s to have %d hits, got %vdn", "foo", 2, h.GetHits("foo"))
	}
	if h.GetHits("bar") != 1 {
		t.Errorf("Expected key %s to have %d hits, got %vdn", "bar", 1, h.GetHits("bar"))
	}
	if h.GetHits("baz") != 3 {
		t.Errorf("Expected key %s to have %d hits, got %vdn", "baz", 3, h.GetHits("baz"))
	}
	if h.GetHits("zoo") != 2 {
		t.Errorf("Expected key %s to have %d hits, got %vdn", "zoo", 3, h.GetHits("zoo"))
	}
}

func TestHotKeysClone(t *testing.T) {
	h := NewHotKeyPool()
	h.Add([]string{"foo", "foo", "bar", "baz", "baz", "baz"})

	rotated := h.Rotate()

	// Validate old top keys
	top_keys := rotated.GetTopKeys()
	expected_keys := []string{"baz", "foo", "bar"}
	if !stringsEqual(top_keys, expected_keys) {
		t.Errorf("Expected top keys %v, got %v\n", expected_keys, top_keys)
	}

	// Validate new top keys
	top_keys = h.GetTopKeys()
	expected_keys = []string{}
	if !stringsEqual(top_keys, expected_keys) {
		t.Errorf("Expected top keys %v, got %v\n", expected_keys, top_keys)
	}

}
