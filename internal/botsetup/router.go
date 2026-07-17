package botsetup

import (
	"context"
	"cu-timepad-bot/internal/config"
	"cu-timepad-bot/internal/handler"
	"cu-timepad-bot/internal/middleware"

	"github.com/go-telegram/bot"
)

func Handle(ctx context.Context, h handler.Handler) (*bot.Bot, error) {
	cfg, _ := config.GetConfig(ctx)
	bot_opts := []bot.Option{
		bot.WithMiddlewares(middleware.Logging),
	}
	if cfg.TelegramAPIURL != "" {
		bot_opts = append(bot_opts, bot.WithServerURL(cfg.TelegramAPIURL))
	}
	b, err := bot.New(cfg.BotToken, bot_opts...)
	if err != nil {
		return nil, err
	}

	b.RegisterHandler(bot.HandlerTypeMessageText, "/start", bot.MatchTypeExact, h.Start)
	return b, nil
}
