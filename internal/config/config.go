package config

import (
	"context"
	"cu-timepad-bot/internal/domain"
	"os"

	"github.com/caarlos0/env/v11"
	"go.yaml.in/yaml/v4"
)

type Config struct {
	TelegramAPIURL string `env:"TELEGRAM_API_URL"`
	BotToken       string `env:"BOT_TOKEN,required"`
	WebhookURL     string `env:"WEBHOOK_URL"`

	LogLevel string `env:"LOG_LEVEL" envDefault:"warn"`

	MongoURI         string `env:"MONGODB"`
	DBName           string `env:"MONGO_DBNAME" envDefault:"timepad-bot"`
	MemCacheDuration int    `env:"MEMCACHE_DURATION envDefault:300"`

	PROXYURL string `env:"PROXY_URL"`

	Events               []domain.Event `yaml:"events"`
	TimepadFetchInterval int            `yaml:"fetch_interval"`
	TimepadTimeout       int            `yaml:"timepad_timeout"`
}

func Load() (Config, error) {
	var cfg Config

	cfg.TimepadFetchInterval = 10 * 60
	cfg.TimepadTimeout = 30

	err := env.Parse(&cfg)
	if err != nil {
		return cfg, nil
	}

	f, err := os.OpenFile("config.yaml", os.O_RDONLY, 0)
	if err != nil {
		return cfg, err
	}

	err = yaml.NewDecoder(f).Decode(&cfg)
	if err != nil {
		return cfg, err
	}

	return cfg, nil
}

type ctxKey int

const configKey ctxKey = iota

func InjectConfig(ctx context.Context, cfg *Config) context.Context {
	return context.WithValue(ctx, configKey, cfg)
}

func GetConfig(ctx context.Context) *Config {
	cfg, ok := ctx.Value(configKey).(*Config)
	if !ok {
		panic("No config")
	}
	return cfg
}
