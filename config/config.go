package config

import (
	"fmt"

	"golang.org/x/exp/slog"
)

// Config 统一配置结构体，消除全局可变状态
type Config struct {
	Host          string // 主站域名，如 "afdian.com"
	HostUrl       string // 完整 URL，如 "https://afdian.com"
	DataDir       string // 数据存储目录（存放作者文件夹）
	CookiePath    string // cookie 文件路径
	DownloadMedia bool   // 是否下载音频和视频，默认 false
	SkipFailed    bool   // 下载失败时是否跳过继续，默认 false（终止程序）
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

// HandleErr 统一处理可跳过的错误：开启 SkipFailed 时记录日志并返回 nil，否则原样返回错误
func (c *Config) HandleErr(err error, msg string, kv ...any) error {
	if err == nil {
		return nil
	}
	if c.SkipFailed {
		slog.Error(msg, append(kv, "err", err)...)
		return nil
	}
	return err
}
