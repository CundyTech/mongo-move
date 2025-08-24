package main

import (
	"os"
	"testing"
)

func TestConfigValidate_MissingSource(t *testing.T) {
	cfg := config{Source: "", Target: "target"}
	err := cfg.validate()
	if err == nil || err.Error() != "config value \"Source\" is missing" {
		t.Errorf("expected missing Source error, got %v", err)
	}
}

func TestConfigValidate_MissingTarget(t *testing.T) {
	cfg := config{Source: "source", Target: ""}
	err := cfg.validate()
	if err == nil || err.Error() != "config value \"Target\" is missing" {
		t.Errorf("expected missing Target error, got %v", err)
	}
}

func TestConfigValidate_Valid(t *testing.T) {
	cfg := config{Source: "source", Target: "target"}
	err := cfg.validate()
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

func TestLoadConfig_FileNotFound(t *testing.T) {
	// Temporarily rename config.json if it exists
	_ = os.Remove("config.json")
	_, err := load()
	if err == nil {
		t.Error("expected error for missing config.json file")
	}
}
