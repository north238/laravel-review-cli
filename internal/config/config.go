package config

import (
	"os"
	"strconv"
)

type Config struct {
	APIKey  string
	Model   string
	Timeout int
}

func NewConfig() *Config {
	apikey := os.Getenv("ANTHROPIC_API_KEY")

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

	return config
}
