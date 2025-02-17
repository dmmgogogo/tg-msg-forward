package bot

import (
	"fmt"
	"log"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"telegram-shell-bot/db"
	"telegram-shell-bot/types"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Bot struct {
	api        *tgbotapi.BotAPI
	userConfig *types.UserConfig
	running    bool
	stopChan   chan struct{}
	mu         sync.Mutex
	updates    tgbotapi.UpdatesChannel
}

var (
	// 存储所有运行中的机器人
	runningBots = make(map[string]*Bot)
	// 用于保护 runningBots 的互斥锁
	botsLock sync.RWMutex
)

func New(config *types.UserConfig) (*Bot, error) {
	api, err := tgbotapi.NewBotAPI(config.Token)
	if err != nil {
		return nil, fmt.Errorf("failed to create bot: %w", err)
	}

	// 添加调试模式
	api.Debug = true

	// 打印机器人信息
	log.Printf("消息转发Bot [%s] 配置中...", config.Name)
	log.Printf("- Username: [%s]", api.Self.UserName)
	log.Printf("- First Name: [%s]", api.Self.FirstName)
	log.Printf("- Can Join Groups: [%v]", api.Self.CanJoinGroups)
	log.Printf("- Can Read Group Messages: [%v]", api.Self.CanReadAllGroupMessages)
	log.Printf("- Target Chat ID: [%d]", config.TargetChatID)

	return &Bot{
		api:        api,
		userConfig: config,
		stopChan:   make(chan struct{}),
	}, nil
}

// StartAll 启动所有配置的机器人
func StartAll(configs []types.UserConfig) error {
	botsLock.Lock()
	defer botsLock.Unlock()

	for _, config := range configs {
		bot, err := New(&config)
		if err != nil {
			log.Printf("Failed to create bot %s: %v", config.Name, err)
			continue
		}
		runningBots[config.Name] = bot
		go bot.Start()
	}

	return nil
}

func (b *Bot) Start() error {
	b.mu.Lock()
	if b.running {
		b.mu.Unlock()
		return nil
	}
	b.running = true
	b.stopChan = make(chan struct{})
	b.mu.Unlock()

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	b.updates = b.api.GetUpdatesChan(u)

	log.Printf("消息转发Bot [%s] 已启动...", b.userConfig.Name)

	for {
		select {
		case <-b.stopChan:
			log.Printf("消息转发Bot [%s] 已停止...", b.userConfig.Name)
			return nil
		case update, ok := <-b.updates:
			if !ok {
				return nil
			}
			if update.Message == nil {
				continue
			}

			log.Printf("[%s] 收到消息: MessageID: [%d] %s (from: %s, chat_id: %d)",
				b.userConfig.Name,
				update.Message.MessageID,
				update.Message.Text,
				update.Message.From.UserName,
				update.Message.Chat.ID)

			if update.Message.IsCommand() {
				log.Printf("[%s] 命令消息: %s", b.userConfig.Name, update.Message.Command())

				if b.userConfig.StartCmdMessage != "" {
					b.sendMessage(update.Message.Chat.ID, b.userConfig.StartCmdMessage)
				}
				continue
			}

			b.handleCommand(update.Message)
		}
	}
}

func (b *Bot) handleCommand(message *tgbotapi.Message) {
	log.Printf("🤖机器人[%s]接收到消息: %s", b.userConfig.Name, message.Text)

	// 检查是否有任何内容需要处理
	hasContent := message.Text != "" ||
		message.Sticker != nil ||
		message.Animation != nil ||
		message.Video != nil ||
		message.Location != nil ||
		message.Poll != nil ||
		message.Document != nil ||
		message.Photo != nil ||
		message.Voice != nil

	if !hasContent {
		return
	}

	// 使用用户配置中的 targetChatID
	var targetChatID int64
	var senderInfo string
	var forwardText string

	// 获取当前message的chatId
	if message.Chat.ID != b.userConfig.TargetChatID {
		// 说明来自转发
		targetChatID = b.userConfig.TargetChatID

		// 构建转发的消息内容，包含发送者信息
		senderInfo = "⚠️回复这条消息跟客户沟通⚠️"
		senderInfo += "\n——————————————"
		senderInfo += fmt.Sprintf("\n来自用户: @%s", message.From.UserName)
		if message.From.UserName == "" {
			senderInfo = fmt.Sprintf("来自用户: %s %s", message.From.FirstName, message.From.LastName)
		}

		// 加上群组ID，将来回复消息的时候需要用到
		senderInfo += fmt.Sprintf("\n群组ID: %d", message.Chat.ID)
		forwardText = fmt.Sprintf("%s\n\n消息内容：", senderInfo)
	} else {
		// 说明来自转发群里面的管理人员回复
		if message.ReplyToMessage != nil {
			// 从回复的消息中提取群组ID
			lines := strings.Split(message.ReplyToMessage.Text, "\n")
			for _, line := range lines {
				if strings.HasPrefix(line, "群组ID: ") {
					idStr := strings.TrimPrefix(line, "群组ID: ")
					if id, err := strconv.ParseInt(idStr, 10, 64); err == nil {
						targetChatID = id
						log.Printf("🤖机器人[%s]提取到目标群组ID: %d", b.userConfig.Name, targetChatID)
						break
					}
				}
			}

			if targetChatID == b.userConfig.TargetChatID {
				log.Printf("推断：测试信息，不需要转发", b.userConfig.Name)
				return
			}
		}
	}

	if targetChatID == 0 {
		log.Printf("🤖机器人[%s]没有找到目标群组ID", b.userConfig.Name)
		return
	}

	// 处理文本消息
	if message.Text != "" {
		forwardText += message.Text
		msg := tgbotapi.NewMessage(targetChatID, forwardText)
		b.sendWithLog(msg, "text message")
	}

	// 处理 GIF 动画
	// if message.Animation != nil {
	// 	animation := tgbotapi.NewAnimation(targetChatID, tgbotapi.FileID(message.Animation.FileID))
	// 	animation.Caption = senderInfo
	// 	b.sendWithLog(animation, "animation")
	// }

	// 处理贴纸
	if message.Sticker != nil {
		stickerMsg := tgbotapi.NewSticker(targetChatID, tgbotapi.FileID(message.Sticker.FileID))
		b.sendWithLog(stickerMsg, "sticker")
		// 贴纸不支持 Caption，单独发送发送者信息
		infoMsg := tgbotapi.NewMessage(targetChatID, senderInfo)
		b.sendWithLog(infoMsg, "sticker sender info")
	}

	// 处理文档（包括 GIF）
	if message.Document != nil {
		doc := tgbotapi.NewDocument(targetChatID, tgbotapi.FileID(message.Document.FileID))
		caption := senderInfo
		// if !isGif(message.Document.FileName) {
		// caption = fmt.Sprintf("%s\n文件名: %s", senderInfo, message.Document.FileName)
		// }
		doc.Caption = caption
		b.sendWithLog(doc, "document")
	}

	// 处理图片
	if message.Photo != nil && len(message.Photo) > 0 {
		photo := message.Photo[len(message.Photo)-1]
		photoMsg := tgbotapi.NewPhoto(targetChatID, tgbotapi.FileID(photo.FileID))
		photoMsg.Caption = senderInfo
		b.sendWithLog(photoMsg, "photo")
	}

	// 处理语音消息
	if message.Voice != nil {
		voice := tgbotapi.NewVoice(targetChatID, tgbotapi.FileID(message.Voice.FileID))
		voice.Caption = senderInfo
		b.sendWithLog(voice, "voice message")
	}

	// 处理视频
	if message.Video != nil {
		videoMsg := tgbotapi.NewVideo(targetChatID, tgbotapi.FileID(message.Video.FileID))
		videoMsg.Caption = senderInfo
		b.sendWithLog(videoMsg, "video")
	}

	// 处理位置信息
	if message.Location != nil {
		loc := tgbotapi.NewLocation(targetChatID, message.Location.Latitude, message.Location.Longitude)
		b.sendWithLog(loc, "location")
		// 位置信息不支持 Caption，单独发送发送者信息
		infoMsg := tgbotapi.NewMessage(targetChatID, senderInfo)
		b.sendWithLog(infoMsg, "location sender info")
	}

	// 处理投票
	if message.Poll != nil {
		// 将 PollOption 转换为字符串切片
		options := make([]string, len(message.Poll.Options))
		for i, opt := range message.Poll.Options {
			options[i] = opt.Text
		}

		poll := tgbotapi.NewPoll(targetChatID, message.Poll.Question, options...)
		poll.IsAnonymous = message.Poll.IsAnonymous
		poll.Type = message.Poll.Type
		poll.AllowsMultipleAnswers = message.Poll.AllowsMultipleAnswers

		b.sendWithLog(poll, "poll")
		// 投票不支持 Caption，单独发送发送者信息
		infoMsg := tgbotapi.NewMessage(targetChatID, senderInfo)
		b.sendWithLog(infoMsg, "poll sender info")
	}
}

// sendWithLog 统一处理消息发送和错误日志
func (b *Bot) sendWithLog(msg tgbotapi.Chattable, msgType string) {
	_, err := b.api.Send(msg)
	if err != nil {
		log.Printf("Failed to forward %s: %v", msgType, err)
	}
	log.Printf("消息【%s】发送成功", msgType)
}

// 检查文件是否是 GIF
func isGif(fileName string) bool {
	if fileName == "" {
		return false
	}
	return strings.ToLower(filepath.Ext(fileName)) == ".gif"
}

func (b *Bot) sendMessage(chatID int64, text string) {
	msg := tgbotapi.NewMessage(chatID, text)
	_, err := b.api.Send(msg)
	if err != nil {
		log.Printf("Failed to send message: %v", err)
	}
}

// RestartBots 重新启动所有机器人
func RestartBots() error {
	botsLock.Lock()
	defer botsLock.Unlock()

	log.Println("开始重启所有机器人...")

	// 停止所有运行中的机器人
	for name, bot := range runningBots {
		log.Printf("正在停止机器人: %s", name)
		if bot != nil {
			bot.Stop()
		}
	}

	// 清空运行中的机器人列表
	runningBots = make(map[string]*Bot)

	// 获取最新配置
	configs, err := db.GetAllConfigs()
	if err != nil {
		return fmt.Errorf("获取配置失败: %w", err)
	}

	log.Printf("获取到 %d 个机器人配置", len(configs))

	// 启动所有机器人
	for _, config := range configs {
		log.Printf("正在启动机器人: %s", config.Name)
		bot, err := New(&config)
		if err != nil {
			log.Printf("创建机器人失败 %s: %v", config.Name, err)
			continue
		}
		runningBots[config.Name] = bot
		go bot.Start()
	}

	log.Println("所有机器人重启完成")
	return nil
}

// Stop 停止机器人
func (b *Bot) Stop() {
	b.mu.Lock()
	defer b.mu.Unlock()

	if !b.running {
		return
	}

	log.Printf("正在停止机器人 [%s]...", b.userConfig.Name)

	// 先标记为非运行状态
	b.running = false

	// 关闭停止通道
	close(b.stopChan)

	// 停止接收更新
	b.api.StopReceivingUpdates()

	// 不再等待清空通道
	b.updates = nil

	log.Printf("消息转发Bot [%s] 已停止", b.userConfig.Name)
}
