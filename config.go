package main

import (
	"os"

	"gopkg.in/yaml.v3"
)

type DBConfig struct {
	URL    string
	Tables []string
}
type Config struct {
	Addr string
	DB   DBConfig
}

func NewConfig(configPath string) (*Config, error) {
	// Create config structure
	config := &Config{}

	// Open config file
	file, err := os.Open(configPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// Init new YAML decode
	d := yaml.NewDecoder(file)

	// Start YAML decoding from file
	if err := d.Decode(&config); err != nil {
		return nil, err
	}

	return config, nil
}
