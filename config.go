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
	Quiet            bool           `json:"quiet"`
	OutputFile       string         `json:"output_file"`
	ShowErrors       bool           `json:"show_errors"`
}

func NewConfig(config_data []byte) (config Config, err error) {
	config = Config{
		Regexps:          []RegexpConfig{},
		Interval:         5,
		Interface:        "any",
		Port:             11211,
		NumItemsToReport: 20,
		Quiet:            false,
		ShowErrors:       true,
	}
	err = json.Unmarshal(config_data, &config)
	if err != nil {
		return config, err
	}
	return config, nil
}
