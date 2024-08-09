package main

import (
	"fmt"
)

type config struct {
	Source string `json:"sourceServer"`
	Target string `json:"targetServer"`
}

func load() (config, error) {
	var cfg = &config{}
	config, err := loadJSON[config]("config.json")
	if err != nil {
		return *cfg, err
	}

	return config, err
}

func (c config) validate() error {
	if c.Source == "" {
		return fmt.Errorf("config value \"Source\" is missing")
	} else if c.Target == "" {
		return fmt.Errorf("config value \"Target\" is missing")
	} else {
		return nil
	}
}
