package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
)

// Config is the class for reading and managing the goHomeAuto config.
type Config struct {
	Entities []Entity
	Board    uint8
}

// LoadConfig creates a new Config instance and automatically loads data from the given file
func LoadConfig(path string) Config {
	config := Config{}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		log.Fatalf("Given Config '%s' does not exist.\n", path)
	}

	fileData, err := ioutil.ReadFile(path)
	if err != nil {
		log.Fatalf("Could not read Config file '%s': '%s'", path, err)
	}

	if err := json.Unmarshal(fileData, &config); err != nil {
		log.Fatalf("Error reading JSON Data of Config: '%s'", err)
	}
	return config
}
