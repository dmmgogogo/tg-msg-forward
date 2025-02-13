# Telegram Shell Bot

这是一个 Telegram 机器人，可以在群组中响应命令并执行本地 shell 命令。

## 功能特性

1. 监听群组消息
2. 响应 @机器人 的提及
3. 执行本地 shell 命令并返回结果
4. 支持查询版本信息

## 使用方法

1. 在群组中 @机器人，格式如下：
   - @bot version - 获取版本信息
   - @bot shell <命令> - 执行 shell 命令

## 安全说明

- 机器人只响应特定用户的命令
- shell 命令执行有访问限制，仅允许执行安全命令
- 所有命令执行都会被记录

## 配置说明

在 config.yaml 中配置以下信息：
- BOT_TOKEN: Telegram 机器人的访问令牌
- ALLOWED_USERS: 允许执行命令的用户 ID 列表
- VERSION: 当前版本号 