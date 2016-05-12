package mcsauna

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
	h.Add([]string{"foo", "foo", "bar", "baz", "baz", "baz"})
	top_keys := h.GetTopKeys()
	expected_keys := []string{"baz", "foo", "bar"}
	if !stringsEqual(top_keys, expected_keys) {
		t.Errorf("Expected top keys %v, got %v\n", expected_keys, top_keys)
	}
}
