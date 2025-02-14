package main

import (
	"log"
	"os"
	"path/filepath"
	"telegram-shell-bot/bot"
	"telegram-shell-bot/config"
	"telegram-shell-bot/db"
	"time"
)

func main() {
	// 初始化配置
	if err := config.Init(); err != nil {
		log.Fatalf("Failed to initialize config: %v", err)
	}

	// 确保数据目录存在
	dataDir := filepath.Dir(config.AppConfig.DBPath)
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		log.Fatalf("Failed to create data directory: %v", err)
	}

	// 初始化数据库
	if err := db.InitDB(config.AppConfig.DBPath); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// 获取所有机器人配置
	configs, err := db.GetAllConfigs()
	if err != nil {
		log.Fatalf("Failed to get configs from database: %v", err)
	}

	// if len(configs) == 0 {
	// 	log.Fatalf("No bot configurations found in database. Please add configurations first.")
	// }

	// 启动管理员机器人
	go bot.StartAdminBot(config.AppConfig.AdminName, config.AppConfig.AdminBotToken)

	// 启动所有机器人
	if err := bot.StartAll(configs); err != nil {
		log.Fatalf("Failed to start bots: %v", err)
	}

	// 保持程序运行
	for {
		time.Sleep(time.Hour)
		log.Println("机器人运行中...")
	}
}
