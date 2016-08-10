package main

import (
	"testing"
)

type RegexpTestItem struct {
	Regexp          string
	Name            string
	ExpectedMatches []string
}

var REGEXP_TEST_CASES = []RegexpTestItem{
	RegexpTestItem{"^[a-f|0-9]{32}$", "MD5", []string{"aaaabbbbccccdddd1111222233334444"}},
	RegexpTestItem{"^Comment_[0-9]+$", "Comment", []string{"Comment_19837"}},
	RegexpTestItem{"^registered_vote_[0-9]+_t[0-9]{1}_[0-9|a-z]+$", "Registered Vote", []string{"registered_vote_9_t5_40z"}},
}

func TestRegexp(t *testing.T) {
	for _, test_item := range REGEXP_TEST_CASES {
		regexp_keys := NewRegexpKeys()
		regexp_key, _ := NewRegexpKey(test_item.Regexp, test_item.Name)
		regexp_keys.Add(regexp_key)

		// Expected matches
		for _, key := range test_item.ExpectedMatches {
			match, _ := regexp_keys.Match(key)
			if match != test_item.Name {
				t.Errorf("Expected match %s, got %s\n", test_item.Name, match)
			}
		}
	}
}
