# Telegram 消息转发机器人

这是一个 Telegram 消息转发机器人，可以将发送给机器人的消息转发到指定的聊天或群组。

## 功能特点

- 支持转发的消息类型：
  - 文本消息
  - 图片
  - 文件
  - 语音消息
  - 贴纸消息
  - GIF 动画
  - 视频
  - 文档
  - 投票
  - 位置信息
- 转发的消息会包含原始发送者的信息
- 支持命令过滤（忽略 /start 等命令消息）
- 支持配置文件动态设置目标聊天 ID

## 配置说明

在 config.yaml 中配置基本信息：

    project_name: "telegram-shell-bot"  # 项目名称
    version: "1.0.0"                    # 版本号
    db_name: "telegram_shell_bot"       # 数据库名称
    db_path: "data/bot.db"             # 数据库路径

使用命令行工具管理机器人配置：

# 添加机器人配置
go run cmd/botctl/main.go add -name "少奶奶" -token "your-token" -chat -4604394005

# 列出所有配置
go run cmd/botctl/main.go list

# 删除配置
go run cmd/botctl/main.go delete -name "少奶奶"

## 使用方法

1. 启动机器人：

    go run main.go

2. 发送消息给机器人：
   - 直接给机器人发送消息
   - 支持发送文本、图片、文件、语音消息
   - 机器人会自动将消息转发到配置的目标聊天

## 获取聊天 ID

1. 个人聊天 ID：
   - 给 [@userinfobot](https://t.me/userinfobot) 发送消息
   - 机器人会返回你的用户 ID

2. 群组 ID：
   - 将机器人添加到群组
   - 在群组中发送消息
   - 查看机器人日志中的 chat_id（带负号的数字）

## 注意事项

1. 确保机器人有权限发送消息到目标聊天
2. 群组 ID 通常是负数（如 -4604394005），个人聊天 ID 是正数
3. 机器人需要先被添加到群组才能获取群组 ID
4. 命令消息（如 /start）会被自动过滤，不会被转发

## 开发计划

- [ ] 添加用户白名单功能
- [ ] 支持更多消息类型（视频、贴纸等）
- [ ] 添加消息过滤功能
- [ ] 支持多目标转发 

## 发布脚本：
```bash
git pull
go mod tidy
# GOOS=linux GOARCH=amd64 go build -o app main.go
CC=x86_64-linux-gnu-gcc CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -o app main.go
zip app.zip app
```

## 上传脚本：
```bash
./upload.sh app.zip
```

## 运行脚本：
```bash
nohup ./app > output.log 2>&1 &
```   

## docker部署
```bash
docker run -it --rm -v /Users/mmx/Documents/work/TOOLS/tg_msg_forward:/workspace ubuntu:20.04 bash

# 安装依赖工具
apt-get install -y wget

# 下载 Go 最新版本（可以更改为你需要的版本）
cd /tmp
wget https://go.dev/dl/go1.24.0.linux-amd64.tar.gz

# 删除旧版 Go
rm -rf /usr/local/go

# 解压安装新的 Go 版本
tar -C /usr/local -xzf go1.24.0.linux-amd64.tar.gz

# 配置 Go 环境变量
export PATH=$PATH:/usr/local/go/bin

# 确保环境变量设置正确
echo "export PATH=\$PATH:/usr/local/go/bin" >> ~/.bashrc
source ~/.bashrc

# 检查 Go 版本
go version

# 安装其他依赖工具
apt-get install -y ca-certificates
apt-get install -y gcc-x86-64-linux-gnu
apt-get install -y build-essential

which x86_64-linux-gnu-gcc

cd /workspace
CC=x86_64-linux-gnu-gcc CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -o app main.go

```
