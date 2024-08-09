package main

import (
	"encoding/json"
	"os"
)

func loadJSON[T any](filePath string) (T, error) {
	var data T
	fileData, err := os.ReadFile(filePath)
	if err != nil {
		return data, err
	}
	return data, json.Unmarshal(fileData, &data)
}
