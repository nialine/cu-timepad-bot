package middleware

import (
	"context"
	"log/slog"
	"sync"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

func Logging(next bot.HandlerFunc) bot.HandlerFunc {
	return func(ctx context.Context, b *bot.Bot, update *models.Update) {
		if update.Message != nil {
			slog.LogAttrs(ctx,
				slog.LevelDebug,
				"Got message",
				slog.Int64("chatid", update.Message.Chat.ID),
				slog.String("message", update.Message.Text),
			)
		}
		next(ctx, b, update)
	}
}

func AutoRespond(next bot.HandlerFunc) bot.HandlerFunc {
	return func(ctx context.Context, b *bot.Bot, update *models.Update) {
		if update.Message != nil {
			b.SendChatAction(ctx, &bot.SendChatActionParams{
				ChatID: update.Message.Chat.ID,
				Action: models.ChatActionTyping,
			})
		}
		next(ctx, b, update)
	}
}

func SingleFlight(next bot.HandlerFunc) bot.HandlerFunc {
	sf := sync.Map{}

	return func(ctx context.Context, b *bot.Bot, update *models.Update) {
		if update.CallbackQuery != nil {
			key := update.CallbackQuery.ID
			if _, loaded := sf.LoadOrStore(key, nil); loaded {
				return
			}
			defer sf.Delete(key)
		}
		next(ctx, b, update)
	}
}
