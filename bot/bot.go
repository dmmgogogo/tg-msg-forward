package bot

import (
	"fmt"
	"log"
	"os"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"gopkg.in/yaml.v2"
)

type Bot struct {
	api          *tgbotapi.BotAPI
	targetChatID int64
}

// Config 结构体用于映射 config.yaml 文件
type Config struct {
	TargetChatID int64 `yaml:"targetChatID"` // 映射 targetChatID 字段
}

func New(token string) (*Bot, error) {
	// 读取配置文件
	config, err := readConfig("config.yaml")
	if err != nil {
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	log.Printf("Target Chat ID: [%d]", config.TargetChatID)

	api, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return nil, fmt.Errorf("failed to create bot: %w", err)
	}

	// 添加调试模式
	api.Debug = true

	// 打印机器人信息
	log.Printf("Bot Information:")
	log.Printf("- Username: %s", api.Self.UserName)
	log.Printf("- First Name: %s", api.Self.FirstName)
	log.Printf("- Can Join Groups: %v", api.Self.CanJoinGroups)
	log.Printf("- Can Read Group Messages: %v", api.Self.CanReadAllGroupMessages)

	return &Bot{api: api, targetChatID: config.TargetChatID}, nil
}

// 读取配置文件的函数
func readConfig(filePath string) (*Config, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var config Config
	decoder := yaml.NewDecoder(file)
	if err := decoder.Decode(&config); err != nil {
		return nil, err
	}
	return &config, nil
}

func (b *Bot) Start() error {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	// 添加所有类型的更新
	updates := b.api.GetUpdatesChan(u)

	log.Println("Bot is now running...") // 添加启动日志

	for update := range updates {
		// 打印所有收到的更新，用于调试
		log.Printf("Received update: %+v", update)

		if update.Message == nil {
			log.Println("Received nil message, skipping...")
			continue
		}

		// 打印消息详情
		log.Printf("Message: [%d] %s (from: %s, chat_id: %d)",
			update.Message.MessageID,
			update.Message.Text,
			update.Message.From.UserName,
			update.Message.Chat.ID)

		// 检查是否是直接命令
		if update.Message.IsCommand() {
			log.Printf("Received command: %s", update.Message.Command())
			continue
		}

		// 检查消息中是否包含机器人的用户名
		b.handleCommand(update.Message)
	}
	return nil
}

func (b *Bot) handleCommand(message *tgbotapi.Message) {
	log.Printf("Handling command from message: %s", message.Text)

	// 使用从配置文件中读取的 targetChatID
	targetChatID := b.targetChatID

	// 构建转发的消息内容，包含发送者信息
	senderInfo := fmt.Sprintf("来自用户: @%s", message.From.UserName)
	if message.From.UserName == "" {
		senderInfo = fmt.Sprintf("来自用户: %s %s", message.From.FirstName, message.From.LastName)
	}

	// 如果消息包含文本，直接发送
	if message.Text != "" {
		forwardText := fmt.Sprintf("%s\n\n消息内容：%s", senderInfo, message.Text)
		msg := tgbotapi.NewMessage(targetChatID, forwardText)
		_, err := b.api.Send(msg)
		if err != nil {
			log.Printf("Failed to send text message: %v", err)
		}
	}

	// 如果消息包含图片
	if message.Photo != nil && len(message.Photo) > 0 {
		photo := message.Photo[len(message.Photo)-1]
		photoMsg := tgbotapi.NewPhoto(targetChatID, tgbotapi.FileID(photo.FileID))
		photoMsg.Caption = senderInfo
		_, err := b.api.Send(photoMsg)
		if err != nil {
			log.Printf("Failed to forward photo: %v", err)
		}
	}

	// 如果消息包含文件
	if message.Document != nil {
		doc := tgbotapi.NewDocument(targetChatID, tgbotapi.FileID(message.Document.FileID))
		doc.Caption = senderInfo
		_, err := b.api.Send(doc)
		if err != nil {
			log.Printf("Failed to forward document: %v", err)
		}
	}

	// 如果消息包含语音消息
	if message.Voice != nil {
		voice := tgbotapi.NewVoice(targetChatID, tgbotapi.FileID(message.Voice.FileID))
		voice.Caption = senderInfo
		_, err := b.api.Send(voice)
		if err != nil {
			log.Printf("Failed to forward voice message: %v", err)
		}
	}
}

// func (b *Bot) sendMessage(chatID int64, text string) {
// 	msg := tgbotapi.NewMessage(chatID, text)
// 	_, err := b.api.Send(msg)
// 	if err != nil {
// 		log.Printf("Failed to send message: %v", err)
// 	}
// }
