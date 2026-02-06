package config

import "fmt"

// Config 统一配置结构体，消除全局可变状态
type Config struct {
	Host       string // 主站域名，如 "afdian.com"
	HostUrl    string // 完整 URL，如 "https://afdian.com"
	CookiePath string // cookie 文件路径
}

// NewConfig 创建配置，自动根据 host 生成 HostUrl
func NewConfig(host string, cookiePath string) *Config {
	return &Config{
		Host:       host,
		HostUrl:    fmt.Sprintf("https://%s", host),
		CookiePath: cookiePath,
	}
}
