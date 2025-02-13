package config

import (
	"os"

	"gopkg.in/yaml.v2"
)

type Config struct {
	BotToken     string  `yaml:"bot_token"`
	AllowedUsers []int64 `yaml:"allowed_users"`
	Version      string  `yaml:"version"`
	ServerIP     string  `yaml:"server_ip"`
}

var AppConfig Config

func Init() error {
	data, err := os.ReadFile("config.yaml")
	if err != nil {
		return err
	}

	err = yaml.Unmarshal(data, &AppConfig)
	if err != nil {
		return err
	}

	return nil
}
