package config

import (
	"os"
	"telegram-shell-bot/types"

	"gopkg.in/yaml.v2"
)

type Config struct {
	Version string             `yaml:"version"`
	Users   []types.UserConfig `yaml:"users"`
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
