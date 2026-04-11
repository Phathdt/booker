package logger

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"sync"
	"time"
)

// ANSI color codes
const (
	colorReset      = "\033[0m"
	colorRed        = "\033[31m"
	colorGreen      = "\033[32m"
	colorYellow     = "\033[33m"
	colorPurple     = "\033[35m"
	colorCyan       = "\033[36m"
	colorGray       = "\033[90m"
	colorWhite      = "\033[97m"
	colorBoldRed    = "\033[1;31m"
	colorBoldGreen  = "\033[1;32m"
	colorBoldYellow = "\033[1;33m"
	colorBoldCyan   = "\033[1;36m"
)

// Logger is the logging interface used across the application
type Logger interface {
	Debug(msg string, args ...any)
	Info(msg string, args ...any)
	Warn(msg string, args ...any)
	Error(msg string, args ...any)
	Fatal(msg string, args ...any)
	With(args ...any) Logger
	WithGroup(name string) Logger
	DebugContext(ctx context.Context, msg string, args ...any)
	InfoContext(ctx context.Context, msg string, args ...any)
	WarnContext(ctx context.Context, msg string, args ...any)
	ErrorContext(ctx context.Context, msg string, args ...any)
}

// SlogLogger wraps slog.Logger
type SlogLogger struct {
	slog *slog.Logger
}

// New creates a new Logger
func New(format, level string) Logger {
	var handler slog.Handler
	opts := &slog.HandlerOptions{Level: parseLevel(level)}

	switch format {
	case "json":
		handler = slog.NewJSONHandler(os.Stdout, opts)
	case "plain":
		handler = slog.NewTextHandler(os.Stdout, opts)
	default:
		handler = NewColoredTextHandler(os.Stdout, opts)
	}

	return &SlogLogger{slog: slog.New(handler)}
}

func parseLevel(level string) slog.Level {
	switch level {
	case "debug":
		return slog.LevelDebug
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

func (l *SlogLogger) With(args ...any) Logger       { return &SlogLogger{slog: l.slog.With(args...)} }
func (l *SlogLogger) WithGroup(name string) Logger  { return &SlogLogger{slog: l.slog.WithGroup(name)} }
func (l *SlogLogger) Debug(msg string, args ...any) { l.slog.Debug(msg, args...) }
func (l *SlogLogger) Info(msg string, args ...any)  { l.slog.Info(msg, args...) }
func (l *SlogLogger) Warn(msg string, args ...any)  { l.slog.Warn(msg, args...) }
func (l *SlogLogger) Error(msg string, args ...any) { l.slog.Error(msg, args...) }
func (l *SlogLogger) Fatal(msg string, args ...any) { l.slog.Error(msg, args...); os.Exit(1) }

func (l *SlogLogger) DebugContext(ctx context.Context, msg string, args ...any) {
	l.slog.DebugContext(ctx, msg, args...)
}

func (l *SlogLogger) InfoContext(ctx context.Context, msg string, args ...any) {
	l.slog.InfoContext(ctx, msg, args...)
}

func (l *SlogLogger) WarnContext(ctx context.Context, msg string, args ...any) {
	l.slog.WarnContext(ctx, msg, args...)
}

func (l *SlogLogger) ErrorContext(ctx context.Context, msg string, args ...any) {
	l.slog.ErrorContext(ctx, msg, args...)
}

// ColoredTextHandler outputs colored log lines
type ColoredTextHandler struct {
	opts   slog.HandlerOptions
	mu     *sync.Mutex
	out    io.Writer
	attrs  []slog.Attr
	groups []string
}

func NewColoredTextHandler(out io.Writer, opts *slog.HandlerOptions) *ColoredTextHandler {
	if opts == nil {
		opts = &slog.HandlerOptions{}
	}
	return &ColoredTextHandler{opts: *opts, mu: &sync.Mutex{}, out: out}
}

func (h *ColoredTextHandler) Enabled(_ context.Context, level slog.Level) bool {
	minLevel := slog.LevelInfo
	if h.opts.Level != nil {
		minLevel = h.opts.Level.Level()
	}
	return level >= minLevel
}

func (h *ColoredTextHandler) Handle(_ context.Context, r slog.Record) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	timeStr := r.Time.Format(time.DateTime)
	levelColor, levelText := h.levelColorAndText(r.Level)

	var attrs string
	for _, a := range h.attrs {
		attrs += h.formatAttr(a)
	}
	r.Attrs(func(a slog.Attr) bool {
		attrs += h.formatAttr(a)
		return true
	})

	line := fmt.Sprintf("%s%s%s %s%-5s%s %s%s%s%s\n",
		colorGray, timeStr, colorReset,
		levelColor, levelText, colorReset,
		colorWhite, r.Message, colorReset,
		attrs,
	)
	_, err := h.out.Write([]byte(line))
	return err
}

func (h *ColoredTextHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	newAttrs := make([]slog.Attr, len(h.attrs)+len(attrs))
	copy(newAttrs, h.attrs)
	copy(newAttrs[len(h.attrs):], attrs)
	return &ColoredTextHandler{opts: h.opts, mu: h.mu, out: h.out, attrs: newAttrs, groups: h.groups}
}

func (h *ColoredTextHandler) WithGroup(name string) slog.Handler {
	newGroups := make([]string, len(h.groups)+1)
	copy(newGroups, h.groups)
	newGroups[len(h.groups)] = name
	return &ColoredTextHandler{opts: h.opts, mu: h.mu, out: h.out, attrs: h.attrs, groups: newGroups}
}

func (h *ColoredTextHandler) levelColorAndText(level slog.Level) (string, string) {
	switch {
	case level < slog.LevelInfo:
		return colorBoldCyan, "DEBUG"
	case level < slog.LevelWarn:
		return colorBoldGreen, "INFO"
	case level < slog.LevelError:
		return colorBoldYellow, "WARN"
	default:
		return colorBoldRed, "ERROR"
	}
}

func (h *ColoredTextHandler) formatAttr(a slog.Attr) string {
	if a.Equal(slog.Attr{}) {
		return ""
	}
	key := a.Key
	for _, g := range h.groups {
		key = g + "." + key
	}
	value := a.Value.Resolve()
	var valueColor, valueStr string
	switch value.Kind() {
	case slog.KindString:
		valueColor, valueStr = colorGreen, value.String()
	case slog.KindInt64:
		valueColor, valueStr = colorPurple, fmt.Sprintf("%d", value.Int64())
	case slog.KindFloat64:
		valueColor, valueStr = colorPurple, fmt.Sprintf("%g", value.Float64())
	case slog.KindBool:
		valueColor, valueStr = colorYellow, fmt.Sprintf("%t", value.Bool())
	case slog.KindDuration:
		valueColor, valueStr = colorPurple, value.Duration().String()
	default:
		valueColor, valueStr = colorWhite, fmt.Sprintf("%v", value.Any())
	}
	return fmt.Sprintf(" %s%s%s=%s%s%s", colorCyan, key, colorReset, valueColor, valueStr, colorReset)
}
