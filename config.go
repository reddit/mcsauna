package main

type RegexpConfig struct {
	Name string `json:"name,omitempty"`
	Re   string `json:"re"`
}

type Config struct {
	Regexps []RegexpConfig `json:"regexps"`
}
