// 管理员机器人
package bot

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"telegram-shell-bot/db"
	"telegram-shell-bot/types"
	"telegram-shell-bot/utils"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const helpText = `🤖 *机器人管理系统*

*可用命令：*

1️⃣ */add*
   ✅ 添加新机器人
   👉 直接发送 /add 开始配置

2️⃣ */del*
   ❌ 删除现有机器人
   👉 格式：/del 机器人名称
   📝 例如：/del 小助手

3️⃣ */list*
   📋 显示所有机器人
   👉 查看当前配置列表

4️⃣ */help*
   ❓ 显示本帮助信息

发送任意命令开始操作...`

// 用于存储用户的添加机器人状态
type AddBotState struct {
	Step     int    // 当前步骤：1=输入名称，2=输入Token，3=输入群组ID
	Name     string // 机器人名称
	Token    string // 机器人Token
	ChatID   int64  // 目标群组ID
	Finished bool   // 是否完成
}

var (
	// 存储用户添加机器人的状态
	addBotStates = make(map[int64]*AddBotState)
)

func StartAdminBot(adminNames []string, tgToken string) error {
	api, err := tgbotapi.NewBotAPI(tgToken)
	if err != nil {
		return fmt.Errorf("failed to create bot: %w", err)
	}

	// 创建管理员用户名的 map，用于快速查找
	adminMap := make(map[string]bool)
	for _, name := range adminNames {
		adminMap[name] = true
	}

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := api.GetUpdatesChan(u)

	log.Printf("管理员Bot [%s] 已启动...", api.Self.UserName)

	for update := range updates {
		if update.Message == nil {
			continue
		}

		log.Printf("管理员[%s] 收到消息: MessageID: [%d], ChatType: [%s], Text: [%s] (from: %s, chat_id: %d)",
			api.Self.UserName,
			update.Message.MessageID,
			update.Message.Chat.Type,
			update.Message.Text,
			update.Message.From.UserName,
			update.Message.Chat.ID)

		log.Printf("管理员NewChatMembers: %v", update.Message.NewChatMembers)
		if len(update.Message.NewChatMembers) > 0 {
			sendMessage(api, update.Message.Chat.ID, fmt.Sprintf("当前群组ID：%d", update.Message.Chat.ID))
			continue
		}

		// 检查是否是管理员
		if !adminMap[update.Message.From.UserName] {
			sendMessage(api, update.Message.Chat.ID, "⛔️ 抱歉，只有管理员才能使用此机器人")
			continue
		}

		if update.Message.IsCommand() {
			handleCommand(api, update.Message)
		} else {
			handleText(api, update.Message)
		}
	}
	return nil
}

func handleCommand(api *tgbotapi.BotAPI, message *tgbotapi.Message) {
	cmd := message.Command()
	args := message.CommandArguments()

	switch cmd {
	case "start", "help":
		sendMessage(api, message.Chat.ID, helpText)

	case "add":
		// 开始交互式添加流程
		state := &AddBotState{Step: 1}
		addBotStates[message.Chat.ID] = state
		sendMessage(api, message.Chat.ID, "👉 请输入机器人名称：")

	case "del":
		handleDelCommand(api, message.Chat.ID, args)

	case "list":
		handleListCommand(api, message.Chat.ID)

	default:
		sendMessage(api, message.Chat.ID, "❌ 未知命令，请使用 /help 查看支持的命令")
	}
}

