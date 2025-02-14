package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"telegram-shell-bot/config"
	"telegram-shell-bot/db"
	"telegram-shell-bot/types"
)

func main() {
	// 初始化配置
	if err := config.Init(); err != nil {
		log.Fatalf("Failed to initialize config: %v", err)
	}

	// 子命令
	addCmd := flag.NewFlagSet("add", flag.ExitOnError)
	listCmd := flag.NewFlagSet("list", flag.ExitOnError)
	deleteCmd := flag.NewFlagSet("delete", flag.ExitOnError)

	// add 命令的参数
	addName := addCmd.String("name", "", "Bot name")
	addToken := addCmd.String("token", "", "Bot token")
	addChatID := addCmd.Int64("chat", 0, "Target chat ID")
	addMessage := addCmd.String("message", "(*￣︶￣)😊 您好，我在呢，请说...", "Start command response message")

	// delete 命令的参数
	deleteName := deleteCmd.String("name", "", "Bot name to delete")

	if len(os.Args) < 2 {
		fmt.Println("Expected 'add', 'list', or 'delete' subcommands")
		os.Exit(1)
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

	switch os.Args[1] {
	case "add":
		addCmd.Parse(os.Args[2:])
		if *addName == "" || *addToken == "" || *addChatID == 0 {
			addCmd.PrintDefaults()
			os.Exit(1)
		}
		config := &types.UserConfig{
			Name:            *addName,
			Token:           *addToken,
			TargetChatID:    *addChatID,
			StartCmdMessage: *addMessage,
		}
		if err := db.SaveConfig(config); err != nil {
			log.Fatalf("Failed to save config: %v", err)
		}
		fmt.Printf("Added bot configuration for %s\n", *addName)

	case "list":
		listCmd.Parse(os.Args[2:])
		configs, err := db.GetAllConfigs()
		if err != nil {
			log.Fatalf("Failed to get configs: %v", err)
		}
		fmt.Println("Bot configurations:")
		for _, config := range configs {
			fmt.Printf("- Name: %s\n", config.Name)
			fmt.Printf("  Target Chat: %d\n", config.TargetChatID)
			fmt.Printf("  Token: %s...\n", config.Token[:10])
			fmt.Printf("  Message: %s\n\n", config.StartCmdMessage)
		}

	case "delete":
		deleteCmd.Parse(os.Args[2:])
		if *deleteName == "" {
			deleteCmd.PrintDefaults()
			os.Exit(1)
		}
		if err := db.DeleteConfig(*deleteName); err != nil {
			log.Fatalf("Failed to delete config: %v", err)
		}
		fmt.Printf("Deleted bot configuration for %s\n", *deleteName)

	default:
		fmt.Println("Expected 'add', 'list', or 'delete' subcommands")
		os.Exit(1)
	}
}
