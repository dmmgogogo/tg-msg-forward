package utils

import "strings"

// EscapeMarkdown 转义 Markdown 特殊字符
func EscapeMarkdown(text string) string {
	replacements := []struct {
		old string
		new string
	}{
		{"_", "\\_"},
		{"*", "\\*"},
		{"`", "\\`"},
		{"[", "\\["},
	}

	escaped := text
	for _, r := range replacements {
		escaped = strings.ReplaceAll(escaped, r.old, r.new)
	}
	return escaped
}
