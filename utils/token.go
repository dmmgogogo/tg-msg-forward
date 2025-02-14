package utils

// GetTokenPrefix 安全地获取 Token 的前缀
func GetTokenPrefix(token string) string {
	prefixLen := 10
	if len(token) < prefixLen {
		prefixLen = len(token)
	}
	return token[:prefixLen]
}
