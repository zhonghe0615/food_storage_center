package config

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
)

// FulfillmentConfig holds the storage configuration for the application.
type FulfillmentConfig struct {
	// Storage configuration
	NumCoolers int `json:"num_coolers"`
	CoolerCap  int `json:"cooler_cap"`
	NumHeaters int `json:"num_heaters"`
	HeaterCap  int `json:"heater_cap"`
	NumShelves int `json:"num_shelves"`
	ShelfCap   int `json:"shelf_cap"`
}

// DefaultConfig returns the default configuration.
func DefaultConfig() FulfillmentConfig {
	return FulfillmentConfig{
		NumCoolers: 1,
		CoolerCap:  6,
		NumHeaters: 1,
		HeaterCap:  6,
		NumShelves: 1,
		ShelfCap:   12,
	}
}

// LoadConfig loads configuration from a JSON file.
// If the file doesn't exist, it creates one with default values.
func LoadConfig(configPath string) FulfillmentConfig {
	// Ensure directory exists
	dir := filepath.Dir(configPath)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if err := os.MkdirAll(dir, 0755); err != nil {
			log.Printf("Failed to create config directory: %v", err)
			return DefaultConfig()
		}
	}

	// Try to read existing config
	data, err := ioutil.ReadFile(configPath)
	if err != nil {
		// If file doesn't exist, create it with default config
		if os.IsNotExist(err) {
			config := DefaultConfig()
			saveConfig(configPath, config)
			return config
		}
		log.Printf("Error reading config file: %v", err)
		return DefaultConfig()
	}

	// Parse config
	var config FulfillmentConfig
	if err := json.Unmarshal(data, &config); err != nil {
		log.Printf("Error parsing config file: %v", err)
		return DefaultConfig()
	}

	return config
}

// saveConfig saves the configuration to a JSON file
func saveConfig(configPath string, config FulfillmentConfig) {
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		log.Printf("Error serializing config: %v", err)
		return
	}

	if err := ioutil.WriteFile(configPath, data, 0644); err != nil {
		log.Printf("Error writing config file: %v", err)
	}
}
