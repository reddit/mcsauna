package main

import (
	"encoding/json"
)

type RegexpConfig struct {
	Name string `json:"name,omitempty"`
	Re   string `json:"re"`
}

type Config struct {
	Regexps          []RegexpConfig `json:"regexps"`
	Interval         int            `json:"interval"`
	Interface        string         `json:"interface"`
	Port             int            `json:"port"`
	NumItemsToReport int            `json:"num_items_to_report"`
}

func NewConfig(config_data []byte) (config Config, err error) {
	config = Config{
		Regexps:          []RegexpConfig{},
		Interval:         60,
		Interface:        "any",
		Port:             11211,
		NumItemsToReport: 20,
	}
	err = json.Unmarshal(config_data, &config)
	if err != nil {
		return config, err
	}
	return config, nil
}
