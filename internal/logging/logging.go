// Package logging 提供 minijvm 的统一日志门面。
//
// 基于标准库 log/slog（Go 1.21+），零第三方依赖。
// 输出为单行纯文本（写 stderr）：
//
//	15:04:05.123 INFO  starting minijvm version=0.0.1
//	15:04:05.124 DEBUG loading class class=java/lang/Object
//
// 级别由 -Xlog 选项控制（见 internal.Cmd.LogLevel）；
// 书中 -verbose:class 的类加载追踪走 Debug 级别。
package logging

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"strings"
	"sync"
)

// timeFormat 采用短时间格式（不带日期）：JVM 运行通常在分钟级，日期意义不大。
const timeFormat = "15:04:05.000"

var (
	// 默认 info 级别、写 stderr；Setup 之前即可安全调用 Debug/Info 等。
	logger   = slog.New(newPlainHandler(os.Stderr, slog.LevelInfo, nil))
	minLevel = slog.LevelInfo
)

// Setup 按级别字符串初始化全局日志（写 stderr）。
// level 取 "debug" | "info" | "warn" | "error"，大小写不敏感，空串回退 info；
// 其它值返回错误且不改动现有配置。
func Setup(level string) error {
	return SetupTo(level, os.Stderr)
}

// SetupTo 是 Setup 的可测试接缝：额外指定输出目的地，其余行为完全一致
// （与 internal.ParseCmdFrom 的接缝模式相同）。
func SetupTo(level string, w io.Writer) error {
	lv, err := parseLevel(level)
	if err != nil {
		return err
	}
	minLevel = lv
	logger = slog.New(newPlainHandler(w, lv, nil))
	return nil
}

// parseLevel 把级别字符串解析为 slog 级别。
func parseLevel(s string) (slog.Level, error) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "debug":
		return slog.LevelDebug, nil
	case "info", "":
		return slog.LevelInfo, nil
	case "warn", "warning":
		return slog.LevelWarn, nil
	case "error":
		return slog.LevelError, nil
	default:
		return 0, fmt.Errorf("logging: 未知日志级别 %q（可选 debug|info|warn|error）", s)
	}
}

// IsDebug 报告 debug 级别是否启用，供调用方规避昂贵的参数构造。
func IsDebug() bool {
	return slog.LevelDebug >= minLevel
}

// Debug / Info / Warn / Error 是全局 logger 的便捷方法；
// 键值对用法：logging.Info("loaded class", "class", name, "from", entry)。
func Debug(msg string, args ...any) { logger.Debug(msg, args...) }
func Info(msg string, args ...any)  { logger.Info(msg, args...) }
func Warn(msg string, args ...any)  { logger.Warn(msg, args...) }
func Error(msg string, args ...any) { logger.Error(msg, args...) }

// plainHandler 输出单行紧凑文本：<时间> <级别> <消息> <attrs...>，
// 相比 slog.TextHandler 去掉了 time=/level=/msg= 前缀与日期，更贴近 JVM 日志风格。
type plainHandler struct {
	w     io.Writer
	min   slog.Level
	attrs []slog.Attr
	mu    *sync.Mutex
}

func newPlainHandler(w io.Writer, min slog.Level, attrs []slog.Attr) *plainHandler {
	return &plainHandler{w: w, min: min, attrs: attrs, mu: &sync.Mutex{}}
}

func (h *plainHandler) Enabled(_ context.Context, l slog.Level) bool {
	return l >= h.min
}

func (h *plainHandler) Handle(_ context.Context, r slog.Record) error {
	var b strings.Builder
	b.WriteString(r.Time.Format(timeFormat))
	b.WriteByte(' ')
	b.WriteString(levelTag(r.Level))
	b.WriteByte(' ')
	b.WriteString(r.Message)
	for _, a := range h.attrs {
		appendAttr(&b, a)
	}
	r.Attrs(func(a slog.Attr) bool {
		appendAttr(&b, a)
		return true
	})
	b.WriteByte('\n')

	h.mu.Lock()
	defer h.mu.Unlock()
	_, err := io.WriteString(h.w, b.String())
	return err
}

// WithAttrs 返回携带固定属性的新 handler（不可变：不改原 handler）。
func (h *plainHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	merged := make([]slog.Attr, 0, len(h.attrs)+len(attrs))
	merged = append(merged, h.attrs...)
	merged = append(merged, attrs...)
	return newPlainHandler(h.w, h.min, merged)
}

// WithGroup 不支持分组，直接返回自身。
func (h *plainHandler) WithGroup(string) slog.Handler { return h }

// levelTag 返回左对齐 5 字符的级别标签（DEBUG/INFO/WARN/ERROR 均为 5 字符）。
func levelTag(l slog.Level) string {
	switch {
	case l >= slog.LevelError:
		return "ERROR"
	case l >= slog.LevelWarn:
		return "WARN "
	case l >= slog.LevelInfo:
		return "INFO "
	default:
		return "DEBUG"
	}
}

// appendAttr 追加一个 " key=value"。
func appendAttr(b *strings.Builder, a slog.Attr) {
	if a.Equal(slog.Attr{}) {
		return
	}
	b.WriteByte(' ')
	b.WriteString(a.Key)
	b.WriteByte('=')
	b.WriteString(fmt.Sprint(a.Value.Any()))
}
