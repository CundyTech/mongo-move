package main

import (
	"os"
	"testing"
)

type testStruct struct {
	Name string `json:"name"`
	Age  int    `json:"age"`
}

func TestLoadJSON_Valid(t *testing.T) {
	file, err := os.CreateTemp("", "testjson_*.json")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(file.Name())
	file.WriteString(`{"name":"Alice","age":30}`)
	file.Close()

	result, err := loadJSON[testStruct](file.Name())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result.Name != "Alice" || result.Age != 30 {
		t.Errorf("unexpected result: %+v", result)
	}
}

func TestLoadJSON_FileNotFound(t *testing.T) {
	_, err := loadJSON[testStruct]("nonexistent.json")
	if err == nil {
		t.Error("expected error for missing file")
	}
}

func TestLoadJSON_InvalidJSON(t *testing.T) {
	file, err := os.CreateTemp("", "testjson_*.json")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(file.Name())
	file.WriteString(`not a json`)
	file.Close()

	_, err = loadJSON[testStruct](file.Name())
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}
