package main

import (
	"encoding/json"
	"log"
	"os"
)

type Config struct {
	BotToken   string `json:"bot_token"`
	APIBaseURL string `json:"api_base_url"`
}

func readConfig(file string) (*Config, error) {
	data, err := os.ReadFile(file)
	if err != nil {
		return nil, err
	}
	var config Config
	err = json.Unmarshal(data, &config)
	if err != nil {
		return nil, err
	}
	log.Printf("Config: %+v\n", config)
	return &config, nil
}
