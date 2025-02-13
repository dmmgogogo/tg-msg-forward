package types

// UserConfig 结构体用于映射单个用户的配置
type UserConfig struct {
	Name            string `yaml:"name"`
	Token           string `yaml:"token"`
	TargetChatID    int64  `yaml:"targetChatID"`
	StartCmdMessage string `yaml:"startCmdMessage"`
}
