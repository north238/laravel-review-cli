package config

import (
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	APIKey  string
	Model   string
	Timeout int
}

func NewConfig() (*Config, error) {
	apikey := os.Getenv("ANTHROPIC_API_KEY")
	if apikey == "" {
		return nil, fmt.Errorf("invalied to APIKey")
	}

	model := os.Getenv("LRV_MODEL")
	if model == "" {
		model = "claude-sonnet-4-5"
	}

	timeout, err := strconv.Atoi(os.Getenv("LRV_TIMEOUT"))
	if err != nil {
		timeout = 120
	}

	config := &Config{
		APIKey:  apikey,
		Model:   model,
		Timeout: timeout,
	}

	return config, nil
}
