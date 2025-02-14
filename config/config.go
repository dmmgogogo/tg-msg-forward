package config

import (
	"os"

	"gopkg.in/yaml.v2"
)

type Config struct {
	ProjectName   string   `yaml:"project_name"`
	Version       string   `yaml:"version"`
	DBName        string   `yaml:"db_name"`
	DBPath        string   `yaml:"db_path"`
	AdminBotToken string   `yaml:"admin_bot_token"`
	AdminName     []string `yaml:"admin_name"`
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
