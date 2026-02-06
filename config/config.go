package config

import "fmt"

// Config 统一配置结构体，消除全局可变状态
type Config struct {
	Host       string // 主站域名，如 "afdian.com"
	HostUrl    string // 完整 URL，如 "https://afdian.com"
	DataDir    string // 数据存储目录（存放作者文件夹）
	CookiePath string // cookie 文件路径
}

// NewConfig 创建配置，自动根据 host 生成 HostUrl
func NewConfig(host string, dataDir string, cookiePath string) *Config {
	return &Config{
		Host:       host,
		HostUrl:    fmt.Sprintf("https://%s", host),
		DataDir:    dataDir,
		CookiePath: cookiePath,
	}
}
