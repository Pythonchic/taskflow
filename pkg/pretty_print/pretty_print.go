package prettyprint

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/fatih/color"
	"dario.cat/mergo"
)

// ---- Цветовые принтеры (создаются один раз) ----
var (
	debugPrinter    = color.New(color.FgMagenta)
	infoPrinter     = color.New(color.FgGreen)
	warnPrinter     = color.RGB(255, 121, 0)
	errorPrinter    = color.RGB(255, 0, 0)
	fatalPrinter    = color.RGB(255, 0, 75)
	successPrinter  = color.New(color.FgHiGreen)
	progressPrinter = color.New(color.FgCyan)
)

var printerMap = map[string]*color.Color{
	"DEBUG":    debugPrinter,
	"INFO":     infoPrinter,
	"WARN":     warnPrinter,
	"ERROR":    errorPrinter,
	"FATAL":    fatalPrinter,
	"SUCCESS":  successPrinter,
	"PROGRESS": progressPrinter,
}

// ---- КОНФИГУРАЦИЯ ПО УМОЛЧАНИЮ (как net/http.DefaultClient) ----
var DefaultConfig = Config{
	DebugMode:      false,
	ShowTime:       true,
	TimeFormat:     "2006/01/02 15:04:05",
	ShowLevel:      true,
	Colors:         true,
	ProgressSuffix: "...",
}

// Текущая конфигурация (изначально = DefaultConfig)
var currentConfig = DefaultConfig

// Config конфигурация логгера
type Config struct {
	DebugMode      bool
	ShowTime       bool
	TimeFormat     string
	ShowLevel      bool
	Colors         bool
	ProgressSuffix string
}

// ---- ФУНКЦИИ ДЛЯ РАБОТЫ С КОНФИГОМ ----

// Init полная инициализация с указанным конфигом
// Незаданные поля берутся из DefaultConfig
func Init(cfg Config) {
	// Мержим с дефолтным конфигом
	currentConfig = mergeWithDefault(cfg)

	Info("Logger initialized")
	if currentConfig.DebugMode {
		Debug("Debug mode enabled")
	}
}

// SetConfig частичное обновление конфигурации
func SetConfig(cfg Config) {
	oldDebug := currentConfig.DebugMode
	currentConfig = mergeWithDefault(cfg)

	// Если включили debug - пишем об этом
	if !oldDebug && currentConfig.DebugMode {
		Debug("Debug mode enabled")
	}
}

// ResetConfig сброс к конфигурации по умолчанию
func ResetConfig() {
	currentConfig = DefaultConfig
	Info("Logger reset to default config")
}

func mergeWithDefault(cfg Config) Config {
    result := DefaultConfig
    mergo.Merge(&result, cfg, mergo.WithOverride)
    return result
}

// ---- ОСНОВНАЯ ФУНКЦИЯ ЛОГИРОВАНИЯ ----
func write(level, format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)

	var prefix string
	cfg := currentConfig // используем текущую конфигурацию

	if cfg.ShowTime {
		timestamp := time.Now().Format(cfg.TimeFormat)
		prefix = fmt.Sprintf("[%s]", timestamp)
	}

	if cfg.ShowLevel {
		if prefix != "" {
			prefix += fmt.Sprintf(" [%s]", level)
		} else {
			prefix = fmt.Sprintf("[%s]", level)
		}
	}

	if prefix != "" {
		prefix += " "
	}

	fullMsg := prefix + msg

	if cfg.Colors {
		if printer, ok := printerMap[level]; ok {
			printer.Println(fullMsg)
		} else {
			fmt.Println(fullMsg)
		}
	} else {
		fmt.Println(fullMsg)
	}

	// Выход при фатальной ошибке
	if level == "FATAL" {
		os.Exit(1)
	}
}

// ---- ПУБЛИЧНЫЕ ФУНКЦИИ ЛОГИРОВАНИЯ ----

func Debug(format string, args ...interface{}) {
	if currentConfig.DebugMode {
		write("DEBUG", format, args...)
	}
}

func Info(format string, args ...interface{}) {
	write("INFO", format, args...)
}

func Warn(format string, args ...interface{}) {
	write("WARN", format, args...)
}

func Error(format string, args ...interface{}) {
	write("ERROR", format, args...)
}

func Fatal(format string, args ...interface{}) {
	write("FATAL", format, args...)
	// write уже вызывает os.Exit(1) для FATAL
}

func Success(format string, args ...interface{}) {
	write("SUCCESS", format, args...)
}

func Progress(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	if currentConfig.ProgressSuffix != "" && !strings.HasSuffix(msg, currentConfig.ProgressSuffix) {
		msg += currentConfig.ProgressSuffix
	}
	write("PROGRESS", "%s", msg)
}

// ---- ВСПОМОГАТЕЛЬНЫЕ ФУНКЦИИ ----

// SetDebugMode быстрое включение/выключение debug
func SetDebugMode(enabled bool) {
	if currentConfig.DebugMode != enabled {
		currentConfig.DebugMode = enabled
		if enabled {
			Debug("Debug mode enabled")
		}
	}
}

// IsDebugMode проверка режима debug
func IsDebugMode() bool {
	return currentConfig.DebugMode
}

// GetConfig возвращает текущую конфигурацию
func GetConfig() Config {
	return currentConfig
}
