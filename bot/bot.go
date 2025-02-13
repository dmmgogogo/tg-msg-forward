package bot

import (
	"fmt"
	"log"
	"path/filepath"
	"strings"
	"sync"
	"telegram-shell-bot/types"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Bot struct {
	api        *tgbotapi.BotAPI
	userConfig *types.UserConfig
}

func New(config *types.UserConfig) (*Bot, error) {
	api, err := tgbotapi.NewBotAPI(config.Token)
	if err != nil {
		return nil, fmt.Errorf("failed to create bot: %w", err)
	}

	// 添加调试模式
	api.Debug = true

	// 打印机器人信息
	log.Printf("Bot Information for %s:", config.Name)
	log.Printf("- Username: %s", api.Self.UserName)
	log.Printf("- First Name: %s", api.Self.FirstName)
	log.Printf("- Can Join Groups: %v", api.Self.CanJoinGroups)
	log.Printf("- Can Read Group Messages: %v", api.Self.CanReadAllGroupMessages)
	log.Printf("- Target Chat ID: [%d]", config.TargetChatID)

	return &Bot{
		api:        api,
		userConfig: config,
	}, nil
}

// StartAll 启动所有配置的机器人
func StartAll(configs []types.UserConfig) error {
	var wg sync.WaitGroup
	errors := make(chan error, len(configs))

	for _, config := range configs {
		wg.Add(1)
		go func(cfg types.UserConfig) {
			defer wg.Done()

			bot, err := New(&cfg)
			if err != nil {
				errors <- fmt.Errorf("failed to create bot %s: %w", cfg.Name, err)
				return
			}

			if err := bot.Start(); err != nil {
				errors <- fmt.Errorf("bot %s error: %w", cfg.Name, err)
			}
		}(config)
	}

	// 等待所有 goroutine 完成
	go func() {
		wg.Wait()
		close(errors)
	}()

	// 收集错误
	for err := range errors {
		if err != nil {
			return err
		}
	}

	return nil
}

func (b *Bot) Start() error {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := b.api.GetUpdatesChan(u)

	log.Printf("Bot %s is now running...", b.userConfig.Name)

	for update := range updates {
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
	return nil
}

func (b *Bot) handleCommand(message *tgbotapi.Message) {
	log.Printf("Handling command from message: %s", message.Text)

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
	targetChatID := b.userConfig.TargetChatID

	// 构建转发的消息内容，包含发送者信息
	senderInfo := fmt.Sprintf("来自用户: @%s", message.From.UserName)
	if message.From.UserName == "" {
		senderInfo = fmt.Sprintf("来自用户: %s %s", message.From.FirstName, message.From.LastName)
	}

	// 处理文本消息
	if message.Text != "" {
		forwardText := fmt.Sprintf("%s\n\n消息内容：%s", senderInfo, message.Text)
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
