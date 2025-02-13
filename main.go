package main

import (
	"log"
	"telegram-shell-bot/bot"
	"telegram-shell-bot/config"
)

func main() {
	// 初始化配置
	if err := config.Init(); err != nil {
		log.Fatalf("Failed to initialize config: %v", err)
	}

	// 打印每个机器人的 token 前缀（用于调试）
	for _, user := range config.AppConfig.Users {
		log.Printf("Bot %s token prefix: %s...", user.Name, user.Token[:10])
	}

	// 启动所有机器人
	if err := bot.StartAll(config.AppConfig.Users); err != nil {
		log.Fatalf("Failed to start bots: %v", err)
	}
}
