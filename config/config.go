package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	WebhookConfigs map[string]WebhookConfig
	ArchivistaUrl  string
}

type WebhookConfig struct {
	Type         string
	Signer       string
	SignerConfig map[string]any
	Options      map[string]any
}

func New(filePath string) (Config, error) {
	configBytes, err := os.ReadFile(filePath)
	if err != nil {
		return Config{}, nil
	}

	config := Config{}
	if err := yaml.Unmarshal(configBytes, &config); err != nil {
		return config, fmt.Errorf("could not parse config yaml: %w", err)
	}

	return config, nil
}
