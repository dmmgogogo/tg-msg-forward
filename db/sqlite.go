package db

import (
	"database/sql"
	"fmt"
	"log"
	"telegram-shell-bot/types"

	_ "github.com/mattn/go-sqlite3"
)

var DB *sql.DB

// InitDB 初始化数据库连接
func InitDB(dbPath string) error {
	var err error
	DB, err = sql.Open("sqlite3", dbPath)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}

	// 创建用户配置表
	err = createTables()
	if err != nil {
		return fmt.Errorf("failed to create tables: %w", err)
	}

	return nil
}

// 创建必要的表
func createTables() error {
	createTableSQL := `
	CREATE TABLE IF NOT EXISTS bot_configs (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL UNIQUE,
		token TEXT NOT NULL,
		target_chat_id INTEGER NOT NULL,
		start_cmd_message TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);`

	_, err := DB.Exec(createTableSQL)
	return err
}

// GetAllConfigs 获取所有机器人配置
func GetAllConfigs() ([]types.UserConfig, error) {
	rows, err := DB.Query(`
		SELECT name, token, target_chat_id, start_cmd_message 
		FROM bot_configs
		ORDER BY id ASC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var configs []types.UserConfig
	for rows.Next() {
		var config types.UserConfig
		err := rows.Scan(
			&config.Name,
			&config.Token,
			&config.TargetChatID,
			&config.StartCmdMessage,
		)
		if err != nil {
			return nil, err
		}
		configs = append(configs, config)
	}

	return configs, nil
}

// SaveConfig 保存或更新机器人配置
func SaveConfig(config *types.UserConfig) error {
	_, err := DB.Exec(`
		INSERT INTO bot_configs (name, token, target_chat_id, start_cmd_message)
		VALUES (?, ?, ?, ?)
		ON CONFLICT(name) DO UPDATE SET
			token = excluded.token,
			target_chat_id = excluded.target_chat_id,
			start_cmd_message = excluded.start_cmd_message,
			updated_at = CURRENT_TIMESTAMP
	`,
		config.Name,
		config.Token,
		config.TargetChatID,
		config.StartCmdMessage,
	)
	return err
}

// DeleteConfig 删除机器人配置
func DeleteConfig(name string) error {
	_, err := DB.Exec("DELETE FROM bot_configs WHERE name = ?", name)
	return err
}

// ImportFromYAML 从 YAML 配置导入到数据库
func ImportFromYAML(configs []types.UserConfig) error {
	tx, err := DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for _, config := range configs {
		err := SaveConfig(&config)
		if err != nil {
			log.Printf("Failed to import config for %s: %v", config.Name, err)
			return err
		}
	}

	return tx.Commit()
}