func handleText(api *tgbotapi.BotAPI, message *tgbotapi.Message) {
	state, exists := addBotStates[message.Chat.ID]
	if !exists {
		return
	}

	text := strings.TrimSpace(message.Text)

	switch state.Step {
	case 1: // 输入名称
		if len(text) == 0 {
			sendMessage(api, message.Chat.ID, "❌ 名称不能为空\n👉 请重新输入机器人名称：")
			return
		}
		state.Name = text
		state.Step = 2
		sendMessage(api, message.Chat.ID, "👉 请输入机器人Token：")

	case 2: // 输入Token
		// 验证 Token 格式
		if len(text) < 20 { // Telegram bot token 通常很长
			sendMessage(api, message.Chat.ID, "❌ Token 格式不正确，长度太短\n👉 请重新输入正确的 Token：")
			return
		}
		state.Token = text
		state.Step = 3
		sendMessage(api, message.Chat.ID, "👉 请输入目标群组ID（必须是数字，群组ID为负数）：")

	case 3: // 输入群组ID
		chatID, err := strconv.ParseInt(text, 10, 64)
		if err != nil {
			sendMessage(api, message.Chat.ID, "❌ 群组ID格式错误，必须是数字\n👉 请重新输入：")
			return
		}

		// 验证群组ID是否为负数
		if chatID >= 0 {
			sendMessage(api, message.Chat.ID, "❌ 群组ID必须是负数\n👉 请重新输入：")
			return
		}

		state.ChatID = chatID
		state.Finished = true

		// 创建配置
		config := &types.UserConfig{
			Name:            state.Name,
			Token:           state.Token,
			TargetChatID:    state.ChatID,
			StartCmdMessage: "(*￣︶￣)😊 您好，我在呢，请说...",
		}

		// 保存到数据库
		if err := db.SaveConfig(config); err != nil {
			log.Printf("保存配置失败：%v", err)
			sendMessage(api, message.Chat.ID, "❌ 保存配置失败")
			delete(addBotStates, message.Chat.ID)
			return
		}

		// 发送成功消息
		msg := fmt.Sprintf("✅ 成功添加机器人！\n\n"+
			"📝 配置信息：\n"+
			"名称：%s\n"+
			"Token：%s...\n"+
			"群组：%d\n\n",
			config.Name,
			getTokenPrefix(config.Token),
			config.TargetChatID)
		sendMessage(api, message.Chat.ID, msg)

		// 保存配置成功后，重启所有机器人
		if err := RestartBots(); err != nil {
			log.Printf("重启机器人失败：%v", err)
			sendMessage(api, message.Chat.ID, "⚠️ 机器人配置已保存，但重启失败")
			return
		}

		sendMessage(api, message.Chat.ID, "✨ 所有机器人已重新启动")

		// 清理状态
		delete(addBotStates, message.Chat.ID)
	}
}

// 安全地获取 Token 的前缀
func getTokenPrefix(token string) string {
	prefixLen := 15
	if len(token) < prefixLen {
		prefixLen = len(token)
	}
	return token[:prefixLen]
}

func handleDelCommand(api *tgbotapi.BotAPI, chatID int64, args string) {
	name := strings.TrimSpace(args)
	if name == "" {
		sendMessage(api, chatID, "❌ 请提供要删除的机器人名称\n格式：/del 机器人名称")
		return
	}

	// 先检查机器人是否存在
	configs, _ := db.GetAllConfigs()
	found := false
	for _, config := range configs {
		if config.Name == name {
			found = true
			break
		}
	}

	escapedName := utils.EscapeMarkdown(name)
	if !found {
		sendMessage(api, chatID, fmt.Sprintf("❌ 未找到名为 `%s` 的机器人", escapedName))
		return
	}

	if err := db.DeleteConfig(name); err != nil {
		log.Printf("删除机器人失败：%v", err)
		sendMessage(api, chatID, "❌ 删除失败")
		return
	}

	sendMessage(api, chatID, fmt.Sprintf("✅ 成功删除机器人：`%s`", escapedName))

	// 删除成功后，重启所有机器人
	if err := RestartBots(); err != nil {
		log.Printf("重启机器人失败：%v", err)
		sendMessage(api, chatID, "⚠️ 机器人已删除，但重启失败")
		return
	}

	sendMessage(api, chatID, "✨ 所有机器人已重新启动")
}

func handleListCommand(api *tgbotapi.BotAPI, chatID int64) {
	configs, err := db.GetAllConfigs()
	if err != nil {
		log.Printf("获取配置列表失败：%v", err)
		sendMessage(api, chatID, "❌ 获取配置列表失败")
		return
	}

	if len(configs) == 0 {
		sendMessage(api, chatID, "📝 当前没有配置任何机器人")
		return
	}

	var msg strings.Builder
	msg.WriteString("📋 *机器人配置列表*\n")
	msg.WriteString("━━━━━━━━━━━━━━\n\n")

	for i, config := range configs {
		// 机器人序号和名称
		msg.WriteString(fmt.Sprintf("*%d️⃣ %s*\n", i+1, config.Name))

		// 缩进显示详细信息
		msg.WriteString("   📱 群组：`")
		msg.WriteString(fmt.Sprintf("%d", config.TargetChatID))
		msg.WriteString("`\n")

		msg.WriteString("   🔑 Token：`")
		msg.WriteString(fmt.Sprintf("%s", getTokenPrefix(config.Token)))
		msg.WriteString("...`\n")

		// 添加分隔线，最后一个不加
		if i < len(configs)-1 {
			msg.WriteString("\n┄┄┄┄┄┄┄┄┄┄┄┄┄┄\n\n")
		}
	}

	// 添加底部提示
	msg.WriteString("\n━━━━━━━━━━━━━━\n")
	msg.WriteString("✨ 使用 /help 查看更多命令")

	sendMessage(api, chatID, msg.String())
}

func sendMessage(api *tgbotapi.BotAPI, chatID int64, text string) {
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = tgbotapi.ModeMarkdown
	if _, err := api.Send(msg); err != nil {
		log.Printf("Failed to send message: %v", err)
	}
}
