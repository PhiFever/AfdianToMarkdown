package logger

import (
	"context"
	"fmt"
	"golang.org/x/exp/slog"
	"path/filepath"
	"runtime"
	"strings"
)

// ANSI 颜色代码
const (
	Reset  = "\033[0m"
	Red    = "\033[31m"
	Green  = "\033[32m"
	Yellow = "\033[33m"
	Blue   = "\033[34m"
	Purple = "\033[35m"
	Cyan   = "\033[36m"
	Gray   = "\033[37m"
	White  = "\033[97m"
	Bold   = "\033[1m"
)

// ColoredHandler 是一个支持彩色输出的自定义 slog Handler
type ColoredHandler struct {
	level slog.Level
	attrs []slog.Attr
}

// NewColoredHandler 创建一个新的彩色处理器
func NewColoredHandler(level slog.Level) *ColoredHandler {
	return &ColoredHandler{
		level: level,
		attrs: make([]slog.Attr, 0),
	}
}

// Enabled 检查是否应该记录给定级别的日志
func (h *ColoredHandler) Enabled(_ context.Context, level slog.Level) bool {
	return level >= h.level
}

// Handle 处理日志记录
func (h *ColoredHandler) Handle(_ context.Context, r slog.Record) error {
	// 获取颜色和级别字符串
	color, levelStr := h.getLevelColorAndString(r.Level)

	// 格式化时间
	timeStr := r.Time.Format("2006-01-02 15:04:05.000")

	// 获取调用者信息（文件名和行号）
	file, line := h.getCallerInfo()

	// 构建基本日志行
	var builder strings.Builder

	// 时间（灰色）
	builder.WriteString(Gray + timeStr + Reset)
	builder.WriteString(" | ")

	// 级别（带颜色和粗体）
	builder.WriteString(color + Bold + levelStr + Reset)
	builder.WriteString(" | ")

	// 文件名和行号（青色）
	builder.WriteString(Cyan + file + ":" + fmt.Sprintf("%d", line) + Reset)
	builder.WriteString(" | ")

	// 消息（带颜色）
	builder.WriteString(color + r.Message + Reset)

	// 处理属性
	if r.NumAttrs() > 0 || len(h.attrs) > 0 {
		builder.WriteString(" | ")

		// 添加处理器级别的属性
		for _, attr := range h.attrs {
			builder.WriteString(h.formatAttr(attr))
			builder.WriteString(" ")
		}

		// 添加记录级别的属性
		r.Attrs(func(attr slog.Attr) bool {
			builder.WriteString(h.formatAttr(attr))
			builder.WriteString(" ")
			return true
		})
	}

	// 输出到标准输出
	fmt.Println(builder.String())
	return nil
}

// getCallerInfo 获取调用者的文件名和行号
func (h *ColoredHandler) getCallerInfo() (string, int) {
	// 跳过的层级：
	// 0: getCallerInfo
	// 1: Handle
	// 2: slog internal
	// 3: slog.Info/Debug/Error etc.
	// 4: 实际的调用者
	for skip := 4; skip < 10; skip++ {
		_, file, line, ok := runtime.Caller(skip)
		if !ok {
			return "unknown", 0
		}

		// 跳过 slog 包内部的调用
		if !strings.Contains(file, "slog") && !strings.Contains(file, "log") {
			return filepath.Base(file), line
		}
	}
	return "unknown", 0
}

// WithAttrs 返回一个带有给定属性的新处理器
func (h *ColoredHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	newAttrs := make([]slog.Attr, len(h.attrs)+len(attrs))
	copy(newAttrs, h.attrs)
	copy(newAttrs[len(h.attrs):], attrs)

	return &ColoredHandler{
		level: h.level,
		attrs: newAttrs,
	}
}

// WithGroup 返回一个带有给定组名的新处理器
func (h *ColoredHandler) WithGroup(name string) slog.Handler {
	// 简化实现，不支持组
	return h
}

// getLevelColorAndString 根据日志级别返回对应的颜色和字符串
func (h *ColoredHandler) getLevelColorAndString(level slog.Level) (string, string) {
	switch level {
	case slog.LevelDebug:
		return Cyan, "DEBUG"
	case slog.LevelInfo:
		return Green, "INFO "
	case slog.LevelWarn:
		return Yellow, "WARN "
	case slog.LevelError:
		return Red, "ERROR"
	default:
		return White, "TRACE"
	}
}

// formatAttr 格式化属性
func (h *ColoredHandler) formatAttr(attr slog.Attr) string {
	return Purple + attr.Key + Reset + "=" + Blue + fmt.Sprintf("%v", attr.Value) + Reset
}

// SetupLogger 创建配置好的 logger
func SetupLogger(level slog.Level) *slog.Logger {
	handler := NewColoredHandler(level)
	return slog.New(handler)
}
