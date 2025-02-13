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

	// 打印token前几个字符，用于调试
	log.Printf("Bot token prefix: %s...", config.AppConfig.BotToken[:10])

	// 创建并启动机器人
	b, err := bot.New(config.AppConfig.BotToken)
	if err != nil {
		log.Fatalf("Failed to create bot: %v", err)
	}

	log.Println("Bot started...")
	if err := b.Start(); err != nil {
		log.Fatalf("Bot error: %v", err)
	}
}
