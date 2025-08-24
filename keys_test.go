package main

import (
	"testing"
)

func TestDatabaseChoicesHelp(t *testing.T) {
	m := model{keyBindings: keyModel{keys: keys}}
	help := m.databaseChoicesHelp()
	if help == "" {
		t.Error("databaseChoicesHelp should return non-empty string")
	}
}

func TestCollectionChoicesHelp(t *testing.T) {
	m := model{keyBindings: keyModel{keys: keys}}
	help := m.collectionChoicesHelp()
	if help == "" {
		t.Error("collectionChoicesHelp should return non-empty string")
	}
}

func TestCollectionChoicesCopyHelp(t *testing.T) {
	m := model{keyBindings: keyModel{keys: keys}}
	help := m.collectionChoicesCopyHelp()
	if help == "" {
		t.Error("collectionChoicesCopyHelp should return non-empty string")
	}
}

func TestRestartHelp(t *testing.T) {
	m := model{keyBindings: keyModel{keys: keys}}
	help := m.RestartHelp()
	if help == "" {
		t.Error("RestartHelp should return non-empty string")
	}
}
