package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Webhooks      map[string]WebhookConfig `yaml:"webhooks"`
	ArchivistaUrl string                   `yaml:"archivistaUrl"`
}

type WebhookConfig struct {
	Type          string         `yaml:"type"`
	Signer        string         `yaml:"signer"`
	SignerOptions map[string]any `yaml:"signerOptions"`
	Options       map[string]any `yaml:"options"`
}

func New(filePath string) (Config, error) {
	configBytes, err := os.ReadFile(filePath)
	if err != nil {
		return Config{}, fmt.Errorf("could not read config file: %w", err)
	}

	config := Config{}
	if err := yaml.Unmarshal(configBytes, &config); err != nil {
		return config, fmt.Errorf("could not parse config yaml: %w", err)
	}

	return config, nil
}
