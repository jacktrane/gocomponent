package stringlib

import "strings"

// FirstUpper 首字母大写
func FirstUpper(s string) string {
	if s == "" {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}
