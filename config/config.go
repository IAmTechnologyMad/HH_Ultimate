package config

import (
	"log"
	"os"
	"strconv"
	"time"
)

type Config struct {
	TelegramBotToken string
	TelegramChatID   int64
	PollInterval     time.Duration
	RequestTimeout   time.Duration
	MaxRetries       int
	RetryDelay       time.Duration
	Port             string
	DBPath           string
	FirstCryURL      string
	LogLevel         string
	PingURL          string
}

func Load() *Config {
	cfg := &Config{}

	cfg.TelegramBotToken = mustGetEnv("TELEGRAM_BOT_TOKEN")

	chatIDStr := mustGetEnv("TELEGRAM_CHAT_ID")
	chatID, err := strconv.ParseInt(chatIDStr, 10, 64)
	if err != nil {
		log.Fatalf("Invalid TELEGRAM_CHAT_ID: %v", err)
	}
	cfg.TelegramChatID = chatID

	cfg.PollInterval = time.Duration(getEnvInt("POLL_INTERVAL_SECONDS", 30)) * time.Second
	cfg.RequestTimeout = time.Duration(getEnvInt("REQUEST_TIMEOUT_SECONDS", 10)) * time.Second
	cfg.MaxRetries = getEnvInt("MAX_RETRIES", 3)
	cfg.RetryDelay = time.Duration(getEnvInt("RETRY_DELAY_MS", 500)) * time.Millisecond
	cfg.Port = getEnvStr("PORT", "8080")
	cfg.DBPath = getEnvStr("DB_PATH", "/var/data/products.db")
	cfg.LogLevel = getEnvStr("LOG_LEVEL", "info")
	cfg.FirstCryURL = getEnvStr("FIRSTCRY_URL",
		"https://www.firstcry.com/hotwheels/5/0/113?sort=popularity&q=ard-hot%20wheels&ref2=q_ard_hot%20wheels&asid=53241")
	cfg.PingURL = getEnvStr("RENDER_EXTERNAL_URL", "")

	return cfg
}

func mustGetEnv(key string) string {
	val := os.Getenv(key)
	if val == "" {
		log.Fatalf("Required environment variable %s is not set", key)
	}
	return val
}

func getEnvStr(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}

func getEnvInt(key string, defaultVal int) int {
	if val := os.Getenv(key); val != "" {
		if i, err := strconv.Atoi(val); err == nil {
			return i
		}
	}
	return defaultVal
}
