package logger

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"golang.org/x/exp/slog"
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

// MultiHandler 将日志同时输出到多个 Handler
type MultiHandler struct {
	handlers []slog.Handler
}

func NewMultiHandler(handlers ...slog.Handler) *MultiHandler {
	return &MultiHandler{handlers: handlers}
}

func (h *MultiHandler) Enabled(ctx context.Context, level slog.Level) bool {
	for _, handler := range h.handlers {
		if handler.Enabled(ctx, level) {
			return true
		}
	}
	return false
}

func (h *MultiHandler) Handle(ctx context.Context, r slog.Record) error {
	for _, handler := range h.handlers {
		if handler.Enabled(ctx, r.Level) {
			if err := handler.Handle(ctx, r); err != nil {
				return err
			}
		}
	}
	return nil
}

func (h *MultiHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	newHandlers := make([]slog.Handler, len(h.handlers))
	for i, handler := range h.handlers {
		newHandlers[i] = handler.WithAttrs(attrs)
	}
	return &MultiHandler{handlers: newHandlers}
}

func (h *MultiHandler) WithGroup(name string) slog.Handler {
	newHandlers := make([]slog.Handler, len(h.handlers))
	for i, handler := range h.handlers {
		newHandlers[i] = handler.WithGroup(name)
	}
	return &MultiHandler{handlers: newHandlers}
}

// FileHandler 将日志写入文件（无 ANSI 颜色）
type FileHandler struct {
	level  slog.Level
	writer io.Writer
	attrs  []slog.Attr
}

func NewFileHandler(level slog.Level, writer io.Writer) *FileHandler {
	return &FileHandler{
		level:  level,
		writer: writer,
		attrs:  make([]slog.Attr, 0),
	}
}

func (h *FileHandler) Enabled(_ context.Context, level slog.Level) bool {
	return level >= h.level
}

func (h *FileHandler) Handle(_ context.Context, r slog.Record) error {
	timeStr := r.Time.Format("2006-01-02 15:04:05.000")
	levelStr := r.Level.String()

	file, line := h.getCallerInfo()

	var builder strings.Builder
	builder.WriteString(timeStr)
	builder.WriteString(" | ")
	builder.WriteString(levelStr)
	builder.WriteString(" | ")
	builder.WriteString(file + ":" + fmt.Sprintf("%d", line))
	builder.WriteString(" | ")
	builder.WriteString(r.Message)

	if r.NumAttrs() > 0 || len(h.attrs) > 0 {
		builder.WriteString(" |")
		for _, attr := range h.attrs {
			builder.WriteString(" ")
			builder.WriteString(attr.Key)
			builder.WriteString("=")
			builder.WriteString(fmt.Sprintf("%v", attr.Value))
		}
		r.Attrs(func(attr slog.Attr) bool {
			builder.WriteString(" ")
			builder.WriteString(attr.Key)
			builder.WriteString("=")
			builder.WriteString(fmt.Sprintf("%v", attr.Value))
			return true
		})
	}

	builder.WriteString("\n")
	_, err := fmt.Fprint(h.writer, builder.String())
	return err
}

func (h *FileHandler) getCallerInfo() (string, int) {
	for skip := 5; skip < 12; skip++ {
		_, file, line, ok := runtime.Caller(skip)
		if !ok {
			return "unknown", 0
		}
		if !strings.Contains(file, "slog") && !strings.Contains(file, "log") {
			return filepath.Base(file), line
		}
	}
	return "unknown", 0
}

func (h *FileHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	newAttrs := make([]slog.Attr, len(h.attrs)+len(attrs))
	copy(newAttrs, h.attrs)
	copy(newAttrs[len(h.attrs):], attrs)
	return &FileHandler{level: h.level, writer: h.writer, attrs: newAttrs}
}

func (h *FileHandler) WithGroup(name string) slog.Handler {
	return h
}

// SetupLogger 创建同时输出到控制台和文件的 logger
func SetupLogger(level slog.Level) *slog.Logger {
	consoleHandler := NewColoredHandler(level)

	f, err := os.Create("app.log")
	if err != nil {
		return slog.New(consoleHandler)
	}
	fileHandler := NewFileHandler(level, f)
	return slog.New(NewMultiHandler(consoleHandler, fileHandler))
}
