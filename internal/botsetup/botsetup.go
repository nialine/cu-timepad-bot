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
	"golang.org/x/net/proxy"
)

const pollTimeout = 25 * time.Second

func Handle(ctx context.Context, h handler.Handler) (*bot.Bot, error) {
	cfg, _ := config.GetConfig(ctx)
	bot_opts := []bot.Option{
		bot.WithMiddlewares(
			middleware.SingleFlight,
			middleware.Logging,
			middleware.AutoRespond,
		),
	}
	if cfg.TelegramAPIURL != "" {
		bot_opts = append(bot_opts, bot.WithServerURL(cfg.TelegramAPIURL))
	}
	if cfg.PROXYURL != "" {
		proxyURL, err := url.Parse(cfg.PROXYURL)
		if err != nil {
			slog.LogAttrs(ctx,
				slog.LevelError,
				"Invalid proxy",
				slog.Any("error", err),
				slog.String("proxy_url", cfg.PROXYURL),
			)
			return nil, err
		}

		dialer, err := proxy.FromURL(proxyURL, proxy.Direct)
		if err != nil {
			slog.LogAttrs(ctx,
				slog.LevelError,
				"Can't create dialer",
				slog.Any("error", err),
				slog.String("proxy_url", cfg.PROXYURL),
			)
			return nil, err
		}

		contextDialer, ok := dialer.(proxy.ContextDialer)
		if !ok {
			slog.LogAttrs(ctx,
				slog.LevelError,
				"Dialer does not support contextDialer",
				slog.Any("error", err),
				slog.String("proxy_url", cfg.PROXYURL),
			)
			return nil, err
		}

		transport := &http.Transport{
			DialContext: contextDialer.DialContext,
		}

		client := &http.Client{
			Transport: transport,
		}

		bot_opts = append(bot_opts, bot.WithHTTPClient(pollTimeout, client))
	}

	b, err := bot.New(cfg.BotToken, bot_opts...)
	if err != nil {
		return nil, err
	}

	if cfg.WebhookURL != "" {
		slog.LogAttrs(ctx,
			slog.LevelWarn,
			"Using webhook",
			slog.String("webhook_url", cfg.WebhookURL),
		)
		b.SetWebhook(ctx, &bot.SetWebhookParams{
			URL: cfg.WebhookURL,
		})
		go func() {
			http.ListenAndServe(":80", b.WebhookHandler())
		}()
	} else {
		b.DeleteWebhook(ctx, nil)
	}

	b.RegisterHandler(bot.HandlerTypeMessageText, "/start", bot.MatchTypeExact, h.Start)
	return b, nil
}
