package internal

import (
	"encoding/json"
	"fmt"
	"os"
)

type Config struct {
	Token       string `json:"token"`
	Endpoint    string `json:"endpoint"`
	AutoUpdate  bool   `json:"auto_update"`
	CurrentVer  string `json:"current_version"`
}

func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("erro lendo config: %v", err)
	}
	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("erro parseando config: %v", err)
	}
	return &cfg, nil
}
