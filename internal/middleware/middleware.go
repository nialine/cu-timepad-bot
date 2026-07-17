package middleware

import (
	"context"
	"log/slog"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

func Logging(next bot.HandlerFunc) bot.HandlerFunc {
	return func(ctx context.Context, b *bot.Bot, update *models.Update) {
		slog.LogAttrs(ctx,
			slog.LevelDebug,
			"Got message",
			slog.Int64("chatid", update.Message.Chat.ID),
			slog.String("message", update.Message.Text),
		)
		next(ctx, b, update)

	}
}
