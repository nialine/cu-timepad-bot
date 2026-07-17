package config

import (
	"context"

	"github.com/caarlos0/env/v11"
)

type Config struct {
	EnvConfig
}

type EnvConfig struct {
	TelegramAPIURL string `env:"TELEGRAM_API_URL"`
	BotToken       string `env:"BOT_TOKEN,required"`
	LogLevel       string `env:"LOG_LEVEL" envDefault:"warn"`
	MongoURI       string `env:"MONGODB"`
	DBName         string `env:"MONGO_DBNAME" envDefault:"timepad-bot"`

	PROXYURL string `env:"PROXY_URL"`
}

type YamlConfig struct {
	Events []Event `yaml:"events"`
}

type Event struct {
	Name string `yaml:"name"`
	URL  string `yaml:"url"`
}

func Load() (Config, error) {
	var cfg Config

	err := env.Parse(&cfg)

	return cfg, err
}

type ctxKey int

const configKey ctxKey = iota

func InjectConfig(ctx context.Context, cfg Config) context.Context {
	return context.WithValue(ctx, configKey, cfg)
}

func GetConfig(ctx context.Context) (Config, bool) {
	cfg, ok := ctx.Value(configKey).(Config)
	return cfg, ok
}
