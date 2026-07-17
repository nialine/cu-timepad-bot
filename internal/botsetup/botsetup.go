package botsetup

import (
	"context"
	"cu-timepad-bot/internal/config"
	"cu-timepad-bot/internal/handler"
	"cu-timepad-bot/internal/middleware"
	"log/slog"
	"net/http"
	"net/url"
	"time"

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
	if cfg.PROXYURL != "" {
		proxyURL, err := url.Parse(cfg.PROXYURL)
		if err != nil {
			slog.LogAttrs(ctx,
				slog.LevelWarn,
				"Ignoring invalid setting proxy",
				slog.Any("error", err),
				slog.String("proxy_url", cfg.PROXYURL),
			)
		}
		transport := &http.Transport{
			Proxy: http.ProxyURL(proxyURL),
		}

		client := &http.Client{
			Transport: transport,
		}

		bot_opts = append(bot_opts, bot.WithHTTPClient(10*time.Second, client))
	}
	b, err := bot.New(cfg.BotToken, bot_opts...)
	if err != nil {
		return nil, err
	}

	b.RegisterHandler(bot.HandlerTypeMessageText, "/start", bot.MatchTypeExact, h.Start)
	return b, nil
}
