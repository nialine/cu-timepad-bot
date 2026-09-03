package handler

import (
	"context"
	"log/slog"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

func (h *Handler) DefaultHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	if h.svc.InRegistrationProcess(ctx, update.Message.Chat.ID) {
		h.handleRegistration(ctx, b, update)
		return
	}
	slog.LogAttrs(ctx,
		slog.LevelWarn,
		"Invalid message",
		slog.Int64("userid", update.Message.Chat.ID),
	)
}
