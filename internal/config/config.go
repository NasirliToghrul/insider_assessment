package config

import (
	"fmt"
	"os"
)

type Config struct {
	HTTPPort      string
	DBDriver      string
	DBDSN         string
	WebhookURL    string
	WebhookKey    string
	TickerSeconds int
	BatchSize     int
	MsgCharLimit  int
	RedisEnabled  bool
	RedisAddr     string
	RedisPassword string
	RedisDB       int
}

func getenv(key, def string) string {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	return v
}

func boolenv(key string, def bool) bool {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	if v == "1" || v == "true" || v == "TRUE" || v == "yes" {
		return true
	}
	return false
}

func intenv(key string, def int) int {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	var i int
	_, err := fmt.Sscanf(v, "%d", &i)
	if err != nil {
		return def
	}
	return i
}

func Load() Config {
	return Config{
		HTTPPort:      getenv("HTTP_PORT", "8080"),
		DBDriver:      getenv("DB_DRIVER", "postgres"),
		DBDSN:         getenv("DB_DSN", "host=postgres user=postgres password=postgres dbname=insider port=5432 sslmode=disable"),
		WebhookURL:    getenv("WEBHOOK_URL", ""),
		WebhookKey:    getenv("WEBHOOK_KEY", ""),
		TickerSeconds: intenv("TICKER_SECONDS", 120),
		BatchSize:     intenv("BATCH_SIZE", 2),
		MsgCharLimit:  intenv("MESSAGE_CHAR_LIMIT", 160),
		RedisEnabled:  boolenv("REDIS_ENABLED", true),
		RedisAddr:     getenv("REDIS_ADDR", "redis:6379"),
		RedisPassword: getenv("REDIS_PASSWORD", ""),
		RedisDB:       intenv("REDIS_DB", 0),
	}
}
