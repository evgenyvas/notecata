// Package config
package config

import (
	"encoding/json"
	"fmt"
	"os"
)

type Config struct {
	Root    string
	Port    uint16
	Storage string
}

var instance *Config

func init() {
	// read configuration
	file, _ := os.Open("config.json")
	defer file.Close()
	decoder := json.NewDecoder(file)
	var config = Config{}
	err := decoder.Decode(&config)
	if err != nil {
		fmt.Println("error:", err)
	}
	instance = &config
}

func GetConfig() *Config {
	return instance
}
