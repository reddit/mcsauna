package main

import (
	"errors"
	"regexp"
)

// If the user does not specify a name for the regular expression, we will
// use the regular expression itself as the name.  Regular expressions tend
// to have funky characters, so we whitelist ones that we know will not cause
// trouble with graphite.
const WHITELISTED_STAT_CHARS = "[^a-zA-Z0-9-_]"

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
	if name == "" {
		re, err := regexp.Compile(WHITELISTED_STAT_CHARS)
		if err != nil {
			panic(err)
		}
		r.Name = string(re.ReplaceAll([]byte(r.OriginalRegexp), []byte{}))
	} else {
		r.Name = name
	}
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
func (r *RegexpKeys) Match(key string) (string, error) {
	for _, re := range r.regexp_keys {
		if re.CompiledRegexp.Match([]byte(key)) {
			if re.Name != "" {
				return re.Name, nil
			} else {
				return re.OriginalRegexp, nil
			}
		}
	}
	return "", errors.New("Could not match key to regex.")
}
