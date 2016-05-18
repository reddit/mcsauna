package main

import (
	"errors"
	"regexp"
)

type RegexpKey struct {
	OriginalRegexp string
	CompiledRegexp *regexp.Regexp
	Name           string
}

func NewRegexpKey(re string, name string) (regexp_key *RegexpKey, err error) {
	r := &RegexpKey{}
	compiled_regexp, err := regexp.Compile(re)
	if err != nil {
		return r, err
	}
	r.OriginalRegexp = re
	r.CompiledRegexp = compiled_regexp
	r.Name = name
	return r, nil
}

type RegexpKeys struct {
	regexp_keys []*RegexpKey
}

func NewRegexpKeys() *RegexpKeys {
	return &RegexpKeys{}
}

func (r *RegexpKeys) Add(regexp_key *RegexpKey) {
	r.regexp_keys = append(r.regexp_keys, regexp_key)
}

// Match finds the first regexp that a key matches and returns either its
// associated name, or the original regex string used in its compilation
func (r *RegexpKeys) Match(key string) (error, string) {
	for _, re := range r.regexp_keys {
		if re.CompiledRegexp.Match([]byte(key)) {
			if re.Name != "" {
				return nil, re.Name
			} else {
				return nil, re.OriginalRegexp
			}
		}
	}
	return errors.New("Could not match key to regex."), ""
}
